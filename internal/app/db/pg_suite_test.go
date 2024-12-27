package db

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"testing"

	"github.com/aleks0ps/GophKeeper/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DBTestSuite struct {
	suite.Suite
	pgContainer *testhelpers.PostgresContainer
	DB          *PG
	ctx         context.Context
}

func (suite *DBTestSuite) SetupSuite() {
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

func (suite *DBTestSuite) TearDownSuite() {
	if err := suite.pgContainer.Terminate(suite.ctx); err != nil {
		log.Fatalf("error terminating postgres container: %s", err)
	}
}

func (suite *DBTestSuite) TestNewDB() {
	t := suite.T()
	_, err := NewDB(suite.ctx, suite.pgContainer.ConnectionString, suite.DB.Logger, suite.DB.Secret)
	assert.NoError(t, err)
}

func (suite *DBTestSuite) TestRegister() {
	t := suite.T()
	user := User{ID: "", Login: "test", Password: "1234"}
	err := suite.DB.Register(suite.ctx, &user)
	assert.NoError(t, err)
}

func (suite *DBTestSuite) TestLogin() {
	t := suite.T()
	user := User{ID: "", Login: "user", Password: "password"}
	// Register for the first time
	err := suite.DB.Register(suite.ctx, &user)
	assert.NoError(t, err)
	// Login
	err = suite.DB.Login(suite.ctx, &user)
	assert.NoError(t, err)
}

func (suite *DBTestSuite) TestGet() {
	t := suite.T()
	user := User{ID: "", Login: "user1", Password: "pass1"}
	// Register for the first time
	err := suite.DB.Register(suite.ctx, &user)
	assert.NoError(t, err)
	// Put data
	pass := Password{Name: "gmail", Password: "Weak password"}
	rec := Record{Type: SRecordPassword}
	jsonPass, _ := json.Marshal(pass)
	rec.Payload = jsonPass
	// Set from cookies in http handler
	user.ID = user.Login
	err = suite.DB.Put(suite.ctx, &user, &rec)
	assert.NoError(t, err)
	// Get data
	pass = Password{Name: "gmail"}
	rec = Record{Type: SRecordPassword}
	jsonPass, _ = json.Marshal(pass)
	rec.Payload = jsonPass
	// Set from cookies in http handler
	user.ID = user.Login
	res, err := suite.DB.Get(suite.ctx, &user, &rec)
	assert.NoError(t, err)
	log.Printf("TestGet: %+v\n", res)
}

func (suite *DBTestSuite) TestPut() {
	t := suite.T()
	user := User{ID: "", Login: "someUser", Password: "somePAss"}
	// Register for the first time
	err := suite.DB.Register(suite.ctx, &user)
	assert.NoError(t, err)
	pass := Password{Name: "gmail", Password: "Weak password"}
	rec := Record{Type: SRecordPassword}
	jsonPass, _ := json.Marshal(pass)
	rec.Payload = jsonPass
	// Set from cookies in http handler
	user.ID = user.Login
	err = suite.DB.Put(suite.ctx, &user, &rec)
	assert.NoError(t, err)
}

func (suite *DBTestSuite) TestList() {
	t := suite.T()
	user := User{ID: "", Login: "userRO", Password: "1234"}
	// Register for the first time
	err := suite.DB.Register(suite.ctx, &user)
	assert.NoError(t, err)
	pass := Password{Name: "gmail", Password: "Weak password"}
	rec := Record{Type: SRecordPassword}
	jsonPass, _ := json.Marshal(pass)
	rec.Payload = jsonPass
	// Set from cookies in http handler
	user.ID = user.Login
	recs, err := suite.DB.List(suite.ctx, &user)
	assert.NoError(t, err)
	log.Printf("TestList: %+v\n", recs)
}

func TestDBTestSuite(t *testing.T) {
	suite.Run(t, new(DBTestSuite))
}
