package postgresstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	serverenvconfig "github.com/dmitastr/yp_observability_service/internal/config/env_parser/server/server_env_config"
	"github.com/dmitastr/yp_observability_service/internal/domain/models"
	"github.com/dmitastr/yp_observability_service/internal/logger"
	"github.com/dmitastr/yp_observability_service/internal/repository/postgres_storage/pg_err_classifier"
	"github.com/jackc/pgx/v5/pgtype"

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

var maxErrorRetries int = 3
var delaySlope int = 2
var delayRise int = 1
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
    delta = @delta `

func NewPG(ctx context.Context, cfg serverenvconfig.Config) (*Postgres, error) {

	dbConfig, err := pgxpool.ParseConfig(*cfg.DBUrl)
	if err != nil {
		logger.Fatalf("Failed to parse memstorage config: %v", err)
	}
	dbConfig.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger:   logger.GetLogger(),
		LogLevel: tracelog.LogLevelInfo,
	}

	pool, err := pgxpool.NewWithConfig(ctx, dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db with url=%s: %v", *cfg.DBUrl, err)
	}
	logger.Info("Database connection established successfully")

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

func (pg *Postgres) Init(migrationDir string) error {
	db, err := sql.Open("postgres", pg.db.Config().ConnString())
	if err != nil {
		logger.Fatalf("Unable to connect to memstorage: %v", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		logger.Fatal(err)
	}

	m, err := migrate.NewWithDatabaseInstance(migrationDir, "postgres", driver)
	if err != nil {
		logger.Fatal(err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		logger.Fatalf("Migration up failed: %v", err)
	}
	if err := db.Close(); err != nil {
		return err
	}
	logger.Info("Migration up completed successfully")
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
			ctx := exec.Context()
			tx, err := conn.Begin(ctx)
			if err != nil {
				return fmt.Errorf("failed to begin tx: %w", err)
			}

			defer tx.Rollback(ctx)

			if err := fnc(tx); err != nil {
				return err
			}

			if err := tx.Commit(ctx); err != nil {
				return fmt.Errorf("failed to commit: %w", err)
			}

			return err
		})
}

func (pg *Postgres) Update(ctx context.Context, metric models.Metrics) error {
	fun := func(tx pgx.Tx) error {
		args := metric.ToNamedArgs()

		if _, err := tx.Exec(ctx, query, args); err != nil {
			tx.Rollback(ctx)
			logger.Errorf("unable to insert row: %v", err)
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

func (pg *Postgres) GetById(ctx context.Context, names []string) ([]models.Metrics, error) {
	var metrics []models.Metrics
	var err error

	fun := func(tx pgx.Tx) error {
		metrics, err = pg.getByIdWithinTx(ctx, names, tx)
		if err != nil {
			return fmt.Errorf("unable to query metrics: %w", err)
		}
		return nil
	}
	err = pg.ExecuteTX(ctx, pg.db, fun)
	return metrics, err
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

func (pg *Postgres) getByIdWithinTx(ctx context.Context, names []string, conn Cursor) ([]models.Metrics, error) {
	if conn == nil {
		conn = pg.db
	}

	namesArg := &pgtype.Array[string]{}
	namesArg.Elements = names

	query := `SELECT name, mtype, value, delta FROM metrics WHERE name = ANY ($1)`

	rows, err := conn.Query(ctx, query, namesArg)
	if err != nil {
		logger.Errorf("unable to query metrics: %v", err)
		return nil, err
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[models.Metrics])
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
		logger.Errorf("unable to query metrics: %v", err)
		return nil, err
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[models.Metrics])
}
