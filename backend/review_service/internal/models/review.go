package model

import "gorm.io/gorm"

type Review struct {
	gorm.Model
	Content      string  `json:"content"`
	Rating       float64 `json:"rating" gorm:"type:decimal(2,1);check:rating >= 1 AND rating <= 5"`
	AttractionID uint    `json:"attraction_id"`
	UserID       uint    `json:"user_id"`
	Username     string  `json:"username"`
}
