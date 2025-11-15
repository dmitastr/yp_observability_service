package postgresstorage

import (
	"context"
	"testing"

	serverenvconfig "github.com/dmitastr/yp_observability_service/internal/config/env_parser/server/server_env_config"
	"github.com/dmitastr/yp_observability_service/internal/domain/models"
	"github.com/dmitastr/yp_observability_service/internal/mocks/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type MetricsRepoTestSuite struct {
	suite.Suite
	pgContainer *testhelpers.PostgresContainer
	repository  *Postgres
	ctx         context.Context
}

func (suite *MetricsRepoTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	pgContainer, err := testhelpers.CreatePostgresContainer(suite.ctx)
	if err != nil {
		suite.T().Log("Error creating postgres container", err)
	}
	suite.pgContainer = pgContainer
	db, err := NewPG(suite.ctx, &serverenvconfig.Config{DBUrl: &suite.pgContainer.ConnectionString})
	if err != nil {
		suite.T().Log("Error database instance", err)
	}
	if err := db.Init("file://../../../migrations"); err != nil {
		suite.T().Log("Error migrating with prod data", err)
	}
	if err := db.Init("file://../../../migrations/testdata"); err != nil {
		suite.T().Log("Error migrating with test data", err)
	}
	suite.repository = db
}

func (suite *MetricsRepoTestSuite) TearDownSuite() {
	if err := suite.pgContainer.Terminate(suite.ctx); err != nil {
		suite.T().Log("Error while terminating PG container", err)
	}
}

func (suite *MetricsRepoTestSuite) TestPing() {
	t := suite.T()
	assert.NoError(t, suite.repository.Ping(suite.ctx))

}

func (suite *MetricsRepoTestSuite) TestUpdate() {
	t := suite.T()
	m := models.Metrics{ID: "metric", MType: "gauge"}
	m.UpdateDelta(10)

	assert.NoError(t, suite.repository.Update(suite.ctx, m))

	mGot, err := suite.repository.Get(suite.ctx, "metric")
	assert.NoError(t, err)
	assert.Equal(t, m, *mGot)
}

func (suite *MetricsRepoTestSuite) TestBulkUpdate() {
	t := suite.T()
	m := models.Metrics{ID: "metric", MType: "gauge"}
	m.UpdateDelta(10)

	assert.NoError(t, suite.repository.BulkUpdate(suite.ctx, []models.Metrics{m}))

	mGot, err := suite.repository.Get(suite.ctx, "metric")
	assert.NoError(t, err)
	assert.Equal(t, m, *mGot)
}

func (suite *MetricsRepoTestSuite) TestGet() {
	t := suite.T()

	m, err := suite.repository.Get(suite.ctx, "metric")
	assert.NoError(t, err)
	assert.NotNil(t, m)
}

func (suite *MetricsRepoTestSuite) TestGetByID() {
	t := suite.T()

	m, err := suite.repository.GetByID(suite.ctx, []string{"metric"})
	assert.NoError(t, err)
	assert.NotNil(t, m)
}

func (suite *MetricsRepoTestSuite) TestGetAll() {
	t := suite.T()

	m, err := suite.repository.GetAll(suite.ctx)
	assert.NoError(t, err)
	assert.NotNil(t, m)
}

func TestPostgresRepoTestSuite(t *testing.T) {
	suite.Run(t, new(MetricsRepoTestSuite))
}
