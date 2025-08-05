package postgresstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	serverenvconfig "github.com/dmitastr/yp_observability_service/internal/config/env_parser/server/server_env_config"
	"github.com/dmitastr/yp_observability_service/internal/logger"
	models "github.com/dmitastr/yp_observability_service/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Postgres struct {
	db *pgxpool.Pool
}

var pgInstance *Postgres

const query string = `INSERT INTO metrics (name, mtype, value, delta) 
	VALUES (@name, @mtype, @value, @delta)
	ON CONFLICT ON CONSTRAINT metrics_pkey DO UPDATE SET
    value = @value,
    delta = @delta `

func NewPG(ctx context.Context, cfg serverenvconfig.Config) (*Postgres, error) {
	db, err := sql.Open("postgres", *cfg.DBUrl)
	if err != nil {
		logger.GetLogger().Fatalf("Unable to connect to database: %v", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		logger.GetLogger().Fatal(err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
	if err != nil {
		logger.GetLogger().Fatal(err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		logger.GetLogger().Fatalf("Migration up failed: %v", err)
	}
	if err := db.Close(); err != nil {
		return nil, err
	}
	logger.GetLogger().Info("Migration up completed successfully")

	dbConfig, err := pgxpool.ParseConfig(*cfg.DBUrl)
	if err != nil {
		logger.GetLogger().Fatalf("Failed to parse database config: %v", err)
	}
	dbConfig.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger:   logger.GetLogger(),
		LogLevel: tracelog.LogLevelError,
	}

	pool, err := pgxpool.NewWithConfig(ctx, dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db with url=%s: %v", *cfg.DBUrl, err)
	}
	logger.GetLogger().Info("Database connection established successfully")

	pgInstance = &Postgres{db: pool}

	return pgInstance, nil
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

	args := metric.ToNamedArgs()

	if _, err := tx.Exec(ctx, query, args); err != nil {
		tx.Rollback(ctx)
		logger.GetLogger().Errorf("unable to insert row: %v", err)
		return err
	}

	return tx.Commit(ctx)
}

func (pg *Postgres) BulkUpdate(ctx context.Context, metrics []models.Metrics) error {
	tx, err := pg.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("unable to start transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	batch := &pgx.Batch{}
	for _, metric := range metrics {
		args := metric.ToNamedArgs()
		batch.Queue(query, args)
	}
	br := tx.SendBatch(ctx, batch)

	for range metrics {
		_, err := br.Exec()
		if err != nil {
			return fmt.Errorf("batch exec failed at item: %w", err)
		}
	}
	if err := br.Close(); err != nil {
		return fmt.Errorf("failed to close batch results: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}

	return nil
}

func (pg *Postgres) Get(ctx context.Context, name string) (*models.Metrics, error) {
	var metric models.Metrics
	query := `SELECT name, mtype, value, delta FROM metrics WHERE name=@name`

	err := pg.db.QueryRow(ctx, query, pgx.NamedArgs{"name": name}).Scan(&metric.ID, &metric.MType, &metric.Value, &metric.Delta)
	if err != nil {
		return nil, fmt.Errorf("unable to query users: %v", err)
	}
	return &metric, nil
}

func (pg *Postgres) GetAll(ctx context.Context) ([]models.Metrics, error) {
	query := `SELECT name, mtype, value, delta FROM metrics`

	rows, err := pg.db.Query(ctx, query)
	if err != nil {
		logger.GetLogger().Errorf("unable to query users: %v", err)
		return nil, err
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[models.Metrics])
}

// dummy function to satisfy interface
func (pg *Postgres) RunBackup() {}
