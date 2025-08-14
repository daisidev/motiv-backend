-- Change the default value for event status from 'draft' to 'active'
ALTER TABLE events ALTER COLUMN status SET DEFAULT 'active';

-- Update any existing draft events to active (if desired)
-- Uncomment the line below if you want to update existing draft events
-- UPDATE events SET status = 'active' WHERE status = 'draft';
