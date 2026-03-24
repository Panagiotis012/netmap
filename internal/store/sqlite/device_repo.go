// internal/store/sqlite/device_repo.go
package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/netmap/netmap/internal/core/models"
	"github.com/netmap/netmap/internal/store"
)

type DeviceRepo struct {
	db *DB
}

func NewDeviceRepo(db *DB) *DeviceRepo {
	return &DeviceRepo{db: db}
}

func (r *DeviceRepo) Create(ctx context.Context, d *models.Device) error {
	ips, _ := json.Marshal(d.IPAddresses)
	macs, _ := json.Marshal(d.MACAddresses)
	tags, _ := json.Marshal(d.Tags)
	ports, _ := json.Marshal(d.Ports)
	meta := d.Metadata
	if meta == nil {
		meta = json.RawMessage("{}")
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO devices (id, hostname, ip_addresses, mac_addresses, os, status, discovery_method, first_seen_at, last_seen_at, tags, ports, latency_ms, group_id, metadata, map_x, map_y, network_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		d.ID, d.Hostname, string(ips), string(macs), d.OS, d.Status, d.DiscoveryMethod,
		d.FirstSeenAt, d.LastSeenAt, string(tags), string(ports), d.LatencyMs, d.GroupID, string(meta), d.MapX, d.MapY, d.NetworkID,
	)
	return err
}

func (r *DeviceRepo) GetByID(ctx context.Context, id string) (*models.Device, error) {
	return r.scanDevice(r.db.QueryRowContext(ctx,
		`SELECT id, hostname, ip_addresses, mac_addresses, os, status, discovery_method, first_seen_at, last_seen_at, tags, ports, latency_ms, group_id, metadata, map_x, map_y, network_id FROM devices WHERE id = ?`, id))
}

func (r *DeviceRepo) GetByMAC(ctx context.Context, mac string) (*models.Device, error) {
	return r.scanDevice(r.db.QueryRowContext(ctx,
		`SELECT id, hostname, ip_addresses, mac_addresses, os, status, discovery_method, first_seen_at, last_seen_at, tags, ports, latency_ms, group_id, metadata, map_x, map_y, network_id FROM devices WHERE mac_addresses LIKE ?`,
		fmt.Sprintf("%%%s%%", mac)))
}

func (r *DeviceRepo) GetByHostname(ctx context.Context, hostname string) (*models.Device, error) {
	return r.scanDevice(r.db.QueryRowContext(ctx,
		`SELECT id, hostname, ip_addresses, mac_addresses, os, status, discovery_method, first_seen_at, last_seen_at, tags, ports, latency_ms, group_id, metadata, map_x, map_y, network_id FROM devices WHERE hostname = ?`, hostname))
}

func (r *DeviceRepo) GetByIP(ctx context.Context, ip string) (*models.Device, error) {
	return r.scanDevice(r.db.QueryRowContext(ctx,
		`SELECT id, hostname, ip_addresses, mac_addresses, os, status, discovery_method, first_seen_at, last_seen_at, tags, ports, latency_ms, group_id, metadata, map_x, map_y, network_id FROM devices WHERE ip_addresses LIKE ?`,
		fmt.Sprintf("%%%s%%", ip)))
}

func (r *DeviceRepo) scanDevice(row *sql.Row) (*models.Device, error) {
	var d models.Device
	var ips, macs, tags, ports, meta string
	err := row.Scan(&d.ID, &d.Hostname, &ips, &macs, &d.OS, &d.Status, &d.DiscoveryMethod,
		&d.FirstSeenAt, &d.LastSeenAt, &tags, &ports, &d.LatencyMs, &d.GroupID, &meta, &d.MapX, &d.MapY, &d.NetworkID)
	if err == sql.ErrNoRows {
		return nil, store.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(ips), &d.IPAddresses); err != nil {
		return nil, fmt.Errorf("unmarshal ip_addresses: %w", err)
	}
	if err := json.Unmarshal([]byte(macs), &d.MACAddresses); err != nil {
		return nil, fmt.Errorf("unmarshal mac_addresses: %w", err)
	}
	if err := json.Unmarshal([]byte(tags), &d.Tags); err != nil {
		return nil, fmt.Errorf("unmarshal tags: %w", err)
	}
	if ports != "" && ports != "null" {
		_ = json.Unmarshal([]byte(ports), &d.Ports)
	}
	d.Metadata = json.RawMessage(meta)
	return &d, nil
}

func (r *DeviceRepo) List(ctx context.Context, params models.ListParams) (*models.ListResult[models.Device], error) {
	where := []string{"1=1"}
	args := []interface{}{}

	if params.Status != "" {
		where = append(where, "status = ?")
		args = append(args, params.Status)
	}
	if params.Search != "" {
		where = append(where, "(hostname LIKE ? OR ip_addresses LIKE ?)")
		s := "%" + params.Search + "%"
		args = append(args, s, s)
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM devices WHERE "+whereClause, args...).Scan(&total)
	if err != nil {
		return nil, err
	}

	allowedSorts := map[string]bool{
		"hostname": true, "status": true, "last_seen_at": true,
		"first_seen_at": true, "os": true, "ip_addresses": true,
	}
	sort := "last_seen_at"
	if params.Sort != "" && allowedSorts[params.Sort] {
		sort = params.Sort
	}
	order := "DESC"
	if params.Order == "asc" {
		order = "ASC"
	}
	limit := 50
	if params.Limit > 0 {
		limit = params.Limit
	}
	page := 1
	if params.Page > 0 {
		page = params.Page
	}
	offset := (page - 1) * limit

	query := fmt.Sprintf("SELECT id, hostname, ip_addresses, mac_addresses, os, status, discovery_method, first_seen_at, last_seen_at, tags, ports, latency_ms, group_id, metadata, map_x, map_y, network_id FROM devices WHERE %s ORDER BY %s %s LIMIT ? OFFSET ?",
		whereClause, sort, order)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []models.Device
	for rows.Next() {
		var d models.Device
		var ips, macs, tags, ports, meta string
		if err := rows.Scan(&d.ID, &d.Hostname, &ips, &macs, &d.OS, &d.Status, &d.DiscoveryMethod,
			&d.FirstSeenAt, &d.LastSeenAt, &tags, &ports, &d.LatencyMs, &d.GroupID, &meta, &d.MapX, &d.MapY, &d.NetworkID); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(ips), &d.IPAddresses); err != nil {
			return nil, fmt.Errorf("unmarshal ip_addresses: %w", err)
		}
		if err := json.Unmarshal([]byte(macs), &d.MACAddresses); err != nil {
			return nil, fmt.Errorf("unmarshal mac_addresses: %w", err)
		}
		if err := json.Unmarshal([]byte(tags), &d.Tags); err != nil {
			return nil, fmt.Errorf("unmarshal tags: %w", err)
		}
		if ports != "" && ports != "null" {
			_ = json.Unmarshal([]byte(ports), &d.Ports)
		}
		d.Metadata = json.RawMessage(meta)
		devices = append(devices, d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	totalPages := (total + limit - 1) / limit
	return &models.ListResult[models.Device]{
		Items: devices, Total: total, Page: page, TotalPages: totalPages,
	}, nil
}

func (r *DeviceRepo) Update(ctx context.Context, d *models.Device) error {
	ips, _ := json.Marshal(d.IPAddresses)
	macs, _ := json.Marshal(d.MACAddresses)
	tags, _ := json.Marshal(d.Tags)
	ports, _ := json.Marshal(d.Ports)
	meta := d.Metadata
	if meta == nil {
		meta = json.RawMessage("{}")
	}

	_, err := r.db.ExecContext(ctx,
		`UPDATE devices SET hostname=?, ip_addresses=?, mac_addresses=?, os=?, status=?, last_seen_at=?, tags=?, ports=?, latency_ms=?, group_id=?, metadata=?, map_x=?, map_y=?, network_id=? WHERE id=?`,
		d.Hostname, string(ips), string(macs), d.OS, d.Status, d.LastSeenAt,
		string(tags), string(ports), d.LatencyMs, d.GroupID, string(meta), d.MapX, d.MapY, d.NetworkID, d.ID,
	)
	return err
}

func (r *DeviceRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM devices WHERE id = ?", id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return store.ErrNotFound
	}
	return nil
}

func (r *DeviceRepo) UpdateStatus(ctx context.Context, id string, status models.DeviceStatus) error {
	res, err := r.db.ExecContext(ctx, "UPDATE devices SET status = ? WHERE id = ?", status, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return store.ErrNotFound
	}
	return nil
}

func (r *DeviceRepo) UpdatePosition(ctx context.Context, id string, x, y float64) error {
	res, err := r.db.ExecContext(ctx, "UPDATE devices SET map_x = ?, map_y = ? WHERE id = ?", x, y, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return store.ErrNotFound
	}
	return nil
}

func (r *DeviceRepo) CountByStatus(ctx context.Context) (online, offline, unknown int, err error) {
	rows, err := r.db.QueryContext(ctx, "SELECT status, COUNT(*) FROM devices GROUP BY status")
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var status string
		var count int
		if err = rows.Scan(&status, &count); err != nil {
			return
		}
		switch models.DeviceStatus(status) {
		case models.StatusOnline:
			online = count
		case models.StatusOffline:
			offline = count
		case models.StatusUnknown:
			unknown = count
		}
	}
	err = rows.Err()
	return
}

// Ensure DeviceRepo satisfies the store.DeviceRepo interface at compile time.
var _ store.DeviceRepo = (*DeviceRepo)(nil)
