package db

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestGet(t *testing.T) {
	ctx := context.TODO()
	// Prepare database
	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15.3-alpine"),
		postgres.WithDatabase("test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		log.Fatal(err)
	}
	t.Cleanup(func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate pgContainer: %s", err)
		}
	})
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	assert.NoError(t, err)
	logger := log.New(os.Stdout, "PUT: ", log.LstdFlags)
	secret := "71D9AE8F80CE194F24580D1B519854BE"
	pg, err := NewDB(ctx, connStr, logger, secret)
	if err != nil {
		logger.Fatal(err)
	}
	user := User{ID: "", Login: "test", Password: "1234"}
	// Register for the first time
	err = pg.Register(ctx, &user)
	if err != nil {
		log.Fatal(err)
	}
	// Put data
	pass := Password{Name: "gmail", Password: "Weak password"}
	rec := Record{Type: SRecordPassword}
	jsonPass, _ := json.Marshal(pass)
	rec.Payload = jsonPass
	// Set from cookies in http handler
	user.ID = user.Login
	err = pg.Put(ctx, &user, &rec)
	if err != nil {
		log.Fatal(err)
	}
	// Get data
	pass = Password{Name: "gmail"}
	rec = Record{Type: SRecordPassword}
	jsonPass, _ = json.Marshal(pass)
	rec.Payload = jsonPass
	// Set from cookies in http handler
	user.ID = user.Login
	res, err := pg.Get(ctx, &user, &rec)
	if err != nil {
		log.Fatal(err)
	}
	logger.Printf("%+v\n", res)
}
