package config

import (
	"log"

	"github.com/hidenkeys/motiv-backend/models"
)

func MigrateDatabase() {
	// Create custom enum types first before migrating tables
	DB.Exec("DO $ BEGIN CREATE TYPE user_role AS ENUM ('guest', 'host', 'admin', 'superhost'); EXCEPTION WHEN duplicate_object THEN null; END $;")
	DB.Exec("DO $ BEGIN CREATE TYPE event_status AS ENUM ('draft', 'active', 'cancelled'); EXCEPTION WHEN duplicate_object THEN null; END $;")
	DB.Exec("DO $ BEGIN CREATE TYPE payment_status AS ENUM ('pending', 'completed', 'failed', 'refunded'); EXCEPTION WHEN duplicate_object THEN null; END $;")
	DB.Exec("DO $ BEGIN CREATE TYPE payment_method AS ENUM ('bank_transfer', 'card', 'wallet'); EXCEPTION WHEN duplicate_object THEN null; END $;")
	DB.Exec("DO $ BEGIN CREATE TYPE attendee_status AS ENUM ('active', 'checked_in', 'cancelled'); EXCEPTION WHEN duplicate_object THEN null; END $;")

	err := DB.AutoMigrate(
		&models.User{}, 
		&models.Event{}, 
		&models.Ticket{}, 
		&models.TicketType{}, 
		&models.Wishlist{},
		&models.Review{},
		&models.Payment{},
		&models.Payout{},
		&models.EventView{},
		&models.EventAnalytics{},
		&models.HostAnalytics{},
		&models.Attendee{},
	)
	if err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}
}