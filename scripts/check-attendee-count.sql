-- SQL query to check attendees count per event
-- Run this in your PostgreSQL database to verify the data

SELECT 
    e.id as event_id,
    e.title as event_title,
    COUNT(a.id) as attendee_count
FROM 
    events e
LEFT JOIN 
    attendees a ON e.id = a.event_id
GROUP BY 
    e.id, e.title
ORDER BY 
    attendee_count DESC;

-- Also check the attendees table structure
SELECT 
    id,
    event_id,
    attendee_full_name,
    attendee_email,
    checked_in,
    created_at
FROM 
    attendees
ORDER BY 
    created_at DESC
LIMIT 20;
