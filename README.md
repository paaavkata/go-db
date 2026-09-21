# Go Database Library

A shared library for PostgreSQL connections across FileConvert microservices. Thin wrapper around `jackc/pgx/v5` (`pgxpool`).

## Features

- PostgreSQL connection pooling via `pgxpool`
- Configurable pool settings (max/min conns, conn lifetime, idle time), with sane defaults
- Common database operations: `Exec`, `Query`, `QueryRow`, `Begin`, `Ping`

There are no migration utilities or schema-management helpers in this package — schema/migration handling is left to each service.

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
- `NewDBService(dbUri string, config *DBConfig) (*DBService, error)` — `config` may be `nil` to use `DefaultConfig()`
- `(*DBService) Pool() *pgxpool.Pool`
- `(*DBService) Exec(ctx, query string, args ...interface{}) error`
- `(*DBService) Query(ctx, query string, args ...interface{}) (pgx.Rows, error)`
- `(*DBService) QueryRow(ctx, query string, args ...interface{}) pgx.Row`
- `(*DBService) Begin(ctx) (pgx.Tx, error)`
- `(*DBService) Ping(ctx) error`
- `(*DBService) Close()`
