package postgres

import (
	"context"
	"os"
	"testing"
)

func openTestDB(t *testing.T) *DB {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL not set: romanctl up --no-run creates the database and the .env")
	}
	db, err := Open(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(db.Close)
	return db
}
