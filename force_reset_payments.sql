-- Force reset of payment-related tables and types
-- Run this in PostgreSQL to completely reset payment system

-- Drop all payment-related tables
DROP TABLE IF EXISTS payments CASCADE;
DROP TABLE IF EXISTS payouts CASCADE;

-- Drop sequences
DROP SEQUENCE IF EXISTS payments_id_seq CASCADE;
DROP SEQUENCE IF EXISTS payouts_id_seq CASCADE;

-- Drop enum types to force recreation
DROP TYPE IF EXISTS payment_status CASCADE;
DROP TYPE IF EXISTS payment_method CASCADE;

-- Verify cleanup
SELECT tablename FROM pg_tables WHERE tablename LIKE '%payment%';
SELECT typname FROM pg_type WHERE typname LIKE '%payment%';