-- Simple script to reset the payments table with the new structure
-- WARNING: This will delete all existing payment data

-- Drop the payments table
DROP TABLE IF EXISTS payments CASCADE;

-- The table will be recreated automatically by GORM with the new structure
-- when you restart the application