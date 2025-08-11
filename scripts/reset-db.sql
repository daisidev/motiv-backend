-- Script to reset the database if needed
-- Run this if you want to start fresh

-- Drop all tables (in reverse dependency order)
DROP TABLE IF EXISTS attendees CASCADE;
DROP TABLE IF EXISTS host_analytics CASCADE;
DROP TABLE IF EXISTS event_analytics CASCADE;
DROP TABLE IF EXISTS event_views CASCADE;
DROP TABLE IF EXISTS payouts CASCADE;
DROP TABLE IF EXISTS payments CASCADE;
DROP TABLE IF EXISTS reviews CASCADE;
DROP TABLE IF EXISTS wishlists CASCADE;
DROP TABLE IF EXISTS tickets CASCADE;
DROP TABLE IF EXISTS ticket_types CASCADE;
DROP TABLE IF EXISTS events CASCADE;
DROP TABLE IF EXISTS users CASCADE;

-- Drop all custom types
DROP TYPE IF EXISTS attendee_status CASCADE;
DROP TYPE IF EXISTS payment_method CASCADE;
DROP TYPE IF EXISTS payment_status CASCADE;
DROP TYPE IF EXISTS event_status CASCADE;
DROP TYPE IF EXISTS user_role CASCADE;