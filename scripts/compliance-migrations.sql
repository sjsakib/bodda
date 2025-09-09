-- Run these migrations to add compliance tracking tables

-- User consent tracking table
CREATE TABLE IF NOT EXISTS user_consents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    consent_type VARCHAR(50) NOT NULL,
    granted BOOLEAN NOT NULL DEFAULT false,
    granted_at TIMESTAMP WITH TIME ZONE,
    revoked_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Ensure only one consent record per user per type
    UNIQUE(user_id, consent_type)
);

-- Index for efficient consent lookups
CREATE INDEX IF NOT EXISTS idx_user_consents_user_type ON user_consents(user_id, consent_type);
CREATE INDEX IF NOT EXISTS idx_user_consents_granted ON user_consents(granted, consent_type);

-- Compliance audit log table
CREATE TABLE IF NOT EXISTS compliance_audit (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL,
    details JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for audit log queries
CREATE INDEX IF NOT EXISTS idx_compliance_audit_user_id ON compliance_audit(user_id);
CREATE INDEX IF NOT EXISTS idx_compliance_audit_action ON compliance_audit(action);
CREATE INDEX IF NOT EXISTS idx_compliance_audit_created_at ON compliance_audit(created_at DESC);

-- Data retention tracking table
CREATE TABLE IF NOT EXISTS data_retention (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    data_type VARCHAR(50) NOT NULL,
    last_accessed TIMESTAMP WITH TIME ZONE,
    retention_until TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Ensure only one retention record per user per data type
    UNIQUE(user_id, data_type)
);

-- Indexes for retention queries
CREATE INDEX IF NOT EXISTS idx_data_retention_user_id ON data_retention(user_id);
CREATE INDEX IF NOT EXISTS idx_data_retention_type ON data_retention(data_type);
CREATE INDEX IF NOT EXISTS idx_data_retention_expired ON data_retention(retention_until) WHERE retention_until < NOW();

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers to automatically update updated_at
CREATE TRIGGER update_user_consents_updated_at 
    BEFORE UPDATE ON user_consents 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_data_retention_updated_at 
    BEFORE UPDATE ON data_retention 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert default consent types for existing users (if any)
-- This ensures existing users have proper consent records
INSERT INTO user_consents (user_id, consent_type, granted, granted_at)
SELECT 
    id as user_id,
    'data_processing' as consent_type,
    true as granted,
    created_at as granted_at
FROM users
WHERE NOT EXISTS (
    SELECT 1 FROM user_consents 
    WHERE user_consents.user_id = users.id 
    AND user_consents.consent_type = 'data_processing'
);

INSERT INTO user_consents (user_id, consent_type, granted, granted_at)
SELECT 
    id as user_id,
    'strava_access' as consent_type,
    CASE WHEN access_token IS NOT NULL AND access_token != '' THEN true ELSE false END as granted,
    CASE WHEN access_token IS NOT NULL AND access_token != '' THEN created_at ELSE NULL END as granted_at
FROM users
WHERE NOT EXISTS (
    SELECT 1 FROM user_consents 
    WHERE user_consents.user_id = users.id 
    AND user_consents.consent_type = 'strava_access'
);

-- Create initial data retention records for existing users
INSERT INTO data_retention (user_id, data_type, last_accessed, retention_until)
SELECT 
    id as user_id,
    'user_profile' as data_type,
    NOW() as last_accessed,
    NOW() + INTERVAL '5 years' as retention_until
FROM users
WHERE NOT EXISTS (
    SELECT 1 FROM data_retention 
    WHERE data_retention.user_id = users.id 
    AND data_retention.data_type = 'user_profile'
);

-- Add compliance-related columns to users table if they don't exist
ALTER TABLE users 
ADD COLUMN IF NOT EXISTS privacy_settings JSONB DEFAULT '{}',
ADD COLUMN IF NOT EXISTS data_processing_consent_at TIMESTAMP WITH TIME ZONE,
ADD COLUMN IF NOT EXISTS marketing_consent_at TIMESTAMP WITH TIME ZONE;

-- Create view for active consents (for easy querying)
CREATE OR REPLACE VIEW active_user_consents AS
SELECT 
    user_id,
    consent_type,
    granted_at,
    created_at
FROM user_consents
WHERE granted = true AND revoked_at IS NULL;

-- Create view for compliance dashboard
CREATE OR REPLACE VIEW compliance_summary AS
SELECT 
    u.id as user_id,
    u.email,
    u.created_at as user_created_at,
    COUNT(DISTINCT uc.consent_type) FILTER (WHERE uc.granted = true AND uc.revoked_at IS NULL) as active_consents,
    COUNT(DISTINCT ca.id) as audit_entries,
    COUNT(DISTINCT dr.data_type) as tracked_data_types,
    MAX(ca.created_at) as last_activity
FROM users u
LEFT JOIN user_consents uc ON u.id = uc.user_id
LEFT JOIN compliance_audit ca ON u.id = ca.user_id
LEFT JOIN data_retention dr ON u.id = dr.user_id
GROUP BY u.id, u.email, u.created_at;

-- Grant appropriate permissions (adjust as needed for your setup)
-- GRANT SELECT, INSERT, UPDATE, DELETE ON user_consents TO your_app_user;
-- GRANT SELECT, INSERT ON compliance_audit TO your_app_user;
-- GRANT SELECT, INSERT, UPDATE, DELETE ON data_retention TO your_app_user;
-- GRANT SELECT ON active_user_consents TO your_app_user;
-- GRANT SELECT ON compliance_summary TO your_app_user;