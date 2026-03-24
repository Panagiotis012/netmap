package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
}

func Open(dbPath string) (*DB, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	sdb := &DB{db}
	if err := sdb.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return sdb, nil
}

func (db *DB) migrate() error {
	// Ensure the migrations tracking table exists.
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version     INTEGER PRIMARY KEY,
		applied_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	// Find the highest version already applied.
	var maxVersion int
	_ = db.QueryRow(`SELECT COALESCE(MAX(version), 0) FROM schema_migrations`).Scan(&maxVersion)

	for i, m := range migrations {
		version := i + 1
		if version <= maxVersion {
			continue // already applied
		}
		if _, err := db.Exec(m.sql); err != nil {
			// ALTER TABLE fails with "duplicate column name" on fresh installs
			// where the column already exists from the CREATE TABLE migration.
			if strings.Contains(err.Error(), "duplicate column name") {
				// Column already exists — safe to skip.
			} else {
				return fmt.Errorf("migration v%d (%s) failed: %w", version, m.name, err)
			}
		}
		if _, err := db.Exec(`INSERT INTO schema_migrations (version) VALUES (?)`, version); err != nil {
			return fmt.Errorf("record migration v%d: %w", version, err)
		}
	}
	return nil
}
