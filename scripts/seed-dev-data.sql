-- Development seed data for Bodda application
-- This script populates the database with sample data for development and testing

-- Connect to the development database
\c bodda_dev;

-- Insert sample users (these would normally be created through Strava OAuth)
INSERT INTO users (id, strava_id, access_token, refresh_token, token_expiry, first_name, last_name, created_at, updated_at)
VALUES 
    ('550e8400-e29b-41d4-a716-446655440001', 12345678, 'dev_access_token_1', 'dev_refresh_token_1', NOW() + INTERVAL '6 hours', 'John', 'Runner', NOW(), NOW()),
    ('550e8400-e29b-41d4-a716-446655440002', 87654321, 'dev_access_token_2', 'dev_refresh_token_2', NOW() + INTERVAL '6 hours', 'Jane', 'Cyclist', NOW(), NOW()),
    ('550e8400-e29b-41d4-a716-446655440003', 11223344, 'dev_access_token_3', 'dev_refresh_token_3', NOW() + INTERVAL '6 hours', 'Mike', 'Triathlete', NOW(), NOW())
ON CONFLICT (strava_id) DO NOTHING;

-- Insert sample sessions
INSERT INTO sessions (id, user_id, title, created_at, updated_at)
VALUES 
    ('660e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440001', 'Training Plan Discussion', NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days'),
    ('660e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440001', 'Recovery Questions', NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day'),
    ('660e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440002', 'Cycling Technique', NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours'),
    ('660e8400-e29b-41d4-a716-446655440004', '550e8400-e29b-41d4-a716-446655440003', 'Triathlon Preparation', NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour')
ON CONFLICT (id) DO NOTHING;

-- Insert sample messages
INSERT INTO messages (id, session_id, role, content, created_at)
VALUES 
    -- Training Plan Discussion
    ('770e8400-e29b-41d4-a716-446655440001', '660e8400-e29b-41d4-a716-446655440001', 'user', 'Hi! I''m looking to improve my 5K time. Can you help me create a training plan?', NOW() - INTERVAL '2 days'),
    ('770e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440001', 'assistant', 'Absolutely! I''d be happy to help you improve your 5K time. Let me first look at your recent running activities to understand your current fitness level and training patterns.

Based on your Strava data, I can see you''ve been running consistently. Here''s a structured approach to improve your 5K:

## 4-Week 5K Improvement Plan

### Week 1-2: Base Building
- **Easy runs**: 3-4 times per week, 30-40 minutes at conversational pace
- **Tempo run**: 1 time per week, 20 minutes at comfortably hard effort
- **Long run**: 1 time per week, 45-60 minutes easy pace

### Week 3-4: Speed Development
- **Easy runs**: 3 times per week, 30-35 minutes
- **Interval training**: 1 time per week (e.g., 5 x 800m at 5K pace)
- **Tempo run**: 1 time per week, 25 minutes
- **Long run**: 1 time per week, 50-65 minutes

Would you like me to adjust this based on your current weekly mileage?', NOW() - INTERVAL '2 days' + INTERVAL '5 minutes'),
    
    -- Recovery Questions
    ('770e8400-e29b-41d4-a716-446655440003', '660e8400-e29b-41d4-a716-446655440002', 'user', 'I''ve been feeling quite tired lately. How important is recovery in training?', NOW() - INTERVAL '1 day'),
    ('770e8400-e29b-41d4-a716-446655440004', '660e8400-e29b-41d4-a716-446655440002', 'assistant', 'Recovery is absolutely crucial for training adaptation and performance improvement! I''m glad you''re paying attention to how you feel.

## Why Recovery Matters
- **Adaptation**: Your body gets stronger during rest, not during the workout itself
- **Injury prevention**: Adequate recovery reduces overuse injury risk
- **Performance**: Well-recovered athletes perform better in subsequent workouts

## Signs You Need More Recovery
- Persistent fatigue (like you''re experiencing)
- Elevated resting heart rate
- Decreased motivation
- Poor sleep quality
- Increased susceptibility to illness

## Recovery Strategies
1. **Sleep**: Aim for 7-9 hours of quality sleep
2. **Easy days**: Keep easy runs truly easy (conversational pace)
3. **Rest days**: Take 1-2 complete rest days per week
4. **Nutrition**: Proper post-workout nutrition within 30 minutes
5. **Hydration**: Stay consistently hydrated
6. **Active recovery**: Light activities like walking or gentle yoga

Given your current fatigue, I''d recommend taking an extra rest day this week and ensuring your next few runs are at a very comfortable effort. How has your sleep been lately?', NOW() - INTERVAL '1 day' + INTERVAL '10 minutes'),
    
    -- Cycling Technique
    ('770e8400-e29b-41d4-a716-446655440005', '660e8400-e29b-41d4-a716-446655440003', 'user', 'I want to improve my cycling efficiency. Any tips on pedaling technique?', NOW() - INTERVAL '3 hours'),
    ('770e8400-e29b-41d4-a716-446655440006', '660e8400-e29b-41d4-a716-446655440003', 'assistant', 'Great question! Efficient pedaling technique can significantly improve your cycling performance and reduce fatigue.

## Key Elements of Efficient Pedaling

### 1. Cadence
- **Target**: 85-95 RPM for most riding
- **Benefits**: Reduces muscular fatigue, improves cardiovascular efficiency
- **Practice**: Use a cadence sensor and focus on smooth, quick leg turnover

### 2. Pedal Stroke Technique
- **Think circles, not squares**: Apply force throughout the entire pedal stroke
- **12-6 o''clock**: Primary power phase - push down and slightly forward
- **6-9 o''clock**: Pull back and up (if using clipless pedals)
- **9-12 o''clock**: Lift the leg to unweight the pedal

### 3. Body Position
- **Core engagement**: Stable core allows legs to work efficiently
- **Relaxed upper body**: Don''t waste energy gripping handlebars tightly
- **Proper bike fit**: Ensures optimal power transfer

## Drills to Practice
1. **Single-leg drills**: 30 seconds each leg, focus on smooth circles
2. **High cadence intervals**: 2-3 minutes at 100+ RPM
3. **Low cadence climbs**: Build strength with 60-70 RPM on hills

Looking at your recent rides, have you noticed any specific areas where you feel inefficient or experience fatigue?', NOW() - INTERVAL '3 hours' + INTERVAL '8 minutes')
ON CONFLICT (id) DO NOTHING;

-- Insert sample athlete logbooks
INSERT INTO athlete_logbooks (id, user_id, content, updated_at)
VALUES 
    ('880e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440001', 
     '{"profile": {"name": "John Runner", "age": 32, "experience": "intermediate", "goals": ["5K PR", "consistency"]}, "training_history": {"weekly_mileage": 25, "recent_races": ["5K: 22:30"], "strengths": ["consistency", "endurance"], "areas_for_improvement": ["speed", "race tactics"]}, "preferences": {"training_days": ["Monday", "Wednesday", "Friday", "Sunday"], "preferred_workouts": ["tempo runs", "long runs"], "dislikes": ["track intervals"]}, "health_notes": {"injuries": "none recent", "recovery_needs": "standard", "sleep_quality": "good"}}', 
     NOW() - INTERVAL '2 days'),
    
    ('880e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440002', 
     '{"profile": {"name": "Jane Cyclist", "age": 28, "experience": "advanced", "goals": ["century ride", "climbing improvement"]}, "training_history": {"weekly_hours": 8, "recent_events": ["Gran Fondo: 4:15"], "strengths": ["endurance", "consistency"], "areas_for_improvement": ["climbing power", "sprint finish"]}, "preferences": {"training_days": ["Tuesday", "Thursday", "Saturday", "Sunday"], "preferred_workouts": ["long rides", "tempo intervals"], "dislikes": ["indoor trainer sessions"]}, "health_notes": {"injuries": "minor knee sensitivity", "recovery_needs": "extra attention to stretching", "sleep_quality": "excellent"}}', 
     NOW() - INTERVAL '3 hours'),
    
    ('880e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440003', 
     '{"profile": {"name": "Mike Triathlete", "age": 35, "experience": "intermediate", "goals": ["Olympic distance triathlon", "swim improvement"]}, "training_history": {"weekly_hours": 12, "recent_races": ["Sprint tri: 1:15"], "strengths": ["bike leg", "run endurance"], "areas_for_improvement": ["swimming technique", "transitions"]}, "preferences": {"training_days": ["Monday", "Tuesday", "Thursday", "Friday", "Saturday"], "preferred_workouts": ["brick sessions", "open water swims"], "dislikes": ["pool swimming"]}, "health_notes": {"injuries": "none", "recovery_needs": "high due to multi-sport training", "sleep_quality": "variable"}}', 
     NOW() - INTERVAL '1 hour')
ON CONFLICT (user_id) DO NOTHING;

-- Create indexes for better performance in development
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_messages_session_id ON messages(session_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_messages_created_at ON messages(created_at);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_athlete_logbooks_user_id ON athlete_logbooks(user_id);

-- Display summary of seeded data
DO $$
BEGIN
    RAISE NOTICE 'Development data seeded successfully:';
    RAISE NOTICE '- % users', (SELECT COUNT(*) FROM users);
    RAISE NOTICE '- % sessions', (SELECT COUNT(*) FROM sessions);
    RAISE NOTICE '- % messages', (SELECT COUNT(*) FROM messages);
    RAISE NOTICE '- % athlete logbooks', (SELECT COUNT(*) FROM athlete_logbooks);
END $$;