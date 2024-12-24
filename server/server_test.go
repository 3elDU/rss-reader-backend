package server_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/3elDU/rss-reader-backend/server"
	"github.com/3elDU/rss-reader-backend/token"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/jmoiron/sqlx"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "modernc.org/sqlite"
)

var TestDB *sqlx.DB
var TestServer *httptest.Server
var ServerStruct *server.Server

// Utility function that enables authentication for the specific test.
// Returns a function to call with defer
func EnableAuthForThisTest() func() {
	server.NoAuth = false
	return func() {
		server.NoAuth = true
	}
}

func TestMain(t *testing.M) {
	db := sqlx.MustConnect("sqlite", ":memory:")
	defer db.Close()

	// Run database migrations
	driver, err := sqlite.WithInstance(db.DB, &sqlite.Config{})
	if err != nil {
		panic(err)
	}
	m, err := migrate.NewWithDatabaseInstance("file://../migrations", "sqlite", driver)
	if err != nil {
		panic(err)
	}
	if err := m.Up(); err != nil {
		panic(err)
	}

	TestDB = db

	ServerStruct = server.NewServer(db, nil)
	TestServer = httptest.NewServer(ServerStruct)
	defer TestServer.Close()
	// Disable authorization for all requests
	server.NoAuth = true

	code := t.Run()
	os.Exit(code)
}

func TestPing(t *testing.T) {
	defer EnableAuthForThisTest()()

	res, err := http.Get(TestServer.URL + "/ping")
	if err != nil {
		t.Error(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %v", res.StatusCode)
	}
}

func TestPingAuthorized(t *testing.T) {
	defer EnableAuthForThisTest()()

	// Create an authentication token
	tok := token.New(nil)
	if err := tok.Write(TestDB.DB); err != nil {
		t.Errorf("error writing token to database: %v", err)
	}

	req, _ := http.NewRequest("GET", TestServer.URL+"/ping", nil)
	req.Header.Set("Authorization", "Bearer "+tok.Token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("expect status 200, got %v", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Errorf("failed to read body: %v", err)
	}
	if string(body) != "pong" {
		t.Errorf("expected 'pong' as response, got '%v'", body)
	}
}
