package db

import (
	"diplomaPorject/backend/attraction/internal/models"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
)

var DB *gorm.DB

func ConnectDB() {
	dsn := "host=db user=postgres password=123456 dbname=TravelApp port=5432 sslmode=disable"
	dbInstance, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	DB = dbInstance

	err = DB.AutoMigrate(&models.Attraction{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	fmt.Println("Connected to PostgreSQL database!")
}
