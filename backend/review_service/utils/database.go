package utils

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"review_service/internal/models"
)

var DB *gorm.DB

func ConnectDB() {
	dsn := "host=db user=postgres password=123456 dbname=TravelApp port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	DB = db

	if DB == nil {
		log.Fatal("Database connection is nil after initialization!")
	}

	err = DB.AutoMigrate(&models.Review{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	log.Println("Connected to PostgreSQL database successfully!")
}
