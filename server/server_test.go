package server_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/3elDU/rss-reader-backend/database"
	"github.com/3elDU/rss-reader-backend/middleware"
	"github.com/3elDU/rss-reader-backend/server"
	"github.com/3elDU/rss-reader-backend/token"
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
	middleware.NoAuth = false
	return func() {
		middleware.NoAuth = true
	}
}

func TestMain(t *testing.M) {
	// Create an in-memory DB
	godb, err := database.NewWithMigrations(":memory:", "../database/migrations")
	if err != nil {
		panic(err)
	}
	db := sqlx.NewDb(godb, "sqlite")

	TestDB = db

	ServerStruct = server.NewServer(db, nil)
	TestServer = httptest.NewServer(ServerStruct)
	defer TestServer.Close()
	// Disable authorization for all requests
	middleware.NoAuth = true

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
	tok := token.New(nil).ToModel()
	tokrepo := database.NewTokenRepository(TestDB)
	if err := tokrepo.Insert(&tok); err != nil {
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
