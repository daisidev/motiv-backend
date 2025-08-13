package main

import (
	"log"

	"github.com/hidenkeys/motiv-backend/config"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

func main() {
	// Load .env file
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Println("Error loading .env file, using environment variables")
	}

	// Connect to the database
	config.ConnectDatabase()
	db := config.DB

	log.Println("Starting payment table migration...")

	// Execute the migration
	err = migratePaymentTable(db)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Payment table migration completed successfully!")
}

func migratePaymentTable(db *gorm.DB) error {
	// Check if we need to migrate (if event_id column doesn't exist)
	var count int64
	err := db.Raw("SELECT COUNT(*) FROM information_schema.columns WHERE table_name = 'payments' AND column_name = 'event_id'").Scan(&count).Error
	if err != nil {
		return err
	}

	if count > 0 {
		log.Println("Payment table already migrated, skipping...")
		return nil
	}

	log.Println("Adding event_id and user_id columns...")

	// Add new columns
	err = db.Exec("ALTER TABLE payments ADD COLUMN event_id UUID").Error
	if err != nil {
		return err
	}

	err = db.Exec("ALTER TABLE payments ADD COLUMN user_id UUID").Error
	if err != nil {
		return err
	}

	log.Println("Updating existing payment records...")

	// Update existing payments to have event_id and user_id from their tickets
	err = db.Exec(`
		UPDATE payments 
		SET event_id = t.event_id, user_id = t.user_id
		FROM tickets t 
		WHERE payments.ticket_id = t.id
	`).Error
	if err != nil {
		return err
	}

	log.Println("Making new columns NOT NULL...")

	// Make the new columns NOT NULL
	err = db.Exec("ALTER TABLE payments ALTER COLUMN event_id SET NOT NULL").Error
	if err != nil {
		return err
	}

	err = db.Exec("ALTER TABLE payments ALTER COLUMN user_id SET NOT NULL").Error
	if err != nil {
		return err
	}

	log.Println("Adding foreign key constraints...")

	// Add foreign key constraints
	err = db.Exec("ALTER TABLE payments ADD CONSTRAINT fk_payments_event FOREIGN KEY (event_id) REFERENCES events(id)").Error
	if err != nil {
		return err
	}

	err = db.Exec("ALTER TABLE payments ADD CONSTRAINT fk_payments_user FOREIGN KEY (user_id) REFERENCES users(id)").Error
	if err != nil {
		return err
	}

	log.Println("Dropping old ticket_id column...")

	// Drop the old foreign key constraint first
	err = db.Exec("ALTER TABLE payments DROP CONSTRAINT IF EXISTS fk_payments_ticket").Error
	if err != nil {
		return err
	}

	// Drop the old ticket_id column
	err = db.Exec("ALTER TABLE payments DROP COLUMN ticket_id").Error
	if err != nil {
		return err
	}

	log.Println("Adding indexes...")

	// Add indexes for better performance
	db.Exec("CREATE INDEX IF NOT EXISTS idx_payments_event_id ON payments(event_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_payments_user_id ON payments(user_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_payments_reference ON payments(reference)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status)")

	return nil
}
