// internal/store/sqlite/network_repo_test.go
package sqlite

import (
	"context"
	"testing"

	"github.com/netmap/netmap/internal/core/models"
)

func TestNetworkCreateAndList(t *testing.T) {
	db := testDB(t)
	repo := NewNetworkRepo(db)
	ctx := context.Background()

	net := &models.Network{
		ID: "net-1", Name: "Home LAN", Subnet: "192.168.1.0/24", Gateway: "192.168.1.1",
	}
	if err := repo.Create(ctx, net); err != nil {
		t.Fatal(err)
	}

	list, err := repo.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].Name != "Home LAN" {
		t.Errorf("unexpected list: %+v", list)
	}
}
