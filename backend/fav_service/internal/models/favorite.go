package models

import (
	"time"

	"gorm.io/gorm"
)

type Favorite struct {
	gorm.Model
	UserID      uint      `json:"user_id"`
	ItemID      uint      `json:"item_id"`
	ItemType    string    `json:"item_type"`
	Title       string    `json:"title"`
	ImageURL    string    `json:"image_url"`
	Description string    `json:"description"`
	City        string    `json:"city,omitempty"`
	Location    string    `json:"location,omitempty"`
	Date        time.Time `json:"date,omitempty"`
	Category    string    `json:"category,omitempty"`
}

type FavoriteResponse struct {
	ID          uint      `json:"id"`
	UserID      uint      `json:"user_id"`
	ItemID      uint      `json:"item_id"`
	ItemType    string    `json:"item_type"`
	Title       string    `json:"title"`
	ImageURL    string    `json:"image_url"`
	Description string    `json:"description,omitempty"`
	City        string    `json:"city,omitempty"`
	Location    string    `json:"location,omitempty"`
	Date        time.Time `json:"date,omitempty"`
	Category    string    `json:"category,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type FavoriteRequest struct {
	ItemID      uint      `json:"item_id" binding:"required"`
	ItemType    string    `json:"item_type" binding:"required"`
	Title       string    `json:"title" binding:"required"`
	ImageURL    string    `json:"image_url"`
	Description string    `json:"description"`
	City        string    `json:"city"`
	Location    string    `json:"location"`
	Date        time.Time `json:"date"`
	Category    string    `json:"category"`
}

func (f *Favorite) ToResponse() FavoriteResponse {
	return FavoriteResponse{
		ID:          f.ID,
		UserID:      f.UserID,
		ItemID:      f.ItemID,
		ItemType:    f.ItemType,
		Title:       f.Title,
		ImageURL:    f.ImageURL,
		Description: f.Description,
		City:        f.City,
		Location:    f.Location,
		Date:        f.Date,
		Category:    f.Category,
		CreatedAt:   f.CreatedAt,
	}
}
