package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func RunMigrations(db *pgxpool.Pool) error {
	migrations := []string{
		createUsersTable,
		createSessionsTable,
		createMessagesTable,
		createAthleteLogbooksTable,
		addResponseIdToMessages,
		addLastResponseIdToSessions,
	}

	for i, migration := range migrations {
		if _, err := db.Exec(context.Background(), migration); err != nil {
			return fmt.Errorf("failed to run migration %d: %w", i+1, err)
		}
	}

	return nil
}

const createUsersTable = `
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    strava_id BIGINT UNIQUE NOT NULL,
    access_token TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    token_expiry TIMESTAMP NOT NULL,
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);`

const createSessionsTable = `
CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);`

const createMessagesTable = `
CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID REFERENCES sessions(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL CHECK (role IN ('user', 'assistant')),
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);`

const createAthleteLogbooksTable = `
CREATE TABLE IF NOT EXISTS athlete_logbooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE UNIQUE,
    content TEXT,
    updated_at TIMESTAMP DEFAULT NOW()
);`

const addResponseIdToMessages = `
ALTER TABLE messages 
ADD COLUMN IF NOT EXISTS response_id TEXT;`

const addLastResponseIdToSessions = `
ALTER TABLE sessions 
ADD COLUMN IF NOT EXISTS last_response_id TEXT;`