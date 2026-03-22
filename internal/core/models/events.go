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
