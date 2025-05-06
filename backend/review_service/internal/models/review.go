package models

import "time"

type Review struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	AttractionID uint      `json:"attraction_id"`
	UserID       uint      `json:"user_id"`
	Username     string    `json:"username"`
	Rating       int       `json:"rating"`
	Comment      string    `json:"comment"`
	ImageURL     string    `json:"image_url"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
