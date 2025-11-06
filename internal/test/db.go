// Package test provides testing utilities for Kaunta
package test

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/peterldowns/pgtestdb"
	"github.com/peterldowns/pgtestdb/migrators/golangmigrator"
)

// TestDB holds database connection for tests
type TestDB struct {
	DB *sql.DB
}

// NewTestDB creates a fresh test database with migrations applied
func NewTestDB(t *testing.T) *TestDB {
	t.Helper()

	// Find the project root by looking for migrations directory
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	var migrationsPath string

	// Walk up directories to find migrations
	currentPath := wd
	for {
		testPath := filepath.Join(currentPath, "internal", "database", "migrations")
		if _, err := os.Stat(testPath); err == nil {
			migrationsPath = testPath
			break
		}
		parent := filepath.Dir(currentPath)
		if parent == currentPath {
			t.Fatalf("could not find migrations directory")
		}
		currentPath = parent
	}

	// Create isolated test database using template cloning
	// This is much faster than running migrations for each test (~20ms per test)
	db := pgtestdb.New(t, pgtestdb.Config{
		DriverName: "pgx",
		Host:       "localhost",
		User:       "postgres",
		Password:   "postgres",
		Port:       "5432",
		Options:    "sslmode=disable",
	}, golangmigrator.New(migrationsPath))

	return &TestDB{
		DB: db,
	}
}

// Close closes the database connection
func (tdb *TestDB) Close() error {
	if tdb.DB != nil {
		return tdb.DB.Close()
	}
	return nil
}

// Exec executes a raw SQL query for test setup/teardown
func (tdb *TestDB) Exec(ctx context.Context, query string, args ...interface{}) error {
	_, err := tdb.DB.ExecContext(ctx, query, args...)
	return err
}

// QueryRow executes a query returning a single row
func (tdb *TestDB) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return tdb.DB.QueryRowContext(ctx, query, args...)
}

// Query executes a query returning multiple rows
func (tdb *TestDB) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return tdb.DB.QueryContext(ctx, query, args...)
}
