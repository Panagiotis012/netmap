// internal/core/config/config.go
package config

import (
	"flag"
	"io"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	Port         int
	DataDir      string
	DBDriver     string
	DBURL        string
	ScanInterval time.Duration
	ScanWorkers  int
}

func Default() *Config {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return &Config{
		Port:         8080,
		DataDir:      filepath.Join(home, ".netmap"),
		DBDriver:     "sqlite",
		DBURL:        "",
		ScanInterval: 5 * time.Minute,
		ScanWorkers:  50,
	}
}

func (c *Config) ParseFlags(args []string) error {
	fs := flag.NewFlagSet("netmap", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.IntVar(&c.Port, "port", c.Port, "HTTP server port")
	fs.StringVar(&c.DataDir, "data-dir", c.DataDir, "Data directory")
	fs.StringVar(&c.DBURL, "db-url", c.DBURL, "PostgreSQL connection URL (overrides SQLite)")
	fs.DurationVar(&c.ScanInterval, "scan-interval", c.ScanInterval, "Scan interval")
	fs.IntVar(&c.ScanWorkers, "scan-workers", c.ScanWorkers, "Concurrent scan workers")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if c.DBURL != "" {
		c.DBDriver = "postgres"
	}
	return nil
}

func (c *Config) DBPath() string {
	return filepath.Join(c.DataDir, "netmap.db")
}
