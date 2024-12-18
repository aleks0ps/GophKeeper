package db

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestNewDB(t *testing.T) {
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
	logger := log.New(os.Stdout, "NEWDB: ", log.LstdFlags)
	secret := "71D9AE8F80CE194F24580D1B519854BE"
	_, err = NewDB(ctx, connStr, logger, secret)
	if err != nil {
		logger.Fatal(err)
	}
}

func TestRegister(t *testing.T) {
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
	logger := log.New(os.Stdout, "NEWDB: ", log.LstdFlags)
	secret := "71D9AE8F80CE194F24580D1B519854BE"
	pg, err := NewDB(ctx, connStr, logger, secret)
	if err != nil {
		logger.Fatal(err)
	}
	user := User{ID: "", Login: "test", Password: "1234"}
	err = pg.Register(ctx, &user)
	if err != nil {
		log.Fatal(err)
	}
}

func TestLogin(t *testing.T) {
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
	logger := log.New(os.Stdout, "NEWDB: ", log.LstdFlags)
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
	// Login
	err = pg.Login(ctx, &user)
	if err != nil {
		log.Fatal(err)
	}
}
