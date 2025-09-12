-- Add exclusive subscription field to users table
ALTER TABLE users ADD COLUMN exclusive BOOLEAN DEFAULT FALSE;
