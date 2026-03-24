package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/netmap/netmap/internal/core/models"
	"github.com/netmap/netmap/internal/store"
)

type MonitorRepo struct {
	db *DB
}

func NewMonitorRepo(db *DB) *MonitorRepo {
	return &MonitorRepo{db: db}
}

func (r *MonitorRepo) Create(ctx context.Context, m *models.Monitor) error {
	active := 0
	if m.Active {
		active = 1
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO monitors (id, name, type, url, host, port, interval_secs, timeout_secs, method, expected_status, keyword, active, notify_webhook, status, last_checked_at, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.ID, m.Name, string(m.Type), m.URL, m.Host, m.Port,
		m.Interval, m.Timeout, m.Method, m.ExpectedStatus,
		m.Keyword, active, m.NotifyWebhook, string(m.Status),
		m.LastCheckedAt, m.CreatedAt,
	)
	return err
}

func (r *MonitorRepo) List(ctx context.Context) ([]models.Monitor, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, type, url, host, port, interval_secs, timeout_secs, method, expected_status, keyword, active, notify_webhook, status, last_checked_at, created_at
		 FROM monitors ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMonitors(rows)
}

func (r *MonitorRepo) GetByID(ctx context.Context, id string) (*models.Monitor, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, type, url, host, port, interval_secs, timeout_secs, method, expected_status, keyword, active, notify_webhook, status, last_checked_at, created_at
		 FROM monitors WHERE id = ?`, id)
	m, err := scanMonitor(row)
	if err == sql.ErrNoRows {
		return nil, store.ErrNotFound
	}
	return m, err
}

func (r *MonitorRepo) Update(ctx context.Context, m *models.Monitor) error {
	active := 0
	if m.Active {
		active = 1
	}
	_, err := r.db.ExecContext(ctx,
		`UPDATE monitors SET name=?, type=?, url=?, host=?, port=?, interval_secs=?, timeout_secs=?, method=?, expected_status=?, keyword=?, active=?, notify_webhook=?, status=?, last_checked_at=?
		 WHERE id=?`,
		m.Name, string(m.Type), m.URL, m.Host, m.Port,
		m.Interval, m.Timeout, m.Method, m.ExpectedStatus,
		m.Keyword, active, m.NotifyWebhook, string(m.Status),
		m.LastCheckedAt, m.ID,
	)
	return err
}

func (r *MonitorRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM monitors WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return store.ErrNotFound
	}
	return nil
}

func (r *MonitorRepo) UpdateStatus(ctx context.Context, id string, status models.MonitorStatus, lastCheckedAt time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE monitors SET status=?, last_checked_at=? WHERE id=?`,
		string(status), lastCheckedAt, id,
	)
	return err
}

func (r *MonitorRepo) ListActive(ctx context.Context) ([]models.Monitor, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, type, url, host, port, interval_secs, timeout_secs, method, expected_status, keyword, active, notify_webhook, status, last_checked_at, created_at
		 FROM monitors WHERE active = 1 ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMonitors(rows)
}

func (r *MonitorRepo) CreateCheck(ctx context.Context, c *models.MonitorCheck) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO monitor_checks (id, monitor_id, status, response_time_ms, status_code, error, checked_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		c.ID, c.MonitorID, string(c.Status), c.ResponseTimeMs, c.StatusCode, c.Error, c.CheckedAt,
	)
	return err
}

func (r *MonitorRepo) ListChecks(ctx context.Context, monitorID string, limit int) ([]models.MonitorCheck, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, monitor_id, status, response_time_ms, status_code, error, checked_at
		 FROM monitor_checks WHERE monitor_id = ? ORDER BY checked_at DESC LIMIT ?`,
		monitorID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var checks []models.MonitorCheck
	for rows.Next() {
		var c models.MonitorCheck
		if err := rows.Scan(&c.ID, &c.MonitorID, &c.Status, &c.ResponseTimeMs, &c.StatusCode, &c.Error, &c.CheckedAt); err != nil {
			return nil, err
		}
		checks = append(checks, c)
	}
	return checks, rows.Err()
}

func (r *MonitorRepo) DeleteOldChecks(ctx context.Context, monitorID string, keepCount int) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM monitor_checks WHERE monitor_id = ? AND id NOT IN (
			SELECT id FROM monitor_checks WHERE monitor_id = ? ORDER BY checked_at DESC LIMIT ?
		)`,
		monitorID, monitorID, keepCount,
	)
	return err
}

func (r *MonitorRepo) UptimePercent(ctx context.Context, monitorID string, since time.Time) (float64, error) {
	var total, up int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM monitor_checks WHERE monitor_id = ? AND checked_at > ?`,
		monitorID, since,
	).Scan(&total)
	if err != nil {
		return 0, err
	}
	if total == 0 {
		return 0.0, nil
	}
	err = r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM monitor_checks WHERE monitor_id = ? AND checked_at > ? AND status = 'up'`,
		monitorID, since,
	).Scan(&up)
	if err != nil {
		return 0, err
	}
	return float64(up) / float64(total) * 100.0, nil
}

// scanMonitor scans a single monitor row from a QueryRow result.
func scanMonitor(row *sql.Row) (*models.Monitor, error) {
	var m models.Monitor
	var active int
	var lastCheckedAt sql.NullTime
	err := row.Scan(
		&m.ID, &m.Name, &m.Type, &m.URL, &m.Host, &m.Port,
		&m.Interval, &m.Timeout, &m.Method, &m.ExpectedStatus,
		&m.Keyword, &active, &m.NotifyWebhook, &m.Status,
		&lastCheckedAt, &m.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	m.Active = active != 0
	if lastCheckedAt.Valid {
		t := lastCheckedAt.Time
		m.LastCheckedAt = &t
	}
	return &m, nil
}

// scanMonitors scans multiple monitor rows from a Rows result.
func scanMonitors(rows *sql.Rows) ([]models.Monitor, error) {
	var monitors []models.Monitor
	for rows.Next() {
		var m models.Monitor
		var active int
		var lastCheckedAt sql.NullTime
		err := rows.Scan(
			&m.ID, &m.Name, &m.Type, &m.URL, &m.Host, &m.Port,
			&m.Interval, &m.Timeout, &m.Method, &m.ExpectedStatus,
			&m.Keyword, &active, &m.NotifyWebhook, &m.Status,
			&lastCheckedAt, &m.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		m.Active = active != 0
		if lastCheckedAt.Valid {
			t := lastCheckedAt.Time
			m.LastCheckedAt = &t
		}
		monitors = append(monitors, m)
	}
	return monitors, rows.Err()
}

// Ensure MonitorRepo satisfies store.MonitorRepo at compile time.
var _ store.MonitorRepo = (*MonitorRepo)(nil)
