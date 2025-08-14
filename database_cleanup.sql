-- Comprehensive database cleanup script
-- Run this to fix foreign key constraint violations

-- 1. Check for orphaned events (events without valid hosts)
DO $$
DECLARE
    orphaned_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO orphaned_count
    FROM events e
    LEFT JOIN users u ON e.host_id = u.id
    WHERE u.id IS NULL;
    
    RAISE NOTICE 'Found % orphaned events', orphaned_count;
END $$;

-- 2. Check for orphaned ticket_types (ticket_types without valid events)
DO $$
DECLARE
    orphaned_tickets INTEGER;
BEGIN
    SELECT COUNT(*) INTO orphaned_tickets
    FROM ticket_types tt
    LEFT JOIN events e ON tt.event_id = e.id
    WHERE e.id IS NULL;
    
    RAISE NOTICE 'Found % orphaned ticket types', orphaned_tickets;
END $$;

-- 3. Fix orphaned events by creating placeholder users
INSERT INTO users (id, email, name, role, created_at, updated_at)
SELECT DISTINCT 
    e.host_id,
    'placeholder_' || REPLACE(e.host_id::text, '-', '') || '@motiv.placeholder',
    'Placeholder Host ' || SUBSTRING(e.host_id::text, 1, 8),
    'host',
    NOW(),
    NOW()
FROM events e
LEFT JOIN users u ON e.host_id = u.id
WHERE u.id IS NULL
ON CONFLICT (id) DO NOTHING;

-- 4. Clean up orphaned ticket types
DELETE FROM ticket_types 
WHERE event_id NOT IN (SELECT id FROM events WHERE id IS NOT NULL);

-- 5. Clean up orphaned tickets
DELETE FROM tickets 
WHERE event_id NOT IN (SELECT id FROM events WHERE id IS NOT NULL);

-- 6. Clean up orphaned payments (if they reference non-existent events or users)
DELETE FROM payments 
WHERE event_id IS NOT NULL 
AND event_id NOT IN (SELECT id FROM events WHERE id IS NOT NULL);

DELETE FROM payments 
WHERE user_id IS NOT NULL 
AND user_id NOT IN (SELECT id FROM users WHERE id IS NOT NULL);

-- 7. Verify the cleanup
SELECT 
    'events' as table_name,
    COUNT(*) as total_records,
    COUNT(CASE WHEN u.id IS NULL THEN 1 END) as orphaned_records
FROM events e
LEFT JOIN users u ON e.host_id = u.id

UNION ALL

SELECT 
    'ticket_types' as table_name,
    COUNT(*) as total_records,
    COUNT(CASE WHEN e.id IS NULL THEN 1 END) as orphaned_records
FROM ticket_types tt
LEFT JOIN events e ON tt.event_id = e.id

UNION ALL

SELECT 
    'tickets' as table_name,
    COUNT(*) as total_records,
    COUNT(CASE WHEN e.id IS NULL THEN 1 END) as orphaned_records
FROM tickets t
LEFT JOIN events e ON t.event_id = e.id;

RAISE NOTICE 'Database cleanup completed successfully';