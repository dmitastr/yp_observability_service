package postgresstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	serverenvconfig "github.com/dmitastr/yp_observability_service/internal/config/env_parser/server/server_env_config"
	pgerrors "github.com/dmitastr/yp_observability_service/internal/datasources/postgres_storage/pg_err_classifier"
	"github.com/dmitastr/yp_observability_service/internal/logger"
	models "github.com/dmitastr/yp_observability_service/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"

	"github.com/failsafe-go/failsafe-go"
	"github.com/failsafe-go/failsafe-go/retrypolicy"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Conn abstracts pgx transactions creators: pgx.Conn and pgxpool.Pool.
type Conn interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

type Cursor interface {
	Query(ctx context.Context, query string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, query string, args ...any) pgx.Row
}

var maxErrorRetries = 3
var delaySlope = 2
var delayRise = 1
var linearBackOff = func(attempt, slope, rise int) time.Duration {
	return time.Duration(slope*attempt+rise) * time.Second
}

type ExecuteWithRetryFunc func(pgx.Tx) error

type Postgres struct {
	db          *pgxpool.Pool
	retryPolicy retrypolicy.RetryPolicy[any]
}

const query string = `INSERT INTO metrics (name, mtype, value, delta) 
	VALUES (@name, @mtype, @value, @delta) 
	ON CONFLICT ON CONSTRAINT metrics_pkey DO UPDATE SET 
	value = @value, 
    delta = metrics.delta + @delta `

func NewPG(ctx context.Context, cfg serverenvconfig.Config) (*Postgres, error) {

	dbConfig, err := pgxpool.ParseConfig(*cfg.DBUrl)
	if err != nil {
		logger.GetLogger().Fatalf("Failed to parse database config: %v", err)
	}
	dbConfig.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger:   logger.GetLogger(),
		LogLevel: tracelog.LogLevelInfo,
	}

	pool, err := pgxpool.NewWithConfig(ctx, dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db with url=%s: %v", *cfg.DBUrl, err)
	}
	logger.GetLogger().Info("Database connection established successfully")

	delayFunc := func(exec failsafe.ExecutionAttempt[any]) time.Duration {
		return linearBackOff(exec.Attempts(), delaySlope, delayRise)
	}
	retry := retrypolicy.Builder[any]().HandleIf(func(_ any, err error) bool {
		pgerrClassifier := pgerrors.NewPostgresErrorClassifier()
		return pgerrClassifier.Classify(err) == pgerrors.Retriable

	}).WithMaxRetries(maxErrorRetries).
		WithDelayFunc(delayFunc).Build()

	pg := &Postgres{db: pool, retryPolicy: retry}

	return pg, nil
}

func (pg *Postgres) Init() error {
	db, err := sql.Open("postgres", pg.db.Config().ConnString())
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
		return err
	}
	logger.GetLogger().Info("Migration up completed successfully")
	return nil

}

func (pg *Postgres) Ping(ctx context.Context) error {
	return pg.db.Ping(ctx)
}

func (pg *Postgres) Close() error {
	pg.db.Close()
	return nil
}

func (pg *Postgres) ExecuteTX(ctx context.Context, conn Conn, fnc ExecuteWithRetryFunc) error {
	// nolint: wrapcheck
	return failsafe.NewExecutor(pg.retryPolicy).
		WithContext(ctx).
		RunWithExecution(func(exec failsafe.Execution[any]) (err error) { //nolint:contextcheck
			ctxTry := exec.Context()
			tx, err := conn.Begin(ctxTry)
			if err != nil {
				return fmt.Errorf("failed to begin tx: %w", err)
			}

			if err := fnc(tx); err != nil {
				return err
			}

			if err := tx.Commit(ctx); err != nil {
				if err := tx.Rollback(ctx); err != nil {
					logger.GetLogger().Errorf("Failed to rollback tx: %v", err)
				}
				return fmt.Errorf("failed to commit: %w", err)
			}

			return err
		})
}

func (pg *Postgres) Update(ctx context.Context, metric models.Metrics) error {
	fun := func(tx pgx.Tx) error {
		args := metric.ToNamedArgs()

		if _, err := tx.Exec(ctx, query, args); err != nil {
			err := tx.Rollback(ctx)
			if err != nil {
				return fmt.Errorf("failed to rollback tx: %w", err)
			}
			logger.GetLogger().Errorf("unable to insert row: %v", err)
			return err
		}
		return nil
	}
	return pg.ExecuteTX(ctx, pg.db, fun)
}

func (pg *Postgres) BulkUpdate(ctx context.Context, metrics []models.Metrics) error {
	fun := func(tx pgx.Tx) error {
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
		return nil
	}
	return pg.ExecuteTX(ctx, pg.db, fun)
}

func (pg *Postgres) Get(ctx context.Context, name string) (*models.Metrics, error) {
	var metric *models.Metrics

	fun := func(tx pgx.Tx) error {
		m, err := pg.getWithinTx(ctx, name, tx)
		if err != nil {
			return fmt.Errorf("unable to query metrics: %w", err)
		}
		metric = m
		return nil
	}
	err := pg.ExecuteTX(ctx, pg.db, fun)
	return metric, err
}

func (pg *Postgres) GetAll(ctx context.Context) ([]models.Metrics, error) {
	var metrics []models.Metrics
	fun := func(tx pgx.Tx) error {
		m, err := pg.getAllWithinTx(ctx, tx)
		if err != nil {
			return fmt.Errorf("unable to query users: %w", err)
		}
		metrics = m
		return err
	}
	err := pg.ExecuteTX(ctx, pg.db, fun)
	return metrics, err
}

func (pg *Postgres) getWithinTx(ctx context.Context, name string, conn Cursor) (*models.Metrics, error) {
	if conn == nil {
		conn = pg.db
	}

	var metric models.Metrics
	query := `SELECT name, mtype, value, delta FROM metrics WHERE name=@name`

	err := conn.QueryRow(ctx, query, pgx.NamedArgs{"name": name}).Scan(&metric.ID, &metric.MType, &metric.Value, &metric.Delta)
	if err != nil {
		return nil, fmt.Errorf("unable to query metrics: %v", err)
	}
	return &metric, nil
}

func (pg *Postgres) getAllWithinTx(ctx context.Context, conn Cursor) ([]models.Metrics, error) {
	if conn == nil {
		conn = pg.db
	}

	query := `SELECT name, mtype, value, delta FROM metrics`

	rows, err := conn.Query(ctx, query)
	if err != nil {
		logger.GetLogger().Errorf("unable to query users: %v", err)
		return nil, err
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[models.Metrics])
}
