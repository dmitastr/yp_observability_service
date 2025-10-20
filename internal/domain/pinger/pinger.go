package pinger

import (
	"context"

	dbinterface "github.com/dmitastr/yp_observability_service/internal/repository"
)

// Pinger is an interface with Ping method for checking database
type Pinger interface {
	Ping(context.Context, dbinterface.Database) error
}
