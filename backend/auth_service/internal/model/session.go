package model

import (
	"gorm.io/gorm"
	"time"
)

type Session struct {
	gorm.Model
	Token     string    `gorm:"uniqueIndex;not null"` // Unique session token
	ExpiresAt time.Time `gorm:"not null"`             // Expiration time
	UserID    uint      `gorm:"not null"`             // User reference
}
