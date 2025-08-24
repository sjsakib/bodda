# Database Package

This package contains the database models, migrations, and repository implementations for the Bodda application.

## Structure

- `connection.go` - Database connection utilities
- `migrations.go` - Database schema migrations
- `repository.go` - Main repository coordinator
- `*_repository.go` - Individual repository implementations for each model
- `*_test.go` - Unit tests for repositories and models
- `test_db.go` - Test database utilities

## Models

The package implements the following models:

- **User** - Strava authenticated users with tokens and profile information
- **Session** - Chat conversation sessions belonging to users
- **Message** - Individual messages within sessions (user or assistant)
- **AthleteLogbook** - Persistent athlete profile and training insights

## Repository Pattern

Each model has a corresponding repository that provides CRUD operations:

- `UserRepository` - User management and Strava ID lookups
- `SessionRepository` - Session management and user session retrieval
- `MessageRepository` - Message persistence and conversation history
- `LogbookRepository` - Athlete logbook management with upsert functionality

## Running Tests

### Unit Tests (No Database Required)

```bash
# Run model tests
go test ./internal/models/... -v

# Run migration syntax tests
go test ./internal/database/ -run TestMigration -v
```

### Integration Tests (Database Required)

To run the full repository tests, you need a PostgreSQL test database:

1. **Using Docker:**
```bash
docker run --name bodda-test-db -e POSTGRES_PASSWORD=password -e POSTGRES_DB=bodda_test -p 5432:5432 -d postgres:15
```

2. **Set environment variable:**
```bash
export TEST_DATABASE_URL="postgres://postgres:password@localhost:5432/bodda_test?sslmode=disable"
```

3. **Run tests:**
```bash
go test ./internal/database/... -v
```

### Test Database Cleanup

The test suite automatically:
- Creates the test database schema using migrations
- Cleans all tables between tests
- Skips tests gracefully if no database is available

## Usage Example

```go
package main

import (
    "context"
    "bodda/internal/database"
    "bodda/internal/models"
)

func main() {
    // Connect to database
    db, err := database.Connect("postgres://...")
    if err != nil {
        panic(err)
    }
    defer db.Close()

    // Run migrations
    if err := database.RunMigrations(db); err != nil {
        panic(err)
    }

    // Create repository
    repo := database.NewRepository(db)

    // Use repositories
    user := &models.User{
        StravaID: 12345,
        // ... other fields
    }
    
    err = repo.User.Create(context.Background(), user)
    if err != nil {
        panic(err)
    }
}
```

## Database Schema

The schema includes proper foreign key relationships and constraints:

- Users have unique Strava IDs
- Sessions belong to users (CASCADE DELETE)
- Messages belong to sessions (CASCADE DELETE)
- Athlete logbooks have unique user associations
- Message roles are constrained to 'user' or 'assistant'

All tables use UUID primary keys and include appropriate timestamps.