// events_service/utils/db/db.go
package db

import (
	"diplomaPorject/backend/events_service/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
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

	err = DB.AutoMigrate(&models.Event{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	log.Println("Connected to PostgreSQL database successfully!")
}
