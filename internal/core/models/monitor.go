package models

import "time"

type MonitorType string

const (
	MonitorHTTP MonitorType = "http"
	MonitorTCP  MonitorType = "tcp"
	MonitorPing MonitorType = "ping"
)

type MonitorStatus string

const (
	MonitorStatusUp      MonitorStatus = "up"
	MonitorStatusDown    MonitorStatus = "down"
	MonitorStatusPending MonitorStatus = "pending"
)

type Monitor struct {
	ID             string        `json:"id"`
	Name           string        `json:"name"`
	Type           MonitorType   `json:"type"`
	URL            string        `json:"url"`             // for http
	Host           string        `json:"host"`            // for tcp/ping
	Port           int           `json:"port"`            // for tcp
	Interval       int           `json:"interval"`        // seconds, default 60
	Timeout        int           `json:"timeout"`         // seconds, default 10
	Method         string        `json:"method"`          // GET, POST, HEAD
	ExpectedStatus int           `json:"expected_status"` // 200
	Keyword        string        `json:"keyword"`         // optional body check
	Active         bool          `json:"active"`
	NotifyWebhook  string        `json:"notify_webhook"` // Discord/Slack URL
	Status         MonitorStatus `json:"status"`         // current status
	LastCheckedAt  *time.Time    `json:"last_checked_at,omitempty"`
	UptimeDay      float64       `json:"uptime_day"`  // % last 24h
	UptimeWeek     float64       `json:"uptime_week"` // % last 7d
	CreatedAt      time.Time     `json:"created_at"`
}

type MonitorCheck struct {
	ID             string        `json:"id"`
	MonitorID      string        `json:"monitor_id"`
	Status         MonitorStatus `json:"status"`
	ResponseTimeMs int           `json:"response_time_ms"`
	StatusCode     int           `json:"status_code,omitempty"`
	Error          string        `json:"error,omitempty"`
	CheckedAt      time.Time     `json:"checked_at"`
}
