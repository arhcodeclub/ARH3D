package db

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/arhcodeclub/arh3d/internal/models"
)

var DB *gorm.DB

func Connect() {
	dsn := "host=localhost user=admin password=admin dbname=arh3d port=5432 sslmode=disable"
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("[DB] Failed to connect to database:", err)
	}

	log.Println("[DB] Connected to database.")

	err = database.AutoMigrate(
		&models.PrintRequest{},
		&models.User{},
		&models.LoginToken{},
	)

    if err != nil {
		log.Fatal("[DB] Failed to migrate database:", err)
	}

	log.Println("[DB] Database migration complete.")

	DB = database
}
