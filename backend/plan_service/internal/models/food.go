package models

type FoodPlaceResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	City        string `json:"city"`
	Location    string `json:"location"`
	Address     string `json:"address"`
	Type        string `json:"type"`
	PriceRange  string `json:"price_range"`
	ImageURL    string `json:"image_url"`
	Category    string `json:"category"`
}
