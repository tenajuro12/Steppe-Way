package controllers

import (
	"diplomaPorject/backend/events_service/internal/models"
	"diplomaPorject/backend/events_service/utils/db"
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"
)

type EventRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	Location    string `json:"location"`
	Capacity    int    `json:"capacity"`
	Category    string `json:"category"`
	ImageURL    string `json:"image_url"`
}

func CreateEvent(w http.ResponseWriter, r *http.Request) {
	adminIDValue := r.Context().Value("admin_id")
	if adminIDValue == nil {
		log.Printf("No admin_id found in context")
		http.Error(w, "Unauthorized - Admin ID missing", http.StatusUnauthorized)
		return
	}

	adminID, ok := adminIDValue.(uint)
	if !ok {
		log.Printf("Invalid admin_id type in context: %T", adminIDValue)
		http.Error(w, "Internal Server Error - Invalid admin ID", http.StatusInternalServerError)
		return
	}

	var req EventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to decode request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	startDate, err := time.Parse(time.RFC3339, req.StartDate)
	if err != nil {
		log.Printf("Invalid start date format: %v", err)
		http.Error(w, "Invalid start date format", http.StatusBadRequest)
		return
	}

	endDate, err := time.Parse(time.RFC3339, req.EndDate)
	if err != nil {
		log.Printf("Invalid end date format: %v", err)
		http.Error(w, "Invalid end date format", http.StatusBadRequest)
		return
	}

	event := models.Event{
		Title:       req.Title,
		Description: req.Description,
		StartDate:   startDate,
		EndDate:     endDate,
		Location:    req.Location,
		Capacity:    req.Capacity,
		AdminID:     adminID,
		Category:    req.Category,
		ImageURL:    req.ImageURL,
	}

	if err := db.DB.Create(&event).Error; err != nil {
		log.Printf("Failed to create event: %v", err)
		http.Error(w, "Failed to create event", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(event)
}

func GetEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var event models.Event
	if err := db.DB.First(&event, id).Error; err != nil {
		http.Error(w, "Not found event", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(event)
}

func UpdateEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req EventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error with decoding", http.StatusBadRequest)
	}

	var event models.Event
	if err := db.DB.First(&event, id).Error; err != nil {
		http.Error(w, "Error with searching event", http.StatusBadRequest)
	}

	startDate, err := time.Parse(time.RFC1123Z, req.StartDate)
	if err != nil {
		http.Error(w, "Error with parsing", http.StatusBadRequest)
	}

	endDate, err := time.Parse(time.RFC1123Z, req.EndDate)
	if err != nil {
		http.Error(w, "Error with parsing", http.StatusBadRequest)
	}

	event.Title = req.Title
	event.Description = req.Description
	event.StartDate = startDate
	event.EndDate = endDate
	event.Location = req.Location
	event.Capacity = req.Capacity
	event.Category = req.Category
	event.ImageURL = req.ImageURL

	if err := db.DB.Save(&event).Error; err != nil {
		http.Error(w, "Cant update event", http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(&event)
}

func DeleteEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	err := db.DB.Delete(&models.Event{}, id).Error
	if err != nil {
		http.Error(w, "Cant delete event", http.StatusBadRequest)
	}
	w.WriteHeader(http.StatusNoContent)
}

func ListEvents(w http.ResponseWriter, r *http.Request) {
	var events []models.Event
	query := db.DB
	if category := r.URL.Query().Get("category"); category != "" {
		query = query.Where("category = ?", category)
	}

	if err := query.Find(&events).Error; err != nil {
		http.Error(w, "Failed to fetch events", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(events)
}
func PublishEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := db.DB.Model(&models.Event{}).Where("id = ?", id).Update("is_published", true).Error; err != nil {
		http.Error(w, "Failed to publish event", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func UnpublishEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := db.DB.Model(&models.Event{}).Where("id = ?", id).Update("is_published", false).Error; err != nil {
		http.Error(w, "Failed to unpublish event", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func ListPublishedEvents(w http.ResponseWriter, r *http.Request) {
	var events []models.Event
	query := db.DB.Where("is_published = ?", true)

	if category := r.URL.Query().Get("category"); category != "" {
		query = query.Where("category = ?", category)
	}

	page := 1
	pageSize := 10
	if pageParam := r.URL.Query().Get("page"); pageParam != "" {
		if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
			page = p
		}
	}

	offset := (page - 1) * pageSize

	var totalCount int64
	query.Model(&models.Event{}).Count(&totalCount)

	if err := query.
		Order("start_date ASC"). // Sort by start date
		Offset(offset).
		Limit(pageSize).
		Find(&events).Error; err != nil {
		http.Error(w, "Failed to fetch events", http.StatusInternalServerError)
		return
	}

	response := struct {
		Events     []models.Event `json:"events"`
		Total      int64          `json:"total"`
		Page       int            `json:"page"`
		PageSize   int            `json:"page_size"`
		TotalPages int            `json:"total_pages"`
	}{
		Events:     events,
		Total:      totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(pageSize))),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
