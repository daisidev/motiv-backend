package config

import (
	"log"

	"github.com/hidenkeys/motiv-backend/models"
)

func MigrateDatabase() {
	// First, migrate the basic models without custom enums
	err := DB.AutoMigrate(
		&models.User{}, 
		&models.Event{}, 
		&models.Ticket{}, 
		&models.TicketType{}, 
		&models.Wishlist{},
	)
	if err != nil {
		log.Printf("Warning: failed to migrate basic models: %v", err)
	}

	// Try to create enum types and migrate advanced models
	// If this fails, the basic functionality will still work
	createEnumIfNotExists := func(enumName, enumValues string) {
		var exists bool
		err := DB.Raw("SELECT EXISTS (SELECT 1 FROM pg_type WHERE typname = ?)", enumName).Scan(&exists).Error
		if err != nil {
			log.Printf("Warning: failed to check enum %s: %v", enumName, err)
			return
		}
		if !exists {
			err := DB.Exec("CREATE TYPE " + enumName + " AS ENUM " + enumValues).Error
			if err != nil {
				log.Printf("Warning: failed to create enum %s: %v", enumName, err)
			}
		}
	}

	createEnumIfNotExists("user_role", "('guest', 'host', 'admin', 'superhost')")
	createEnumIfNotExists("event_status", "('draft', 'active', 'cancelled')")
	createEnumIfNotExists("payment_status", "('pending', 'completed', 'failed', 'refunded')")
	createEnumIfNotExists("payment_method", "('bank_transfer', 'card', 'wallet')")
	createEnumIfNotExists("attendee_status", "('active', 'checked_in', 'cancelled')")

	// Try to migrate advanced models
	err = DB.AutoMigrate(
		&models.Review{},
		&models.Payment{},
		&models.Payout{},
		&models.EventView{},
		&models.EventAnalytics{},
		&models.HostAnalytics{},
		&models.Attendee{},
	)
	if err != nil {
		log.Printf("Warning: failed to migrate advanced models: %v", err)
		log.Println("Basic functionality will still work. Advanced features may be limited.")
	}

	log.Println("Database migration completed")
}