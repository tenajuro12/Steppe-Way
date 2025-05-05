package models

import "time"

type EventResponse struct {
	ID          uint      `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	Location    string    `json:"location"`
	ImageURL    string    `json:"image_url"`
	Category    string    `json:"category"`
}
