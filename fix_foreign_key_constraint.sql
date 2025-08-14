-- Fix foreign key constraint violation for events table
-- This script will clean up orphaned events and ensure data consistency

-- First, let's see what events have invalid host_ids
SELECT e.id, e.title, e.host_id, u.id as user_exists
FROM events e
LEFT JOIN users u ON e.host_id = u.id
WHERE u.id IS NULL;

-- Option 1: Delete orphaned events (events without valid hosts)
-- Uncomment the following lines if you want to delete orphaned events
-- DELETE FROM events 
-- WHERE host_id NOT IN (SELECT id FROM users WHERE id IS NOT NULL);

-- Option 2: Create placeholder users for orphaned events
-- This preserves the events but creates dummy host accounts
INSERT INTO users (id, email, name, role, created_at, updated_at)
SELECT DISTINCT 
    e.host_id,
    'placeholder_' || e.host_id || '@example.com',
    'Placeholder Host',
    'host',
    NOW(),
    NOW()
FROM events e
LEFT JOIN users u ON e.host_id = u.id
WHERE u.id IS NULL
ON CONFLICT (id) DO NOTHING;

-- Verify the fix
SELECT COUNT(*) as orphaned_events
FROM events e
LEFT JOIN users u ON e.host_id = u.id
WHERE u.id IS NULL;