package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/3elDU/rss-reader-backend/refresh"
	"github.com/3elDU/rss-reader-backend/server"
	"github.com/3elDU/rss-reader-backend/token"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/jmoiron/sqlx"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "modernc.org/sqlite"
)

var (
	createToken = flag.Bool(
		"createtoken",
		false,
		"Create a new authentication token, output it and exit.",
	)
	listenAddr = flag.String(
		"listen",
		"[::1]:8080",
		"Address to listen on, with port.",
	)
	database = flag.String(
		"db",
		"database.sqlite",
		"Path to the database file to use.",
	)
	validFor = flag.Duration(
		"validfor",
		time.Duration(0),
		"Used with 'createToken'. The duration for which the token will be valid. The default is no expiration.",
	)
	refreshFreq = flag.Duration(
		"refresh",
		time.Minute*15,
		"Frequency with which feeds will be updated",
	)
)

func runMigrations() {
	db, err := sql.Open("sqlite", *database)
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

func main() {
	// Show file and line number where the log originated
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flag.BoolVar(
		&server.NoAuth,
		"noauth",
		false,
		"Disable authentication entirely. Useful for debugging.",
	)

	flag.Parse()
	if server.NoAuth && !*createToken {
		log.Printf("*** RUNNING WITH AUTHENTICATION DISABLED ***")
	}

	runMigrations()

	// Instantiate the database
	db := sqlx.MustConnect("sqlite", *database)
	if !*createToken {
		log.Printf("Connected to the database in '%v'", *database)
	}
	defer db.Close()

	if *createToken {
		// If the duration is unset, make it nil
		var d *time.Duration
		if *validFor != 0 {
			d = validFor
		}

		t := token.New(d)
		if err := t.Write(db.DB); err != nil {
			panic(err)
		}
		fmt.Println(t.Token)

		return
	}

	log.Printf("Running the web server on %v", *listenAddr)

	task := refresh.NewTask(db, *refreshFreq)
	server := server.NewServer(db, task)

	go runServer(server)
	task.Run()
}

func runServer(server *server.Server) {
	err := http.ListenAndServe(*listenAddr, server)
	if err != nil {
		panic(err)
	}
}
