
package config

import (
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var DB *gorm.DB

func ConnectDatabase() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		panic("DATABASE_URL environment variable is not set!")
	}

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: false, // use plural table names
		},
	})
	if err != nil {
		panic("Failed to connect to database!")
	}

	DB = database
}
