package database

import (
	"database/sql"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "modernc.org/sqlite"
)

// NewWithMigrations creates the database at the given path, and runs the migrations using "file" driver
func NewWithMigrations(dbPath string, migrationsPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	driver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"sqlite", driver,
	)
	if err != nil {
		return nil, err
	}

	// Run database migrations
	if err := m.Up(); err != nil && err.Error() != "no change" {
		return nil, err
	}

	return db, nil
}
