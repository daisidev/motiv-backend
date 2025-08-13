-- Reset payments table to fix foreign key issues
-- Run this in your PostgreSQL database before starting the backend

-- Drop payments table (this will remove all payment data)
DROP TABLE IF EXISTS payments CASCADE;

-- The payments table will be recreated automatically by GORM 
-- when you restart the Go backend with the new structure