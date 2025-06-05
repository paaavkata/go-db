# Go Database Library

A shared library for database connections and operations across FileConvert microservices.

## Features

- PostgreSQL connection management
- Connection pooling configuration
- Database migration utilities
- Common database operations
- Schema management

## Usage

```go
import "github.com/your-org/go-db"

// Create a new database connection
db, err := godb.NewDBService(dbUri)
if err != nil {
    log.Fatal(err)
}
defer db.Close()

// Use the connection
// ...
```

## Configuration

The library supports the following configuration options:

- Connection pooling settings
- Schema management
- Migration handling
- Error handling strategies 