package postgresstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	serverenvconfig "github.com/dmitastr/yp_observability_service/internal/config/env_parser/server/server_env_config"
	"github.com/dmitastr/yp_observability_service/internal/errs"
	"github.com/dmitastr/yp_observability_service/internal/logger"
	models "github.com/dmitastr/yp_observability_service/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/golang-migrate/migrate/v4"
    "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

type Postgres struct {
	db *pgxpool.Pool
}

var (
	pgInstance *Postgres
	pgOnce     sync.Once
)

func NewPG(ctx context.Context, cfg serverenvconfig.Config) (*Postgres, error) {
	var errConnecting error
	pgOnce.Do(func() {
		dbConfig, err := pgxpool.ParseConfig(*cfg.DBUrl)
		if err != nil {
			logger.GetLogger().Fatalf("Failed to parse database config: %v", err)
		}
		pool, err := pgxpool.NewWithConfig(ctx, dbConfig)
		if err != nil {
			errConnecting = fmt.Errorf("failed to connect to db with url=%s: %v", *cfg.DBUrl, err)
			return 
		}
		logger.GetLogger().Info("Database connection established successfully")
		
		pgInstance = &Postgres{db: pool}
		
	})
	
	db, err := sql.Open("postgres", *cfg.DBUrl)
	if err != nil {
		logger.GetLogger().Fatalf("Unable to connect to database: %v", err)
	}
	// Create migration instance
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		logger.GetLogger().Fatal(err)
	}
	
	// Point to your migration files. Here we're using local files, but it could be other sources.
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		logger.GetLogger().Fatal(err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		logger.GetLogger().Fatalf("Migration up failed: %v", err)
	}
	fmt.Println("Migration up completed successfully")

	return pgInstance, errConnecting
}



func (pg *Postgres) Ping(ctx context.Context) error {
	return pg.db.Ping(ctx)
}

func (pg *Postgres) Close() error {
	pg.db.Close()
	return nil
}

func (pg *Postgres) Update(ctx context.Context, metric models.Metrics) error {
	tx, err := pg.db.Begin(ctx)
	if err != nil {
		return err
	}
	query := `INSERT INTO metrics (name, mtype, value, delta) 
	VALUES (@name, @mtype, @value, @delta)
	ON CONFLICT ON CONSTRAINT metrics_pkey DO UPDATE SET
    value = @value,
    delta = @delta `

	args := metric.ToNamedArgs()
	
	if _, err := tx.Exec(ctx, query, args); err != nil {
		tx.Rollback(ctx)
		logger.GetLogger().Errorf("unable to insert row: %v", err)
		return err
	}

	return tx.Commit(ctx)
}

func (pg *Postgres) BulkUpdate(ctx context.Context, metrics []models.Metrics) error {
	entries := [][]any{}
	columns := []string{"name", "mtype", "value", "delta"}
	tableName := "metrics"

	for _, metric := range metrics {
		entries = append(entries, []any{metric.ID, metric.MType, metric.Value, metric.Delta})
	}

	_, err := pg.db.CopyFrom(
		ctx,
		pgx.Identifier{tableName},
		columns,
		pgx.CopyFromRows(entries),
	)

	if err != nil {
		return fmt.Errorf("error copying into %s table: %v", tableName, err)
	}

	return nil
} 

func (pg *Postgres) Get(ctx context.Context, name string) (*models.Metrics, error) {
	query := `SELECT name, mtype, value, delta FROM metrics WHERE name=@name`

	rows, err := pg.db.Query(ctx, query, pgx.NamedArgs{"name": name})
	if err != nil {
		logger.GetLogger().Errorf("unable to query users: %v", err)
		return nil, err
	}
	defer rows.Close()

	metrics, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Metrics])
	if err != nil {
		return nil, err
	}
	if len(metrics) > 1 {
		return &metrics[0], nil
	}
	return nil, errs.ErrorMetricDoesNotExist
}

func (pg *Postgres) GetAll(ctx context.Context) ([]models.Metrics, error) {
	query := `SELECT name, mtype, value, delta FROM metrics`

	rows, err := pg.db.Query(ctx, query)
	if err != nil {
		logger.GetLogger().Errorf("unable to query users: %v", err)
		return nil, err
	}
	defer rows.Close()

	return  pgx.CollectRows(rows, pgx.RowToStructByName[models.Metrics])
}

// dummy function to satisfy interface
func (pg *Postgres) RunBackup() {}