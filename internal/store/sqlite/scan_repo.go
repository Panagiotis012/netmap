// internal/store/sqlite/scan_repo.go
package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/netmap/netmap/internal/core/models"
	"github.com/netmap/netmap/internal/store"
)

type ScanRepo struct {
	db *DB
}

func NewScanRepo(db *DB) *ScanRepo {
	return &ScanRepo{db: db}
}

var _ store.ScanRepo = (*ScanRepo)(nil)

func (r *ScanRepo) Create(ctx context.Context, s *models.ScanJob) error {
	results := s.Results
	if results == nil {
		results = json.RawMessage("{}")
	}
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO scan_jobs (id, type, target, status, started_at, completed_at, results) VALUES (?, ?, ?, ?, ?, ?, ?)",
		s.ID, s.Type, s.Target, s.Status, s.StartedAt, s.CompletedAt, string(results))
	return err
}

func (r *ScanRepo) GetByID(ctx context.Context, id string) (*models.ScanJob, error) {
	var s models.ScanJob
	var results string
	var startedAt, completedAt sql.NullTime
	err := r.db.QueryRowContext(ctx,
		"SELECT id, type, target, status, started_at, completed_at, results FROM scan_jobs WHERE id = ?", id).
		Scan(&s.ID, &s.Type, &s.Target, &s.Status, &startedAt, &completedAt, &results)
	if err == sql.ErrNoRows {
		return nil, store.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if startedAt.Valid {
		s.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		s.CompletedAt = &completedAt.Time
	}
	s.Results = json.RawMessage(results)
	return &s, nil
}

func (r *ScanRepo) List(ctx context.Context, params models.ListParams) (*models.ListResult[models.ScanJob], error) {
	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM scan_jobs").Scan(&total); err != nil {
		return nil, err
	}

	limit := 50
	if params.Limit > 0 {
		limit = params.Limit
	}
	page := 1
	if params.Page > 0 {
		page = params.Page
	}
	offset := (page - 1) * limit

	rows, err := r.db.QueryContext(ctx,
		"SELECT id, type, target, status, started_at, completed_at, results FROM scan_jobs ORDER BY started_at DESC LIMIT ? OFFSET ?",
		limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scans []models.ScanJob
	for rows.Next() {
		var s models.ScanJob
		var results string
		var startedAt, completedAt sql.NullTime
		if err := rows.Scan(&s.ID, &s.Type, &s.Target, &s.Status, &startedAt, &completedAt, &results); err != nil {
			return nil, err
		}
		if startedAt.Valid {
			s.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			s.CompletedAt = &completedAt.Time
		}
		s.Results = json.RawMessage(results)
		scans = append(scans, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	totalPages := (total + limit - 1) / limit
	return &models.ListResult[models.ScanJob]{
		Items: scans, Total: total, Page: page, TotalPages: totalPages,
	}, nil
}

func (r *ScanRepo) Update(ctx context.Context, s *models.ScanJob) error {
	results := s.Results
	if results == nil {
		results = json.RawMessage("{}")
	}
	res, err := r.db.ExecContext(ctx,
		"UPDATE scan_jobs SET status=?, started_at=?, completed_at=?, results=? WHERE id=?",
		s.Status, s.StartedAt, s.CompletedAt, string(results), s.ID)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return store.ErrNotFound
	}
	return nil
}

func (r *ScanRepo) DeleteOlderThan(ctx context.Context, keepCount int) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM scan_jobs WHERE id NOT IN (SELECT id FROM scan_jobs ORDER BY started_at DESC LIMIT ?)`,
		keepCount)
	return err
}
