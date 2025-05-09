package models

import (
	"gorm.io/gorm"
	"time"
)

type Event struct {
	gorm.Model
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	StartDate    time.Time `json:"start_date"`
	EndDate      time.Time `json:"end_date"`
	Location     string    `json:"location"`
	Address      string    `json:"address"`
	Link         string    `json:"link"`
	Capacity     int       `json:"capacity"`
	IsPublished  bool      `json:"is_published" gorm:"default:false"`
	AdminID      uint      `json:"admin_id"`
	CurrentCount int       `json:"current_count" gorm:"default:0"`
	ImageURL     string    `json:"image_url"`
	Category     string    `json:"category"`
}
