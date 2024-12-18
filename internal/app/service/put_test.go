package svc

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/aleks0ps/GophKeeper/internal/app/db"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/net/publicsuffix"
)

func TestPut(t *testing.T) {
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
	logger := log.New(os.Stdout, "AUTH: ", log.LstdFlags)
	secret := "71D9AE8F80CE194F24580D1B519854BE"
	pg, err := db.NewDB(ctx, connStr, logger, secret)
	if err != nil {
		logger.Fatal(err)
	}
	// Create service
	s := Svc{Logger: logger, DB: pg, DataDir: "/tmp"}
	_ = s
	// Start test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RequestURI() == "/register" {
			s.Register(w, r)
		} else if r.URL.RequestURI() == "/put" {
			s.Put(w, r)
		}
	}))
	defer ts.Close()
	u, err := url.Parse(ts.URL)
	if err != nil {
		log.Fatal(err)
	}
	// cookies placeholder
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		log.Fatal(err)
	}
	// create a client
	client := &http.Client{
		Jar: jar,
	}
	// specify test user
	user := db.User{ID: "", Login: "test", Password: "1234"}
	payload, err := json.Marshal(&user)
	if err != nil {
		log.Fatal(err)
	}
	buf := bytes.NewBuffer(payload)
	// make a reg request
	resp, err := client.Post(u.String()+"/register", "application/json", buf)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("TestRegister: %v\n", resp.Status)
	// Put some data
	rec := db.Record{Type: db.SRecordPassword}
	pass := db.Password{Name: "wifi", Password: "123456"}
	jsonPass, _ := json.Marshal(pass)
	rec.Payload = jsonPass
	payload, _ = json.Marshal(rec)
	buf = bytes.NewBuffer(payload)
	resp, err = client.Post(u.String()+"/put", "application/json", buf)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("TestPut: %v\n", resp.Status)
}
