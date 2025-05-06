package models

import (
	"gorm.io/gorm"
)

type Attraction struct {
	gorm.Model
	Title       string `json:"title"`
	Description string `json:"description"`
	City        string `json:"city"`
	Location    string `json:"location"`
	Address     string `json:"address"`
	IsPublished bool   `json:"is_published" gorm:"default:false"`
	AdminID     uint   `json:"admin_id"`
	ImageURL    string `json:"image_url"`
	Category    string `json:"category"`
}
