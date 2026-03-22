// internal/store/repository.go
package store

import (
	"context"
	"errors"

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

type Store struct {
	Devices  DeviceRepo
	Networks NetworkRepo
	Scans    ScanRepo
}
