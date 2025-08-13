-- Migration to update payment table structure
-- Remove foreign key constraint to tickets and add event_id and user_id

-- First, drop the foreign key constraint
ALTER TABLE payments DROP CONSTRAINT IF EXISTS fk_payments_ticket;

-- Add new columns
ALTER TABLE payments ADD COLUMN IF NOT EXISTS event_id UUID;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS user_id UUID;

-- Update existing payments to have event_id and user_id from their tickets
UPDATE payments 
SET event_id = t.event_id, user_id = t.user_id
FROM tickets t 
WHERE payments.ticket_id = t.id;

-- Make the new columns NOT NULL after updating
ALTER TABLE payments ALTER COLUMN event_id SET NOT NULL;
ALTER TABLE payments ALTER COLUMN user_id SET NOT NULL;

-- Add foreign key constraints for the new columns
ALTER TABLE payments ADD CONSTRAINT fk_payments_event FOREIGN KEY (event_id) REFERENCES events(id);
ALTER TABLE payments ADD CONSTRAINT fk_payments_user FOREIGN KEY (user_id) REFERENCES users(id);

-- Drop the old ticket_id column
ALTER TABLE payments DROP COLUMN IF EXISTS ticket_id;

-- Add indexes for better performance
CREATE INDEX IF NOT EXISTS idx_payments_event_id ON payments(event_id);
CREATE INDEX IF NOT EXISTS idx_payments_user_id ON payments(user_id);
CREATE INDEX IF NOT EXISTS idx_payments_reference ON payments(reference);
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);