-- Add qr_data column to tickets table
ALTER TABLE tickets ADD COLUMN qr_data TEXT;

-- Update existing tickets to populate qr_data field from qr_code if it contains MOTIV-TICKET format
UPDATE tickets 
SET qr_data = qr_code 
WHERE qr_code LIKE 'MOTIV-TICKET:%' AND (qr_data IS NULL OR qr_data = '');

-- For tickets that have base64 QR codes, we'll need to regenerate the qr_data
-- This will be handled by the regenerate QR endpoint