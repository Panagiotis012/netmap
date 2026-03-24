// internal/store/repository.go
package store

import (
	"context"
	"errors"
	"time"

	"github.com/netmap/netmap/internal/core/models"
)

// ErrNotFound is returned when a requested resource does not exist.
var ErrNotFound = errors.New("not found")

type DeviceRepo interface {
	List(ctx context.Context, params models.ListParams) (*models.ListResult[models.Device], error)
	GetByID(ctx context.Context, id string) (*models.Device, error)
	GetByMAC(ctx context.Context, mac string) (*models.Device, error)
	GetByHostname(ctx context.Context, hostname string) (*models.Device, error)
	GetByIP(ctx context.Context, ip string) (*models.Device, error)
	Create(ctx context.Context, device *models.Device) error
	Update(ctx context.Context, device *models.Device) error
	Delete(ctx context.Context, id string) error
	UpdateStatus(ctx context.Context, id string, status models.DeviceStatus) error
	UpdatePosition(ctx context.Context, id string, x, y float64) error
	CountByStatus(ctx context.Context) (online, offline, unknown int, err error)
}

type NetworkRepo interface {
	List(ctx context.Context) ([]models.Network, error)
	GetByID(ctx context.Context, id string) (*models.Network, error)
	GetBySubnet(ctx context.Context, subnet string) (*models.Network, error)
	Create(ctx context.Context, network *models.Network) error
	Update(ctx context.Context, network *models.Network) error
	Delete(ctx context.Context, id string) error
}

type ScanRepo interface {
	List(ctx context.Context, params models.ListParams) (*models.ListResult[models.ScanJob], error)
	GetByID(ctx context.Context, id string) (*models.ScanJob, error)
	Create(ctx context.Context, scan *models.ScanJob) error
	Update(ctx context.Context, scan *models.ScanJob) error
	DeleteOlderThan(ctx context.Context, keepCount int) error
}

type AlertRepo interface {
	Create(ctx context.Context, alert *models.Alert) error
	List(ctx context.Context, limit int) ([]models.Alert, error)
	MarkAllRead(ctx context.Context) error
	DeleteAll(ctx context.Context) error
	UnreadCount(ctx context.Context) (int, error)
	Trim(ctx context.Context, keep int) error
}

type SessionRepo interface {
	Create(ctx context.Context, token string, ttl time.Duration) error
	Validate(ctx context.Context, token string) (bool, error)
	Delete(ctx context.Context, token string) error
	DeleteExpired(ctx context.Context) error
}

type MonitorRepo interface {
	Create(ctx context.Context, m *models.Monitor) error
	List(ctx context.Context) ([]models.Monitor, error)
	GetByID(ctx context.Context, id string) (*models.Monitor, error)
	Update(ctx context.Context, m *models.Monitor) error
	Delete(ctx context.Context, id string) error
	UpdateStatus(ctx context.Context, id string, status models.MonitorStatus, lastCheckedAt time.Time) error
	ListActive(ctx context.Context) ([]models.Monitor, error)

	CreateCheck(ctx context.Context, c *models.MonitorCheck) error
	ListChecks(ctx context.Context, monitorID string, limit int) ([]models.MonitorCheck, error)
	DeleteOldChecks(ctx context.Context, monitorID string, keepCount int) error
	UptimePercent(ctx context.Context, monitorID string, since time.Time) (float64, error)
}

type Store struct {
	Devices  DeviceRepo
	Networks NetworkRepo
	Scans    ScanRepo
	Alerts   AlertRepo
	Sessions SessionRepo
	Monitors MonitorRepo
}
