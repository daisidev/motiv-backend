-- Add newsletter_subscribed column to users table
ALTER TABLE users ADD COLUMN newsletter_subscribed BOOLEAN DEFAULT FALSE;

-- Update existing users to have newsletter_subscribed as false by default
UPDATE users SET newsletter_subscribed = FALSE WHERE newsletter_subscribed IS NULL;