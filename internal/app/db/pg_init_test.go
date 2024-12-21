package db

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/aleks0ps/GophKeeper/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DbTestSuite struct {
	suite.Suite
	pgContainer *testhelpers.PostgresContainer
	DB          *PG
	ctx         context.Context
}

func (suite *DbTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	pgContainer, err := testhelpers.CreatePostgresContainer(suite.ctx)
	if err != nil {
		log.Fatal(err)
	}
	suite.pgContainer = pgContainer
	logger := log.New(os.Stdout, "TESTDB: ", log.LstdFlags)
	secret := "71D9AE8F80CE194F24580D1B519854BE"
	DB, err := NewDB(suite.ctx, suite.pgContainer.ConnectionString, logger, secret)
	if err != nil {
		logger.Fatal(err)
	}
	suite.DB = DB
}

func (suite *DbTestSuite) TearDownSuite() {
	if err := suite.pgContainer.Terminate(suite.ctx); err != nil {
		log.Fatalf("error terminating postgres container: %s", err)
	}
}

func (suite *DbTestSuite) TestNewDB() {
	t := suite.T()
	_, err := NewDB(suite.ctx, suite.pgContainer.ConnectionString, suite.DB.Logger, suite.DB.Secret)
	assert.NoError(t, err)
}

func (suite *DbTestSuite) TestRegister() {
	t := suite.T()
	user := User{ID: "", Login: "test", Password: "1234"}
	err := suite.DB.Register(suite.ctx, &user)
	assert.NoError(t, err)
}

func (suite *DbTestSuite) TestLogin() {
	t := suite.T()
	user := User{ID: "", Login: "user", Password: "password"}
	// Register for the first time
	err := suite.DB.Register(suite.ctx, &user)
	assert.NoError(t, err)
	// Login
	err = suite.DB.Login(suite.ctx, &user)
	assert.NoError(t, err)
}

func TestDbTestSuite(t *testing.T) {
	suite.Run(t, new(DbTestSuite))
}
