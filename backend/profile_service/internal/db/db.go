package db

import (
	"log"
	"profile_service/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	dsn := "host=db user=postgres password=123456 dbname=TravelApp port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Could not connect to the database: %v", err)
	}

	err = db.AutoMigrate(&models.Profile{})
	if err != nil {
		log.Fatalf("Failed to auto-migrate: %v", err)
	}

	DB = db
}
