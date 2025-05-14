package db

import (
	"food_service/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
)

var DB *gorm.DB

func ConnectDB() {
	dsn := "host=db user=postgres password=123456 dbname=TravelApp port=5432 sslmode=disable"
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		dsn = dbURL
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Connected to the database")

	// Auto migrate the schema
	err = DB.AutoMigrate(
		&models.Place{},
		&models.PlaceImage{},
		&models.Cuisine{},
		&models.Dish{},
		&models.DishImage{},
		&models.FoodReview{},
		&models.FoodReviewImage{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database schemas: %v", err)
	}

	log.Println("Database migration completed")
}
