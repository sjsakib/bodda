-- Simple audit table for compliance tracking
-- Much simpler than the full compliance system

CREATE TABLE IF NOT EXISTS audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    action VARCHAR(100) NOT NULL,
    details JSONB,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for efficient audit queries
CREATE INDEX IF NOT EXISTS idx_audit_log_user_id ON audit_log(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_timestamp ON audit_log(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_audit_log_action ON audit_log(action);

-- Insert some initial audit entries for existing users (optional)
INSERT INTO audit_log (user_id, action, details)
SELECT 
    id as user_id,
    'account_created' as action,
    json_build_object('created_at', created_at) as details
FROM users
WHERE NOT EXISTS (
    SELECT 1 FROM audit_log 
    WHERE audit_log.user_id = users.id 
    AND audit_log.action = 'account_created'
);

-- Add Strava connection audit entries for users who already have tokens
INSERT INTO audit_log (user_id, action, details)
SELECT 
    id as user_id,
    'strava_connected' as action,
    json_build_object('connected_at', created_at) as details
FROM users
WHERE access_token IS NOT NULL 
AND access_token != ''
AND NOT EXISTS (
    SELECT 1 FROM audit_log 
    WHERE audit_log.user_id = users.id 
    AND audit_log.action = 'strava_connected'
);