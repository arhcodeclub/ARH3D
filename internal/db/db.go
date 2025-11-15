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
        log.Fatal("Failed to connect to database:", err)
    }

    // Auto migrate
    err = database.AutoMigrate(&models.PrintRequest{})
    if err != nil {
        log.Fatal("Failed to migrate database:", err)
    }

    DB = database
    log.Println("Connected to PostgreSQL with GORM!")
}

