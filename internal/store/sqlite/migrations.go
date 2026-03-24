package sqlite

type migration struct {
	name string
	sql  string
}

var migrations = []migration{
	{
		name: "create_networks",
		sql: `CREATE TABLE IF NOT EXISTS networks (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			subnet TEXT NOT NULL UNIQUE,
			gateway TEXT NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
	},
	{
		name: "create_devices",
		sql: `CREATE TABLE IF NOT EXISTS devices (
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
			ports TEXT NOT NULL DEFAULT '[]',
			latency_ms REAL NOT NULL DEFAULT 0,
			group_id TEXT,
			metadata TEXT DEFAULT '{}',
			map_x REAL,
			map_y REAL,
			network_id TEXT,
			FOREIGN KEY (network_id) REFERENCES networks(id) ON DELETE SET NULL
		)`,
	},
	{
		name: "create_scan_jobs",
		sql: `CREATE TABLE IF NOT EXISTS scan_jobs (
			id TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			target TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			started_at DATETIME,
			completed_at DATETIME,
			results TEXT DEFAULT '{}'
		)`,
	},
	{
		name: "create_indexes",
		sql: `CREATE INDEX IF NOT EXISTS idx_devices_status ON devices(status)`,
	},
	{
		name: "create_network_index",
		sql: `CREATE INDEX IF NOT EXISTS idx_devices_network ON devices(network_id)`,
	},
	{
		name: "create_scan_index",
		sql: `CREATE INDEX IF NOT EXISTS idx_scan_jobs_status ON scan_jobs(status)`,
	},
	{
		name: "create_config",
		sql: `CREATE TABLE IF NOT EXISTS config (
			key   TEXT PRIMARY KEY,
			value TEXT NOT NULL
		)`,
	},
	{
		name: "create_sessions",
		sql: `CREATE TABLE IF NOT EXISTS sessions (
			token      TEXT PRIMARY KEY,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			expires_at DATETIME NOT NULL
		)`,
	},
	{
		name: "create_alerts",
		sql: `CREATE TABLE IF NOT EXISTS alerts (
			id         TEXT PRIMARY KEY,
			type       TEXT NOT NULL,
			message    TEXT NOT NULL,
			timestamp  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			device_id  TEXT,
			scan_id    TEXT,
			read       INTEGER NOT NULL DEFAULT 0
		)`,
	},
	{
		name: "create_alert_index",
		sql: `CREATE INDEX IF NOT EXISTS idx_alerts_timestamp ON alerts(timestamp DESC)`,
	},
	// Additive column migrations for databases created before these columns existed.
	// ALTER TABLE fails with "duplicate column name" on fresh installs — that error
	// is silently ignored in the migration runner (see sqlite.go).
	{
		name: "add_devices_ports",
		sql:  `ALTER TABLE devices ADD COLUMN ports TEXT NOT NULL DEFAULT '[]'`,
	},
	{
		name: "add_devices_latency_ms",
		sql:  `ALTER TABLE devices ADD COLUMN latency_ms REAL NOT NULL DEFAULT 0`,
	},
}
