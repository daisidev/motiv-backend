-- Migration to add location coordinates to events table
-- Run this SQL in your database to add the new location fields

ALTER TABLE events 
ADD COLUMN latitude DOUBLE PRECISION,
ADD COLUMN longitude DOUBLE PRECISION,
ADD COLUMN place_id VARCHAR(255);

-- Add index for location-based queries
CREATE INDEX idx_events_coordinates ON events(latitude, longitude);

-- Add comment for documentation
COMMENT ON COLUMN events.latitude IS 'Latitude coordinate for event location';
COMMENT ON COLUMN events.longitude IS 'Longitude coordinate for event location';
COMMENT ON COLUMN events.place_id IS 'Google Places API place ID for event location';