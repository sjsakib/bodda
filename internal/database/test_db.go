package database

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TestDB struct {
	Pool *pgxpool.Pool
	t    *testing.T
}

func NewTestDB(t *testing.T) *TestDB {
	// Use environment variable or default to test database
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:password@localhost:5432/bodda_test?sslmode=disable"
	}

	pool, err := Connect(databaseURL)
	if err != nil {
		t.Skipf("Skipping database tests: %v", err)
	}

	testDB := &TestDB{
		Pool: pool,
		t:    t,
	}

	// Run migrations
	if err := RunMigrations(pool); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return testDB
}

func (db *TestDB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}

func (db *TestDB) CleanTables() {
	tables := []string{
		"messages",
		"sessions", 
		"athlete_logbooks",
		"users",
	}

	for _, table := range tables {
		query := fmt.Sprintf("DELETE FROM %s", table)
		_, err := db.Pool.Exec(context.Background(), query)
		if err != nil {
			db.t.Fatalf("Failed to clean table %s: %v", table, err)
		}
	}
}