package models

import "time"

type EventType string

const (
	EventDeviceDiscovered EventType = "device.discovered"
	EventDeviceUpdated    EventType = "device.updated"
	EventDeviceLost       EventType = "device.lost"
	EventScanStarted      EventType = "scan.started"
	EventScanProgress     EventType = "scan.progress"
	EventScanCompleted    EventType = "scan.completed"
)

type Event struct {
	Type      EventType   `json:"type"`
	Payload   interface{} `json:"payload"`
	Timestamp time.Time   `json:"timestamp"`
}

type ScanProgressPayload struct {
	ScanID       string `json:"scan_id"`
	HostsScanned int    `json:"hosts_scanned"`
	HostsTotal   int    `json:"hosts_total"`
	HostsFound   int    `json:"hosts_found"`
	Percent      int    `json:"percent"`
	EtaSeconds   int    `json:"eta_seconds"`
}
