// internal/core/models/device.go
package models

import (
	"encoding/json"
	"time"
)

type DeviceStatus string

const (
	StatusOnline  DeviceStatus = "online"
	StatusOffline DeviceStatus = "offline"
	StatusUnknown DeviceStatus = "unknown"
)

type DiscoveryMethod string

const (
	DiscoveryScan   DiscoveryMethod = "scan"
	DiscoveryAgent  DiscoveryMethod = "agent"
	DiscoveryManual DiscoveryMethod = "manual"
)

type Device struct {
	ID              string          `json:"id" db:"id"`
	Hostname        string          `json:"hostname" db:"hostname"`
	IPAddresses     []string        `json:"ip_addresses" db:"ip_addresses"`
	MACAddresses    []string        `json:"mac_addresses" db:"mac_addresses"`
	OS              string          `json:"os" db:"os"`
	Status          DeviceStatus    `json:"status" db:"status"`
	DiscoveryMethod DiscoveryMethod `json:"discovery_method" db:"discovery_method"`
	FirstSeenAt     time.Time       `json:"first_seen_at" db:"first_seen_at"`
	LastSeenAt      time.Time       `json:"last_seen_at" db:"last_seen_at"`
	Tags            []string        `json:"tags" db:"tags"`
	GroupID         *string         `json:"group_id,omitempty" db:"group_id"`
	Metadata        json.RawMessage `json:"metadata,omitempty" db:"metadata"`
	MapX            *float64        `json:"map_x,omitempty" db:"map_x"`
	MapY            *float64        `json:"map_y,omitempty" db:"map_y"`
	NetworkID       *string         `json:"network_id,omitempty" db:"network_id"`
}

type Network struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Subnet    string    `json:"subnet" db:"subnet"`
	Gateway   string    `json:"gateway" db:"gateway"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type ScanType string

const (
	ScanDiscovery ScanType = "discovery"
	ScanPort      ScanType = "port"
	ScanFull      ScanType = "full"
)

type ScanStatus string

const (
	ScanPending   ScanStatus = "pending"
	ScanRunning   ScanStatus = "running"
	ScanCompleted ScanStatus = "completed"
	ScanFailed    ScanStatus = "failed"
	ScanCancelled ScanStatus = "cancelled"
)

type ScanJob struct {
	ID          string          `json:"id" db:"id"`
	Type        ScanType        `json:"type" db:"type"`
	Target      string          `json:"target" db:"target"`
	Status      ScanStatus      `json:"status" db:"status"`
	StartedAt   *time.Time      `json:"started_at,omitempty" db:"started_at"`
	CompletedAt *time.Time      `json:"completed_at,omitempty" db:"completed_at"`
	Results     json.RawMessage `json:"results,omitempty" db:"results"`
}

type ScanResults struct {
	Hosts []HostResult `json:"hosts"`
	Stats ScanStats    `json:"stats"`
}

type HostStatus string

const (
	HostUp   HostStatus = "up"
	HostDown HostStatus = "down"
)

type HostResult struct {
	IP        string       `json:"ip"`
	MAC       string       `json:"mac"`
	Hostname  string       `json:"hostname"`
	LatencyMs float64      `json:"latency_ms"`
	Ports     []PortResult `json:"ports,omitempty"`
	OSGuess   string       `json:"os_guess,omitempty"`
	Status    HostStatus   `json:"status"`
}

type PortResult struct {
	Number   int    `json:"number"`
	Protocol string `json:"protocol"`
	Service  string `json:"service"`
	State    string `json:"state"`
}

type ScanStats struct {
	HostsScanned int   `json:"hosts_scanned"`
	HostsUp      int   `json:"hosts_up"`
	DurationMs   int64 `json:"duration_ms"`
}

// Pagination
type ListParams struct {
	Page   int    `json:"page"`
	Limit  int    `json:"limit"`
	Sort   string `json:"sort"`
	Order  string `json:"order"`
	Search string `json:"search,omitempty"`
	Status DeviceStatus `json:"status,omitempty"`
	Tag    string `json:"tag,omitempty"`
}

type ListResult[T any] struct {
	Items      []T `json:"items"`
	Total      int `json:"total"`
	Page       int `json:"page"`
	TotalPages int `json:"total_pages"`
}
