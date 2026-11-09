package godb

import (
	"context"
	"fmt"
	"io/fs"
	"path"
)

// Migrate runs every regular file in dir of fsys, in lexical order, inside one
// transaction that first sets a 10 minute statement timeout and takes the
// transaction-level advisory lock lockKey. Concurrent replicas therefore run
// the files one after the other, never at the same time; every file must be
// idempotent. Transaction-level lock and SET LOCAL also survive pgBouncer
// transaction pooling. On a failing file the transaction rolls back, the lock
// is released and the error names the file. Pick lockKey unique per service
// among the lock keys used on the same database.
func (db *DBService) Migrate(ctx context.Context, fsys fs.FS, dir string, lockKey int64) error {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return fmt.Errorf("migrate: read %s: %w", dir, err)
	}

	conn, err := db.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("migrate: acquire conn: %w", err)
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("migrate: begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, "SET LOCAL statement_timeout = '10min'"); err != nil {
		return fmt.Errorf("migrate: set statement_timeout: %w", err)
	}
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", lockKey); err != nil {
		return fmt.Errorf("migrate: acquire advisory lock: %w", err)
	}

	for _, entry := range entries { // fs.ReadDir returns entries sorted by name
		if entry.IsDir() {
			continue
		}
		body, err := fs.ReadFile(fsys, path.Join(dir, entry.Name()))
		if err != nil {
			return fmt.Errorf("migrate: read %s: %w", entry.Name(), err)
		}
		if _, err := tx.Exec(ctx, string(body)); err != nil {
			return fmt.Errorf("migrate %s: %w", entry.Name(), err)
		}
	}
	return tx.Commit(ctx)
}
