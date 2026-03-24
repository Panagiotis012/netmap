package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/netmap/netmap/internal/core/models"
	"github.com/netmap/netmap/internal/store"
)

type AlertRepo struct {
	db *DB
}

func NewAlertRepo(db *DB) *AlertRepo {
	return &AlertRepo{db: db}
}

func (r *AlertRepo) Create(ctx context.Context, a *models.Alert) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO alerts (id, type, message, timestamp, device_id, scan_id, read) VALUES (?, ?, ?, ?, ?, ?, 0)`,
		a.ID, a.Type, a.Message, a.Timestamp, nilStr(a.DeviceID), nilStr(a.ScanID),
	)
	return err
}

func (r *AlertRepo) List(ctx context.Context, limit int) ([]models.Alert, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, type, message, timestamp, device_id, scan_id, read FROM alerts ORDER BY timestamp DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []models.Alert
	for rows.Next() {
		var a models.Alert
		var deviceID, scanID sql.NullString
		if err := rows.Scan(&a.ID, &a.Type, &a.Message, &a.Timestamp, &deviceID, &scanID, &a.Read); err != nil {
			return nil, err
		}
		a.DeviceID = deviceID.String
		a.ScanID = scanID.String
		alerts = append(alerts, a)
	}
	return alerts, rows.Err()
}

func (r *AlertRepo) MarkAllRead(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `UPDATE alerts SET read = 1 WHERE read = 0`)
	return err
}

func (r *AlertRepo) DeleteAll(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM alerts`)
	return err
}

func (r *AlertRepo) UnreadCount(ctx context.Context) (int, error) {
	var n int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM alerts WHERE read = 0`).Scan(&n)
	return n, err
}

// Trim keeps only the most recent `keep` alerts.
func (r *AlertRepo) Trim(ctx context.Context, keep int) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM alerts WHERE id NOT IN (SELECT id FROM alerts ORDER BY timestamp DESC LIMIT ?)`, keep)
	return err
}

func nilStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// Ensure AlertRepo satisfies store.AlertRepo at compile time.
var _ store.AlertRepo = (*AlertRepo)(nil)

// SessionRepo handles auth session persistence.
type SessionRepo struct {
	db *DB
}

func NewSessionRepo(db *DB) *SessionRepo {
	return &SessionRepo{db: db}
}

func (r *SessionRepo) Create(ctx context.Context, token string, ttl time.Duration) error {
	expires := time.Now().Add(ttl)
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO sessions (token, expires_at) VALUES (?, ?)`, token, expires)
	return err
}

func (r *SessionRepo) Validate(ctx context.Context, token string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM sessions WHERE token = ? AND expires_at > CURRENT_TIMESTAMP`, token).Scan(&count)
	return count > 0, err
}

func (r *SessionRepo) Delete(ctx context.Context, token string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE token = ?`, token)
	return err
}

func (r *SessionRepo) DeleteExpired(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE expires_at <= CURRENT_TIMESTAMP`)
	return err
}
