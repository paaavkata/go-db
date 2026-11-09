# Go Database Library

A shared library for PostgreSQL connections across FileConvert microservices. Thin wrapper around `jackc/pgx/v5` (`pgxpool`).

## Features

- PostgreSQL connection pooling via `pgxpool`
- Configurable pool settings (max/min conns, conn lifetime, idle time), with sane defaults
- Common database operations: `Exec`, `Query`, `QueryRow`, `Begin`, `Ping`
- `Migrate`: runs a directory of idempotent SQL files under a transaction-level advisory lock (v0.2.0)

## Pool sizing by environment

With a `nil` config, `NewDBService` starts from `DefaultConfig()` and applies `DB_MAX_CONNS` and
`DB_MIN_CONNS` when they are set (non-negative integers; anything else fails `NewDBService` with an
error naming the variable). An explicit config ignores both variables.

## Migrations

```go
//go:embed sql/*.sql
var sqlFiles embed.FS

const migrationLockKey int64 = 0x7363616e5f6d6701 // unique per service on this database

err := db.Migrate(ctx, sqlFiles, "sql", migrationLockKey)
```

`Migrate` reads every regular file in the directory in lexical order and runs them in one
transaction that first does `SET LOCAL statement_timeout = '10min'` and
`SELECT pg_advisory_xact_lock(lockKey)`. Two replicas starting together run the files one after
the other; the second sees idempotent no-ops. A failing file rolls back the whole run, releases the
lock and returns an error naming the file. There is no version table: every file must be safe to
re-run. `TestMigrateConcurrent` exercises this against `TEST_DATABASE_URL` (skipped under `-short`).

## Usage

```go
import godb "github.com/paaavkata/go-db"

// Optional: customize pool settings, or pass nil to use DefaultConfig()
config := godb.DefaultConfig()
config.MaxConns = 20

// Create a new database service
db, err := godb.NewDBService(dbUri, config)
if err != nil {
    log.Fatal(err)
}
defer db.Close()

// Use the connection
rows, err := db.Query(ctx, "SELECT id FROM users WHERE app_id = $1", appID)
```

## API surface

- `DefaultConfig() *DBConfig` — `MaxConns=10`, `MinConns=2`, `MaxConnLifetime=1h`, `MaxConnIdleTime=30m`
- `NewDBService(dbUri string, config *DBConfig) (*DBService, error)` — `config` may be `nil` to use `DefaultConfig()` plus the `DB_MAX_CONNS` / `DB_MIN_CONNS` overrides
- `(*DBService) Migrate(ctx, fsys fs.FS, dir string, lockKey int64) error`
- `(*DBService) Pool() *pgxpool.Pool`
- `(*DBService) Exec(ctx, query string, args ...interface{}) error`
- `(*DBService) Query(ctx, query string, args ...interface{}) (pgx.Rows, error)`
- `(*DBService) QueryRow(ctx, query string, args ...interface{}) pgx.Row`
- `(*DBService) Begin(ctx) (pgx.Tx, error)`
- `(*DBService) Ping(ctx) error`
- `(*DBService) Close()`
