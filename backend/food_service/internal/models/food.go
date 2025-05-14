package models

import (
	"gorm.io/gorm"
)

type Place struct {
	gorm.Model
	Name          string  `json:"name" gorm:"not null"`
	Description   string  `json:"description"`
	City          string  `json:"city" gorm:"index;not null"`
	Address       string  `json:"address" gorm:"not null"`
	Location      string  `json:"location"`
	Type          string  `json:"type" gorm:"index;not null"`
	PriceRange    string  `json:"price_range"`
	Website       string  `json:"website"`
	Phone         string  `json:"phone"`
	IsPublished   bool    `json:"is_published" gorm:"default:false"`
	AdminID       uint    `json:"admin_id" gorm:"index;not null"`
	AverageRating float64 `json:"average_rating" gorm:"default:0"`

	Cuisines []Cuisine    `json:"cuisines" gorm:"many2many:place_cuisines;"`
	Images   []PlaceImage `json:"images" gorm:"foreignKey:PlaceID;constraint:OnDelete:CASCADE;"`
	Dishes   []Dish       `json:"dishes" gorm:"foreignKey:PlaceID;constraint:OnDelete:CASCADE;"`
	Reviews  []FoodReview `json:"reviews" gorm:"foreignKey:PlaceID;constraint:OnDelete:CASCADE;"`
}

type PlaceImage struct {
	gorm.Model
	PlaceID uint   `json:"place_id" gorm:"index;not null"`
	URL     string `json:"url" gorm:"not null"`
}

type Cuisine struct {
	gorm.Model
	Name        string `json:"name" gorm:"unique;not null"`
	Description string `json:"description"`
	Origin      string `json:"origin"`

	Places []Place `json:"places" gorm:"many2many:place_cuisines;"`
}

type Dish struct {
	gorm.Model
	PlaceID     uint    `json:"place_id" gorm:"index;not null"`
	Name        string  `json:"name" gorm:"not null"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	IsSpecialty bool    `json:"is_specialty" gorm:"default:false"`
	CuisineID   *uint   `json:"cuisine_id"`

	Images []DishImage `json:"images" gorm:"foreignKey:DishID;constraint:OnDelete:CASCADE;"`
}

type DishImage struct {
	gorm.Model
	DishID uint   `json:"dish_id" gorm:"index;not null"`
	URL    string `json:"url" gorm:"not null"`
}

type FoodReview struct {
	gorm.Model
	PlaceID    uint   `json:"place_id" gorm:"index;not null"`
	UserID     uint   `json:"user_id" gorm:"index;not null"`
	Username   string `json:"username"`
	ProfileImg string `json:"profile_img"`
	Rating     int    `json:"rating" gorm:"not null;check:rating >= 1 AND rating <= 5"`
	Comment    string `json:"comment"`

	Images []FoodReviewImage `json:"images" gorm:"foreignKey:ReviewID;constraint:OnDelete:CASCADE;"`
}

type FoodReviewImage struct {
	gorm.Model
	ReviewID uint   `json:"review_id" gorm:"index;not null"`
	URL      string `json:"url" gorm:"not null"`
}

type PlaceWithDistance struct {
	Place
	Distance float64 `json:"distance"`
}
