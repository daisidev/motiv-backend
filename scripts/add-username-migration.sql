-- Migration to add username column to users table
-- Run this if you have existing data and don't want to reset the database

-- Add username column (nullable first)
ALTER TABLE users ADD COLUMN username VARCHAR(255);

-- Create a unique index on username (for future uniqueness)
CREATE UNIQUE INDEX CONCURRENTLY idx_users_username ON users (username) WHERE username IS NOT NULL;

-- Update existing users with temporary usernames based on email
-- You can customize this logic based on your needs
UPDATE users 
SET username = LOWER(SPLIT_PART(email, '@', 1)) || '_' || EXTRACT(EPOCH FROM created_at)::INTEGER
WHERE username IS NULL;

-- Make username NOT NULL after populating
ALTER TABLE users ALTER COLUMN username SET NOT NULL;

-- Add unique constraint
ALTER TABLE users ADD CONSTRAINT users_username_unique UNIQUE (username);