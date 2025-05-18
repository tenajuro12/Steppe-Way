package models

type AccommodationResponse struct {
	ID          uint     `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	City        string   `json:"city"`
	Location    string   `json:"location"`
	Address     string   `json:"address"`
	Type        string   `json:"type"`
	Website     string   `json:"website"`
	ImageURL    string   `json:"image_url"`
	Amenities   []string `json:"amenities"`
}
