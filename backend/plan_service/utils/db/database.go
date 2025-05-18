package database

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"plan_service/internal/config"
)

var DB *gorm.DB

func InitDB(config *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(config.DatabaseURL), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	DB = db
	return db, nil
}
