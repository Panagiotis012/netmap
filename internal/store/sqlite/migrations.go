package sqlite

var migrations = []string{
	`CREATE TABLE IF NOT EXISTS networks (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		subnet TEXT NOT NULL UNIQUE,
		gateway TEXT NOT NULL DEFAULT '',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE TABLE IF NOT EXISTS devices (
		id TEXT PRIMARY KEY,
		hostname TEXT NOT NULL DEFAULT '',
		ip_addresses TEXT NOT NULL DEFAULT '[]',
		mac_addresses TEXT NOT NULL DEFAULT '[]',
		os TEXT NOT NULL DEFAULT '',
		status TEXT NOT NULL DEFAULT 'unknown',
		discovery_method TEXT NOT NULL DEFAULT 'scan',
		first_seen_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		last_seen_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		tags TEXT NOT NULL DEFAULT '[]',
		group_id TEXT,
		metadata TEXT DEFAULT '{}',
		map_x REAL,
		map_y REAL,
		network_id TEXT,
		FOREIGN KEY (network_id) REFERENCES networks(id) ON DELETE SET NULL
	)`,
	`CREATE TABLE IF NOT EXISTS scan_jobs (
		id TEXT PRIMARY KEY,
		type TEXT NOT NULL,
		target TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'pending',
		started_at DATETIME,
		completed_at DATETIME,
		results TEXT DEFAULT '{}'
	)`,
	`CREATE INDEX IF NOT EXISTS idx_devices_status ON devices(status)`,
	`CREATE INDEX IF NOT EXISTS idx_devices_network ON devices(network_id)`,
	`CREATE INDEX IF NOT EXISTS idx_scan_jobs_status ON scan_jobs(status)`,
	`CREATE TABLE IF NOT EXISTS config (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
)`,
	`ALTER TABLE devices ADD COLUMN ports TEXT NOT NULL DEFAULT '[]'`,
	`ALTER TABLE devices ADD COLUMN latency_ms REAL NOT NULL DEFAULT 0`,
}
