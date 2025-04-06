package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/3elDU/rss-reader-backend/database"
	"github.com/3elDU/rss-reader-backend/middleware"
	"github.com/3elDU/rss-reader-backend/refresh"
	"github.com/3elDU/rss-reader-backend/server"
	"github.com/3elDU/rss-reader-backend/token"
	"github.com/jmoiron/sqlx"
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
	databasePath = flag.String(
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

func main() {
	// Show file and line number where the log originated
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flag.BoolVar(
		&middleware.NoAuth,
		"noauth",
		false,
		"Disable authentication entirely. Useful for debugging.",
	)

	flag.Parse()
	if middleware.NoAuth && !*createToken {
		log.Printf("*** RUNNING WITH AUTHENTICATION DISABLED ***")
	}

	dbOrig, err := database.NewWithMigrations(*databasePath, "database/migrations")
	if err != nil {
		log.Fatalf("failed to apply database migrations: %v", err)
	}

	// Instantiate the database
	db := sqlx.NewDb(dbOrig, "sqlite")
	if !*createToken {
		log.Printf("Connected to the database in '%v'", *databasePath)
	}
	defer db.Close()

	if *createToken {
		t := token.New(validFor).ToModel()
		repo := database.NewTokenRepository(db)
		if err := repo.Insert(&t); err != nil {
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
