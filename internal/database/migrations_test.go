package database

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMigrationQueries(t *testing.T) {
	// Test that migration queries are valid SQL and contain expected elements
	
	t.Run("Users table migration", func(t *testing.T) {
		assert.Contains(t, createUsersTable, "CREATE TABLE IF NOT EXISTS users")
		assert.Contains(t, createUsersTable, "id UUID PRIMARY KEY DEFAULT gen_random_uuid()")
		assert.Contains(t, createUsersTable, "strava_id BIGINT UNIQUE NOT NULL")
		assert.Contains(t, createUsersTable, "access_token TEXT NOT NULL")
		assert.Contains(t, createUsersTable, "refresh_token TEXT NOT NULL")
		assert.Contains(t, createUsersTable, "token_expiry TIMESTAMP NOT NULL")
		assert.Contains(t, createUsersTable, "first_name VARCHAR(255)")
		assert.Contains(t, createUsersTable, "last_name VARCHAR(255)")
		assert.Contains(t, createUsersTable, "created_at TIMESTAMP DEFAULT NOW()")
		assert.Contains(t, createUsersTable, "updated_at TIMESTAMP DEFAULT NOW()")
	})

	t.Run("Sessions table migration", func(t *testing.T) {
		assert.Contains(t, createSessionsTable, "CREATE TABLE IF NOT EXISTS sessions")
		assert.Contains(t, createSessionsTable, "id UUID PRIMARY KEY DEFAULT gen_random_uuid()")
		assert.Contains(t, createSessionsTable, "user_id UUID REFERENCES users(id) ON DELETE CASCADE")
		assert.Contains(t, createSessionsTable, "title VARCHAR(255) NOT NULL")
		assert.Contains(t, createSessionsTable, "created_at TIMESTAMP DEFAULT NOW()")
		assert.Contains(t, createSessionsTable, "updated_at TIMESTAMP DEFAULT NOW()")
	})

	t.Run("Messages table migration", func(t *testing.T) {
		assert.Contains(t, createMessagesTable, "CREATE TABLE IF NOT EXISTS messages")
		assert.Contains(t, createMessagesTable, "id UUID PRIMARY KEY DEFAULT gen_random_uuid()")
		assert.Contains(t, createMessagesTable, "session_id UUID REFERENCES sessions(id) ON DELETE CASCADE")
		assert.Contains(t, createMessagesTable, "role VARCHAR(20) NOT NULL CHECK (role IN ('user', 'assistant'))")
		assert.Contains(t, createMessagesTable, "content TEXT NOT NULL")
		assert.Contains(t, createMessagesTable, "created_at TIMESTAMP DEFAULT NOW()")
	})

	t.Run("Athlete logbooks table migration", func(t *testing.T) {
		assert.Contains(t, createAthleteLogbooksTable, "CREATE TABLE IF NOT EXISTS athlete_logbooks")
		assert.Contains(t, createAthleteLogbooksTable, "id UUID PRIMARY KEY DEFAULT gen_random_uuid()")
		assert.Contains(t, createAthleteLogbooksTable, "user_id UUID REFERENCES users(id) ON DELETE CASCADE UNIQUE")
		assert.Contains(t, createAthleteLogbooksTable, "content TEXT")
		assert.Contains(t, createAthleteLogbooksTable, "updated_at TIMESTAMP DEFAULT NOW()")
	})

	t.Run("Add response_id to messages migration", func(t *testing.T) {
		assert.Contains(t, addResponseIdToMessages, "ALTER TABLE messages")
		assert.Contains(t, addResponseIdToMessages, "ADD COLUMN IF NOT EXISTS response_id TEXT")
	})

	t.Run("Add last_response_id to sessions migration", func(t *testing.T) {
		assert.Contains(t, addLastResponseIdToSessions, "ALTER TABLE sessions")
		assert.Contains(t, addLastResponseIdToSessions, "ADD COLUMN IF NOT EXISTS last_response_id TEXT")
	})
}

func TestMigrationOrder(t *testing.T) {
	// Test that migrations are in correct dependency order
	// Users must be created before sessions, sessions before messages, etc.
	
	migrations := []string{
		createUsersTable,
		createSessionsTable,
		createMessagesTable,
		createAthleteLogbooksTable,
	}

	// Check that users table comes first (no foreign key dependencies)
	assert.Contains(t, migrations[0], "CREATE TABLE IF NOT EXISTS users")
	assert.NotContains(t, migrations[0], "REFERENCES")

	// Check that sessions table references users
	assert.Contains(t, migrations[1], "CREATE TABLE IF NOT EXISTS sessions")
	assert.Contains(t, migrations[1], "REFERENCES users(id)")

	// Check that messages table references sessions
	assert.Contains(t, migrations[2], "CREATE TABLE IF NOT EXISTS messages")
	assert.Contains(t, migrations[2], "REFERENCES sessions(id)")

	// Check that athlete_logbooks table references users
	assert.Contains(t, migrations[3], "CREATE TABLE IF NOT EXISTS athlete_logbooks")
	assert.Contains(t, migrations[3], "REFERENCES users(id)")
}

func TestMigrationSQLSyntax(t *testing.T) {
	// Basic syntax validation for migration queries
	migrations := []string{
		createUsersTable,
		createSessionsTable,
		createMessagesTable,
		createAthleteLogbooksTable,
	}

	for i, migration := range migrations {
		t.Run(fmt.Sprintf("Migration %d syntax", i+1), func(t *testing.T) {
			// Should start with CREATE TABLE
			assert.True(t, strings.HasPrefix(strings.TrimSpace(migration), "CREATE TABLE"))
			
			// Should end with semicolon or closing parenthesis
			trimmed := strings.TrimSpace(migration)
			assert.True(t, strings.HasSuffix(trimmed, ";") || strings.HasSuffix(trimmed, ");"))
			
			// Should contain proper UUID primary key
			assert.Contains(t, migration, "id UUID PRIMARY KEY DEFAULT gen_random_uuid()")
			
			// Should use IF NOT EXISTS for safe migrations
			assert.Contains(t, migration, "IF NOT EXISTS")
		})
	}
}

func TestLastResponseIdMigration(t *testing.T) {
	// Test the specific last_response_id migration
	t.Run("Migration SQL syntax", func(t *testing.T) {
		// Should be an ALTER TABLE statement
		assert.Contains(t, addLastResponseIdToSessions, "ALTER TABLE sessions")
		
		// Should add the column with proper syntax
		assert.Contains(t, addLastResponseIdToSessions, "ADD COLUMN IF NOT EXISTS last_response_id TEXT")
		
		// Should be nullable (no NOT NULL constraint)
		assert.NotContains(t, addLastResponseIdToSessions, "NOT NULL")
		
		// Should end with semicolon
		assert.True(t, strings.HasSuffix(strings.TrimSpace(addLastResponseIdToSessions), ";"))
	})

	t.Run("Migration order", func(t *testing.T) {
		// The last_response_id migration should come after sessions table creation
		migrations := []string{
			createUsersTable,
			createSessionsTable,
			createMessagesTable,
			createAthleteLogbooksTable,
			addResponseIdToMessages,
			addLastResponseIdToSessions,
		}
		
		// Find the index of sessions table creation and last_response_id addition
		sessionsTableIndex := -1
		lastResponseIdIndex := -1
		
		for i, migration := range migrations {
			if strings.Contains(migration, "CREATE TABLE IF NOT EXISTS sessions") {
				sessionsTableIndex = i
			}
			if strings.Contains(migration, "ADD COLUMN IF NOT EXISTS last_response_id") {
				lastResponseIdIndex = i
			}
		}
		
		// Ensure sessions table is created before adding the column
		assert.True(t, sessionsTableIndex >= 0, "Sessions table creation migration not found")
		assert.True(t, lastResponseIdIndex >= 0, "Last response ID migration not found")
		assert.True(t, sessionsTableIndex < lastResponseIdIndex, "Sessions table must be created before adding last_response_id column")
	})
}

func TestMigrationExecution(t *testing.T) {
	// Integration test that actually runs migrations against a test database
	testDB := NewTestDB(t)
	defer testDB.Close()

	t.Run("Migration execution succeeds", func(t *testing.T) {
		// Migrations are already run in NewTestDB, so we just need to verify
		// that the last_response_id column exists in the sessions table
		
		// Query the information schema to check if the column exists
		var columnExists bool
		query := `
			SELECT EXISTS (
				SELECT 1 
				FROM information_schema.columns 
				WHERE table_name = 'sessions' 
				AND column_name = 'last_response_id'
			)`
		
		err := testDB.Pool.QueryRow(context.Background(), query).Scan(&columnExists)
		assert.NoError(t, err)
		assert.True(t, columnExists, "last_response_id column should exist in sessions table")
	})

	t.Run("Column properties are correct", func(t *testing.T) {
		// Verify the column has the correct properties
		query := `
			SELECT data_type, is_nullable 
			FROM information_schema.columns 
			WHERE table_name = 'sessions' 
			AND column_name = 'last_response_id'`
		
		var dataType, isNullable string
		err := testDB.Pool.QueryRow(context.Background(), query).Scan(&dataType, &isNullable)
		assert.NoError(t, err)
		assert.Equal(t, "text", dataType, "Column should be TEXT type")
		assert.Equal(t, "YES", isNullable, "Column should be nullable")
	})

	t.Run("Backward compatibility with existing sessions", func(t *testing.T) {
		// Clean tables first
		testDB.CleanTables()
		
		// Create a test user first
		userQuery := `
			INSERT INTO users (strava_id, access_token, refresh_token, token_expiry, first_name, last_name)
			VALUES (12345, 'test_access', 'test_refresh', NOW() + INTERVAL '1 hour', 'Test', 'User')
			RETURNING id`
		
		var userID string
		err := testDB.Pool.QueryRow(context.Background(), userQuery).Scan(&userID)
		assert.NoError(t, err)
		
		// Create a session without specifying last_response_id (should default to NULL)
		sessionQuery := `
			INSERT INTO sessions (user_id, title)
			VALUES ($1, 'Test Session')
			RETURNING id, last_response_id`
		
		var sessionID string
		var lastResponseID *string
		err = testDB.Pool.QueryRow(context.Background(), sessionQuery, userID).Scan(&sessionID, &lastResponseID)
		assert.NoError(t, err)
		assert.Nil(t, lastResponseID, "last_response_id should be NULL for new sessions")
		
		// Verify we can update the last_response_id
		updateQuery := `UPDATE sessions SET last_response_id = $1 WHERE id = $2`
		_, err = testDB.Pool.Exec(context.Background(), updateQuery, "test-response-id", sessionID)
		assert.NoError(t, err)
		
		// Verify the update worked
		selectQuery := `SELECT last_response_id FROM sessions WHERE id = $1`
		err = testDB.Pool.QueryRow(context.Background(), selectQuery, sessionID).Scan(&lastResponseID)
		assert.NoError(t, err)
		assert.NotNil(t, lastResponseID)
		assert.Equal(t, "test-response-id", *lastResponseID)
	})
}

