package models

import "time"

type Alert struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	DeviceID  string    `json:"device_id,omitempty"`
	ScanID    string    `json:"scan_id,omitempty"`
	Read      bool      `json:"read"`
}
