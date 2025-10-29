package postgrespinger

import (
	"context"

	dbinterface "github.com/dmitastr/yp_observability_service/internal/repository"
)

// PGDatabasePinger implements [pinger.Pinger] interface
type PGDatabasePinger struct{}

func New() *PGDatabasePinger {
	return &PGDatabasePinger{}
}

// Ping checks db availability by calling its Ping method
func (pp *PGDatabasePinger) Ping(ctx context.Context, storage dbinterface.Database) error {
	return storage.Ping(ctx)
}
