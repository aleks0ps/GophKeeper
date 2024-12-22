package svc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/aleks0ps/GophKeeper/internal/app/db"
	"github.com/aleks0ps/GophKeeper/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"golang.org/x/net/publicsuffix"
)

type SvcTestSuite struct {
	suite.Suite
	pgContainer *testhelpers.PostgresContainer
	Logger      *log.Logger
	DB          *db.PG
	ctx         context.Context
}

func (suite *SvcTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	pgContainer, err := testhelpers.CreatePostgresContainer(suite.ctx)
	if err != nil {
		log.Fatal(err)
	}
	suite.pgContainer = pgContainer
	suite.Logger = log.New(os.Stdout, "TESTDB: ", log.LstdFlags)
	secret := "71D9AE8F80CE194F24580D1B519854BE"
	DB, err := db.NewDB(suite.ctx, suite.pgContainer.ConnectionString, suite.Logger, secret)
	if err != nil {
		suite.Logger.Fatal(err)
	}
	suite.DB = DB
}

func (suite *SvcTestSuite) TearDownSuite() {
	if err := suite.pgContainer.Terminate(suite.ctx); err != nil {
		log.Fatalf("error terminating postgres container: %s", err)
	}
}

func (suite *SvcTestSuite) TestRegister() {
	t := suite.T()
	// Create service
	s := Svc{Logger: suite.Logger, DB: suite.DB, DataDir: "/tmp"}
	// Start test server
	ts := httptest.NewServer(http.HandlerFunc(s.Register))
	defer ts.Close()
	u, err := url.Parse(ts.URL)
	assert.NoError(t, err)
	// cookies placeholder
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	assert.NoError(t, err)
	// create a client
	client := &http.Client{
		Jar: jar,
	}
	// specify test user
	user := db.User{ID: "", Login: "svc-test", Password: "1234"}
	payload, err := json.Marshal(&user)
	assert.NoError(t, err)
	buf := bytes.NewBuffer(payload)
	assert.NoError(t, err)
	// make a request
	resp, err := client.Post(u.String(), "application/json", buf)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, http.StatusOK, "Should be ok")
	log.Printf("TestRegister: %v\n", resp.Status)
}

func (suite *SvcTestSuite) TestLogin() {
	t := suite.T()
	// Create service
	s := Svc{Logger: suite.Logger, DB: suite.DB, DataDir: "/tmp"}
	// Start test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RequestURI() == "/register" {
			s.Register(w, r)
		} else if r.URL.RequestURI() == "/login" {
			s.Login(w, r)
		}
	}))
	defer ts.Close()
	u, _ := url.Parse(ts.URL)
	// cookies placeholder
	jar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	// create a client
	client := &http.Client{
		Jar: jar,
	}
	// specify test user
	user := db.User{ID: "", Login: "svc-test1", Password: "1234"}
	payload, _ := json.Marshal(&user)
	buf := bytes.NewBuffer(payload)
	// make a reg request
	resp, err := client.Post(u.String()+"/register", "application/json", buf)
	assert.NoError(t, err)
	log.Printf("TestRegister: %v\n", resp.Status)
	// make buffer one more time
	buf = bytes.NewBuffer(payload)
	resp, err = client.Post(u.String()+"/login", "application/json", buf)
	assert.NoError(t, err)
	log.Printf("TestLogin: %v\n", resp.Status)
}

func (suite *SvcTestSuite) TestPut() {
	t := suite.T()
	s := Svc{Logger: suite.Logger, DB: suite.DB, DataDir: "/tmp"}
	// Start test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RequestURI() == "/register" {
			s.Register(w, r)
		} else if r.URL.RequestURI() == "/put" {
			s.Put(w, r)
		}
	}))
	defer ts.Close()
	u, _ := url.Parse(ts.URL)
	// cookies placeholder
	jar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	// create a client
	client := &http.Client{
		Jar: jar,
	}
	// specify test user
	user := db.User{ID: "", Login: "svc-test12", Password: "1234"}
	payload, _ := json.Marshal(&user)
	buf := bytes.NewBuffer(payload)
	// make a reg request
	resp, err := client.Post(u.String()+"/register", "application/json", buf)
	assert.NoError(t, err)
	log.Printf("TestRegister: %v\n", resp.Status)
	// Put some data
	rec := db.Record{Type: db.SRecordPassword}
	pass := db.Password{Name: "wifi", Password: "123456"}
	jsonPass, _ := json.Marshal(pass)
	rec.Payload = jsonPass
	payload, _ = json.Marshal(rec)
	buf = bytes.NewBuffer(payload)
	resp, err = client.Post(u.String()+"/put", "application/json", buf)
	assert.NoError(t, err)
	log.Printf("TestPut: %v\n", resp.Status)
}

func (suite *SvcTestSuite) TestGet() {
	t := suite.T()
	// Create service
	s := Svc{Logger: suite.Logger, DB: suite.DB, DataDir: "/tmp"}
	// Start test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RequestURI() == "/register" {
			s.Register(w, r)
		} else if r.URL.RequestURI() == "/get" {
			s.Get(w, r)
		} else if r.URL.RequestURI() == "/put" {
			s.Put(w, r)
		}
	}))
	defer ts.Close()
	u, _ := url.Parse(ts.URL)
	// cookies placeholder
	jar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	// create a client
	client := &http.Client{
		Jar: jar,
	}
	// specify test user
	user := db.User{ID: "", Login: "svc-test17", Password: "1234"}
	payload, _ := json.Marshal(&user)
	buf := bytes.NewBuffer(payload)
	// make a reg request
	resp, err := client.Post(u.String()+"/register", "application/json", buf)
	assert.NoError(t, err)
	log.Printf("TestRegister: %v\n", resp.Status)
	// Put some data
	rec := db.Record{Type: db.SRecordPassword}
	pass := db.Password{Name: "wifi", Password: "123456"}
	jsonPass, _ := json.Marshal(pass)
	rec.Payload = jsonPass
	payload, _ = json.Marshal(rec)
	buf = bytes.NewBuffer(payload)
	resp, err = client.Post(u.String()+"/put", "application/json", buf)
	assert.NoError(t, err)
	log.Printf("TestPut: %v\n", resp.Status)
	// Get some data
	rec = db.Record{Type: db.SRecordPassword}
	pass = db.Password{Name: "wifi"}
	jsonPass, _ = json.Marshal(pass)
	rec.Payload = jsonPass
	payload, _ = json.Marshal(rec)
	buf = bytes.NewBuffer(payload)
	resp, err = client.Post(u.String()+"/get", "application/json", buf)
	assert.NoError(t, err)
	log.Printf("TestGet: %v\n", resp.Status)
}

func (suite *SvcTestSuite) TestList() {
	t := suite.T()
	// Create service
	s := Svc{Logger: suite.Logger, DB: suite.DB, DataDir: "/tmp"}
	// Start test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RequestURI() == "/register" {
			s.Register(w, r)
		} else if r.URL.RequestURI() == "/list" {
			s.List(w, r)
		} else if r.URL.RequestURI() == "/put" {
			s.Put(w, r)
		}
	}))
	defer ts.Close()
	u, _ := url.Parse(ts.URL)
	// cookies placeholder
	jar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	// create a client
	client := &http.Client{
		Jar: jar,
	}
	// specify test user
	user := db.User{ID: "", Login: "svc-test88", Password: "1234"}
	payload, _ := json.Marshal(&user)
	buf := bytes.NewBuffer(payload)
	// make a reg request
	resp, err := client.Post(u.String()+"/register", "application/json", buf)
	assert.NoError(t, err)
	log.Printf("TestRegister: %v\n", resp.Status)
	// Put some data
	rec := db.Record{Type: db.SRecordPassword}
	pass := db.Password{Name: "wifi", Password: "123456"}
	jsonPass, _ := json.Marshal(pass)
	rec.Payload = jsonPass
	payload, _ = json.Marshal(rec)
	buf = bytes.NewBuffer(payload)
	resp, err = client.Post(u.String()+"/put", "application/json", buf)
	assert.NoError(t, err)
	log.Printf("TestPut: %v\n", resp.Status)
	// List some data
	resp, err = client.Post(u.String()+"/list", "application/json", nil)
	assert.NoError(t, err)
	log.Printf("TestList: %v\n", resp.Status)
}

func (suite *SvcTestSuite) TestPutBinary() {
	t := suite.T()
	s := Svc{Logger: suite.Logger, DB: suite.DB, DataDir: "/tmp"}
	// Start test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RequestURI() == "/register" {
			s.Register(w, r)
		} else if r.URL.RequestURI() == "/put/binary" {
			s.PutBinary(w, r)
		}
	}))
	defer ts.Close()
	u, _ := url.Parse(ts.URL)
	// cookies placeholder
	jar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	// create a client
	client := &http.Client{
		Jar: jar,
	}
	// specify test user
	user := db.User{ID: "", Login: "svc-test56", Password: "1234"}
	payload, _ := json.Marshal(&user)
	buf := bytes.NewBuffer(payload)
	// make a reg request
	resp, err := client.Post(u.String()+"/register", "application/json", buf)
	assert.NoError(t, err)
	// Put some data
	fileName := "Article-OPEN-SOURCE-SOFTWARE-THE-SUCCESS-OF-AN-ALTERNATIVE-INTELLECTUAL-PROPERTY-INCENTIVE-PARADIGM.pdf"
	filePath := "../../../files/" + fileName
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)
	contentType := writer.FormDataContentType()
	go func() {
		file, err := os.Open(filePath)
		assert.NoError(t, err)
		defer file.Close()
		part, err := writer.CreateFormFile("file", fileName)
		assert.NoError(t, err)
		for {
			_, err = io.CopyN(part, file, 4096)
			if err == io.ErrUnexpectedEOF || err == io.EOF {
				break
			}
			assert.NoError(t, err)
		}
		pw.CloseWithError(writer.Close())
	}()
	req, err := http.NewRequest("POST", u.String()+"/put/binary", pr)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", contentType)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	fmt.Printf("%v\n", resp.Status)
}

// Run test suite
func TestSvcTestSuite(t *testing.T) {
	suite.Run(t, new(SvcTestSuite))
}
