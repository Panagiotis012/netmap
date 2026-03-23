package sqlite

import (
	"context"
	"database/sql"
)

type ConfigRepo struct {
	db *sql.DB
}

func NewConfigRepo(db *DB) *ConfigRepo {
	return &ConfigRepo{db: db.DB}
}

func (r *ConfigRepo) Get(ctx context.Context, key string) string {
	var val string
	err := r.db.QueryRowContext(ctx, `SELECT value FROM config WHERE key = ?`, key).Scan(&val)
	if err != nil {
		return ""
	}
	return val
}

func (r *ConfigRepo) Set(ctx context.Context, key, value string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO config (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value`,
		key, value)
	return err
}

func (r *ConfigRepo) GetAll(ctx context.Context) map[string]string {
	rows, err := r.db.QueryContext(ctx, `SELECT key, value FROM config`)
	if err != nil {
		return map[string]string{}
	}
	defer rows.Close()
	result := map[string]string{}
	for rows.Next() {
		var k, v string
		if rows.Scan(&k, &v) == nil {
			result[k] = v
		}
	}
	return result
}
