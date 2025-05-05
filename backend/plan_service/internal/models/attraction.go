package models

type AttractionResponse struct {
	ID          uint   `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	City        string `json:"city"`
	Location    string `json:"location"`
	Address     string `json:"address"`
	ImageURL    string `json:"image_url"`
	Category    string `json:"category"`
}
