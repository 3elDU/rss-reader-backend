package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/3elDU/rss-reader-backend/server"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/jmoiron/sqlx"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "modernc.org/sqlite"
)

func runMigrations() {
	db, err := sql.Open("sqlite", "database.sqlite")
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	driver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		log.Fatalf("failed to create database driver for the migrations: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"sqlite", driver,
	)
	if err != nil {
		log.Fatalf("failed to create migrate instance: %v", err)
	}

	// Run database migrations
	if err := m.Up(); err != nil && err.Error() != "no change" {
		log.Fatalf("failed to apply database migrations: %v", err)
	}
}

// Generate authentication token, add it to the database and print it
func createAuthToken(db *sqlx.DB, validFor time.Duration) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}

	token := base64.URLEncoding.EncodeToString(buf)

	now := time.Now().UTC()
	createdAt := now.Format(time.DateTime)
	var validUntil interface{} = nil // The default is NULL
	if validFor != 0 {
		validUntil = now.Add(validFor).Format(time.DateTime)
	}

	db.MustExec(
		"INSERT INTO auth_tokens (token, created_at, valid_until) VALUES ($1, $2, $3)",
		token, createdAt, validUntil,
	)

	fmt.Println(token)
}

func main() {
	runMigrations()

	// Instantiate the database
	db := sqlx.MustConnect("sqlite", "database.sqlite")
	defer db.Close()

	createToken := flag.Bool("createtoken", false, "Create a new authentication token, output it and exit.")

	validFor := flag.Duration("validfor", time.Duration(0), "Used with 'createToken'. The duration for which the token will be valid. The default is no expiration.")

	listenAddr := flag.String("listen", "127.0.0.1:8080", "Address to listen on, with port.")

	flag.Parse()

	if *createToken {
		createAuthToken(db, *validFor)
		return
	}

	log.Printf("Running the web server on %v", *listenAddr)

	server := server.NewServer(db)
	log.Fatal(http.ListenAndServe(*listenAddr, server))
}
