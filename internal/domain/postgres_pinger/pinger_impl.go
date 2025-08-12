package postgrespinger

import (
	"context"
	dbinterface "github.com/dmitastr/yp_observability_service/internal/repository/database"
)

type PGDatabasePinger struct{}

func New() *PGDatabasePinger {
	return &PGDatabasePinger{}
}

func (pp *PGDatabasePinger) Ping(ctx context.Context, storage dbinterface.Database) error {
	return storage.Ping(ctx)
}