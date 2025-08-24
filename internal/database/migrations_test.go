package database

import (
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

