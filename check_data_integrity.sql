-- Quick data integrity check and fix script

-- 1. Check for events with invalid dates
SELECT id, title, start_date, created_at, updated_at
FROM events 
WHERE start_date = '0000-01-01 00:00:00' OR start_date IS NULL
ORDER BY created_at DESC;

-- 2. Check for orphaned events (events without valid hosts)
SELECT e.id, e.title, e.host_id, u.name as host_name
FROM events e
LEFT JOIN users u ON e.host_id = u.id
WHERE u.id IS NULL;

-- 3. Fix invalid dates by setting them to a reasonable default
UPDATE events 
SET start_date = CURRENT_DATE + INTERVAL '7 days'
WHERE start_date = '0000-01-01 00:00:00' OR start_date IS NULL;

-- 4. Check if all events have valid host_ids after the fix
SELECT COUNT(*) as total_events,
       COUNT(CASE WHEN u.id IS NOT NULL THEN 1 END) as events_with_valid_hosts
FROM events e
LEFT JOIN users u ON e.host_id = u.id;