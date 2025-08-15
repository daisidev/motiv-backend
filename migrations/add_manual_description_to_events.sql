-- Add manual_description column to events table
ALTER TABLE events ADD COLUMN IF NOT EXISTS manual_description TEXT;