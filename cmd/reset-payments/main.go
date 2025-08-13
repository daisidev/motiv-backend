package main

import (
	"log"

	"github.com/hidenkeys/motiv-backend/config"
	"github.com/joho/godotenv"
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

	log.Println("Resetting payment tables...")

	// Drop all payment-related tables and types
	commands := []string{
		"DROP TABLE IF EXISTS payments CASCADE",
		"DROP TABLE IF EXISTS payouts CASCADE",
		"DROP SEQUENCE IF EXISTS payments_id_seq CASCADE",
		"DROP SEQUENCE IF EXISTS payouts_id_seq CASCADE",
		"DROP TYPE IF EXISTS payment_status CASCADE",
		"DROP TYPE IF EXISTS payment_method CASCADE",
	}

	for _, cmd := range commands {
		err := db.Exec(cmd).Error
		if err != nil {
			log.Printf("Warning executing '%s': %v", cmd, err)
		} else {
			log.Printf("✅ Executed: %s", cmd)
		}
	}

	log.Println("✅ Payment tables reset complete!")
	log.Println("Now restart your main application to recreate tables with new structure.")
}
