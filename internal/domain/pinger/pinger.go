package pinger

import (
	"context"

	dbinterface "github.com/dmitastr/yp_observability_service/internal/repository/database"
)

type Pinger interface {
	Ping(context.Context, dbinterface.Database) error
}
