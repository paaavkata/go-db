package godb

import (
	"context"
	"embed"
	"os"
	"sync"
	"testing"
)

//go:embed testdata/migrations
var testMigrations embed.FS

// Pool sizing comes from DB_MAX_CONNS / DB_MIN_CONNS when the config is nil.
// pgxpool.NewWithConfig connects lazily, so no server is needed.
func TestNewDBServiceEnvPoolSize(t *testing.T) {
	t.Setenv("DB_MAX_CONNS", "4")
	t.Setenv("DB_MIN_CONNS", "1")
	db, err := NewDBService("postgres://u:p@127.0.0.1:1/db", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if got := db.Pool().Config().MaxConns; got != 4 {
		t.Fatalf("MaxConns = %d, want 4", got)
	}
	if got := db.Pool().Config().MinConns; got != 1 {
		t.Fatalf("MinConns = %d, want 1", got)
	}

	t.Setenv("DB_MAX_CONNS", "four")
	if _, err := NewDBService("postgres://u:p@127.0.0.1:1/db", nil); err == nil {
		t.Fatal("want error for DB_MAX_CONNS=four")
	}
}

func TestNewDBServiceDefaults(t *testing.T) {
	db, err := NewDBService("postgres://u:p@127.0.0.1:1/db", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if c := db.Pool().Config(); c.MaxConns != DefaultMaxConns || c.MinConns != DefaultMinConns {
		t.Fatalf("defaults not applied: max %d min %d", c.MaxConns, c.MinConns)
	}
}

// Two concurrent Migrate runs against a real database both succeed: one waits
// on the advisory lock, then re-runs the idempotent files.
func TestMigrateConcurrent(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if testing.Short() || url == "" {
		t.Skip("needs TEST_DATABASE_URL and no -short")
	}
	db, err := NewDBService(url, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	defer db.Exec(ctx, "DROP TABLE IF EXISTS godb_migrate_test")

	var wg sync.WaitGroup
	errs := make([]error, 2)
	for i := range errs {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			errs[i] = db.Migrate(ctx, testMigrations, "testdata/migrations", 0x676f64625f746573)
		}(i)
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Fatalf("run %d: %v", i, err)
		}
	}
	var n int
	if err := db.QueryRow(ctx, "SELECT count(*) FROM godb_migrate_test").Scan(&n); err != nil || n != 1 {
		t.Fatalf("rows = %d, err %v", n, err)
	}
}
