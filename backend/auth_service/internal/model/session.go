package model

import (
	"gorm.io/gorm"
	"time"
)

type Session struct {
	gorm.Model
	Token     string    `gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	UserID    uint      `gorm:"not null"`
}
