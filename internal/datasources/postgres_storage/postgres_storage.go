package postgresstorage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	serverenvconfig "github.com/dmitastr/yp_observability_service/internal/config/env_parser/server/server_env_config"
	"github.com/dmitastr/yp_observability_service/internal/errs"
	"github.com/dmitastr/yp_observability_service/internal/logger"
	models "github.com/dmitastr/yp_observability_service/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	db *pgxpool.Pool
	FileName    string
	StreamWrite bool
}

var (
	pgInstance *Postgres
	pgOnce     sync.Once
)

func NewPG(ctx context.Context, cfg serverenvconfig.Config) (*Postgres, error) {
	pgOnce.Do(func() {
		db, err := pgxpool.New(ctx, *cfg.DBUrl)
		if err != nil {
			panic(err)
		}
		streamWrite := false
		if *cfg.StoreInterval == 0 {
			streamWrite = true
		}
		pgInstance = &Postgres{db: db, FileName: *cfg.FileStoragePath, StreamWrite: streamWrite}
		if *cfg.Restore {
			err := pgInstance.Load()
			if err != nil {
				logger.GetLogger().Error(err)
			}
		}
	})
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
	query := `INSERT INTO metrics (name, mtype, value, delta) VALUES (@name, @mtype, @value, @delta)`
	args := metric.ToNamedArgs()
	_, err := pg.db.Exec(ctx, query, args)
	if err != nil {
		logger.GetLogger().Errorf("unable to insert row: %w", err)
		return err
	}

	return nil
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
		return fmt.Errorf("error copying into %s table: %w", tableName, err)
	}

	return nil
} 

func (pg *Postgres) Get(ctx context.Context, name string) (*models.Metrics, error) {
	query := `SELECT name, mtype, value, delta FROM metrics WHERE name=@name`

	rows, err := pg.db.Query(ctx, query, pgx.NamedArgs{"name": name})
	if err != nil {
		logger.GetLogger().Errorf("unable to query users: %w", err)
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
		logger.GetLogger().Errorf("unable to query users: %w", err)
		return nil, err
	}
	defer rows.Close()

	return  pgx.CollectRows(rows, pgx.RowToStructByName[models.Metrics])
}


func (pg *Postgres) CreateFile() *os.File {
	file, err := os.Create(pg.FileName)
	if err != nil {
		logger.GetLogger().Panicf("error while creating file '%s': %s", pg.FileName, err)
	}
	return file
}

func (pg *Postgres) OpenFile() *os.File {
	file, err := os.Open(pg.FileName)
	if err != nil {
		logger.GetLogger().Panicf("error while opening file '%s': %s", pg.FileName, err)
	}
	return file
}

func (pg *Postgres) Flush() error {
	file := pg.CreateFile()
	metrics, err := pg.GetAll(context.TODO())
	if err != nil {
		return err
	}
	if err := json.NewEncoder(file).Encode(metrics); err != nil {
		logger.GetLogger().Error(err)
		return err
	}
	return nil
}

func (pg *Postgres) Load() error {
	file, err := os.Open(pg.FileName)
	if err != nil {
		logger.GetLogger().Error("error while opening file '%s': %s", pg.FileName, err)
		return err
	}

	metrics := make([]models.Metrics, 0)
	if err := json.NewDecoder(file).Decode(&metrics); err != nil {
		logger.GetLogger().Fatal(err)
		return err
	}
	return pg.BulkUpdate(context.TODO(), metrics)
}
