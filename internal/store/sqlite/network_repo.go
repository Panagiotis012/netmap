// internal/store/sqlite/network_repo.go
package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/netmap/netmap/internal/core/models"
	"github.com/netmap/netmap/internal/store"
)

type NetworkRepo struct {
	db *DB
}

func NewNetworkRepo(db *DB) *NetworkRepo {
	return &NetworkRepo{db: db}
}

var _ store.NetworkRepo = (*NetworkRepo)(nil)

func (r *NetworkRepo) List(ctx context.Context) ([]models.Network, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, name, subnet, gateway, created_at, updated_at FROM networks ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var nets []models.Network
	for rows.Next() {
		var n models.Network
		if err := rows.Scan(&n.ID, &n.Name, &n.Subnet, &n.Gateway, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		nets = append(nets, n)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return nets, nil
}

func (r *NetworkRepo) GetByID(ctx context.Context, id string) (*models.Network, error) {
	var n models.Network
	err := r.db.QueryRowContext(ctx, "SELECT id, name, subnet, gateway, created_at, updated_at FROM networks WHERE id = ?", id).
		Scan(&n.ID, &n.Name, &n.Subnet, &n.Gateway, &n.CreatedAt, &n.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, store.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *NetworkRepo) GetBySubnet(ctx context.Context, subnet string) (*models.Network, error) {
	var n models.Network
	err := r.db.QueryRowContext(ctx, "SELECT id, name, subnet, gateway, created_at, updated_at FROM networks WHERE subnet = ?", subnet).
		Scan(&n.ID, &n.Name, &n.Subnet, &n.Gateway, &n.CreatedAt, &n.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, store.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *NetworkRepo) Create(ctx context.Context, n *models.Network) error {
	now := time.Now()
	n.CreatedAt = now
	n.UpdatedAt = now
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO networks (id, name, subnet, gateway, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		n.ID, n.Name, n.Subnet, n.Gateway, n.CreatedAt, n.UpdatedAt)
	return err
}

func (r *NetworkRepo) Update(ctx context.Context, n *models.Network) error {
	n.UpdatedAt = time.Now()
	res, err := r.db.ExecContext(ctx,
		"UPDATE networks SET name=?, subnet=?, gateway=?, updated_at=? WHERE id=?",
		n.Name, n.Subnet, n.Gateway, n.UpdatedAt, n.ID)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return store.ErrNotFound
	}
	return nil
}

func (r *NetworkRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM networks WHERE id = ?", id)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return store.ErrNotFound
	}
	return nil
}
