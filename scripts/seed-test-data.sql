-- Test seed data for Bodda application
-- This script populates the test database with minimal data for testing

-- Connect to the test database
\c bodda_test;

-- Insert minimal test users
INSERT INTO users (id, strava_id, access_token, refresh_token, token_expiry, first_name, last_name, created_at, updated_at)
VALUES 
    ('test-user-1', 999999, 'test_access_token', 'test_refresh_token', NOW() + INTERVAL '1 hour', 'Test', 'User', NOW(), NOW()),
    ('test-user-2', 888888, 'test_access_token_2', 'test_refresh_token_2', NOW() + INTERVAL '1 hour', 'Another', 'Tester', NOW(), NOW())
ON CONFLICT (strava_id) DO NOTHING;

-- Insert test session
INSERT INTO sessions (id, user_id, title, created_at, updated_at)
VALUES 
    ('test-session-1', 'test-user-1', 'Test Session', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Insert test messages
INSERT INTO messages (id, session_id, role, content, created_at)
VALUES 
    ('test-message-1', 'test-session-1', 'user', 'Test user message', NOW()),
    ('test-message-2', 'test-session-1', 'assistant', 'Test assistant response', NOW())
ON CONFLICT (id) DO NOTHING;

-- Insert test logbook
INSERT INTO athlete_logbooks (id, user_id, content, updated_at)
VALUES 
    ('test-logbook-1', 'test-user-1', '{"profile": {"name": "Test User", "goals": ["testing"]}}', NOW())
ON CONFLICT (user_id) DO NOTHING;

-- Display test data summary
DO $$
BEGIN
    RAISE NOTICE 'Test data seeded successfully:';
    RAISE NOTICE '- % users', (SELECT COUNT(*) FROM users);
    RAISE NOTICE '- % sessions', (SELECT COUNT(*) FROM sessions);
    RAISE NOTICE '- % messages', (SELECT COUNT(*) FROM messages);
    RAISE NOTICE '- % athlete logbooks', (SELECT COUNT(*) FROM athlete_logbooks);
END $$;