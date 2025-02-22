package db

import (
	"authorization_service/internal/model"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"math"
	"os"
	"time"
)

var DB *gorm.DB

func ConnectDB() {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "db" // Use the service name from docker-compose
	}

	user := os.Getenv("DB_USER")
	if user == "" {
		user = "postgres"
	}

	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "123456"
	}

	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "TravelApp"
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable",
		host, user, password, dbname,
	)

	var err error
	maxAttempts := 5
	for attempts := 0; attempts < maxAttempts; attempts++ {
		backoffDuration := time.Second * time.Duration(math.Pow(2, float64(attempts)))
		log.Printf("Attempting to connect to database (attempt %d/%d)...\n", attempts+1, maxAttempts)

		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}

		log.Printf("Failed to connect: %v\n", err)
		if attempts < maxAttempts-1 {
			log.Printf("Waiting %v before next attempt...\n", backoffDuration)
			time.Sleep(backoffDuration)
		}
	}

	if err != nil {
		log.Fatalf("Failed to connect to database after %d attempts: %v", maxAttempts, err)
	}

	if DB == nil {
		log.Fatal("Database connection is nil after initialization!")
	}

	// Run AutoMigrate for your models
	err = DB.AutoMigrate(&model.User{}, &model.Session{}) // Add your models here
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	log.Println("Connected to PostgreSQL database successfully!")
}
