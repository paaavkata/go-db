package godb

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	DefaultMaxConns        = 10
	DefaultMinConns        = 2
	DefaultMaxConnLifetime = time.Hour
	DefaultMaxConnIdleTime = 30 * time.Minute
)

// DBConfig holds the configuration for database connections
type DBConfig struct {
	MaxConns        int
	MinConns        int
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

// DefaultConfig returns the default database configuration
func DefaultConfig() *DBConfig {
	return &DBConfig{
		MaxConns:        DefaultMaxConns,
		MinConns:        DefaultMinConns,
		MaxConnLifetime: DefaultMaxConnLifetime,
		MaxConnIdleTime: DefaultMaxConnIdleTime,
	}
}

// DBService provides database connection management
type DBService struct {
	pool *pgxpool.Pool
}

// NewDBService creates a new database service with the given configuration.
// With a nil config it uses DefaultConfig, overridden by the DB_MAX_CONNS and
// DB_MIN_CONNS environment variables when they are set; an unparseable value
// is an error. An explicit config is used as is.
func NewDBService(dbUri string, config *DBConfig) (*DBService, error) {
	if config == nil {
		config = DefaultConfig()
		for env, dst := range map[string]*int{"DB_MAX_CONNS": &config.MaxConns, "DB_MIN_CONNS": &config.MinConns} {
			v := os.Getenv(env)
			if v == "" {
				continue
			}
			n, err := strconv.Atoi(v)
			if err != nil || n < 0 {
				return nil, fmt.Errorf("invalid %s %q: must be a non-negative integer", env, v)
			}
			*dst = n
		}
	}

	poolConfig, err := pgxpool.ParseConfig(dbUri)
	if err != nil {
		return nil, err
	}

	poolConfig.MaxConns = int32(config.MaxConns)
	poolConfig.MinConns = int32(config.MinConns)
	poolConfig.MaxConnLifetime = config.MaxConnLifetime
	poolConfig.MaxConnIdleTime = config.MaxConnIdleTime

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, err
	}

	return &DBService{pool: pool}, nil
}

// Close closes the database connection pool
func (db *DBService) Close() {
	if db.pool != nil {
		db.pool.Close()
	}
}

// Pool returns the underlying connection pool
func (db *DBService) Pool() *pgxpool.Pool {
	return db.pool
}

// Exec executes a query without returning any rows
func (db *DBService) Exec(ctx context.Context, query string, args ...interface{}) error {
	_, err := db.pool.Exec(ctx, query, args...)
	return err
}

// Query executes a query that returns rows
func (db *DBService) Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	return db.pool.Query(ctx, query, args...)
}

// QueryRow executes a query that returns a single row
func (db *DBService) QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	return db.pool.QueryRow(ctx, query, args...)
}

// Begin starts a new transaction
func (db *DBService) Begin(ctx context.Context) (pgx.Tx, error) {
	return db.pool.Begin(ctx)
}

// Ping checks the database connection
func (db *DBService) Ping(ctx context.Context) error {
	return db.pool.Ping(ctx)
}
