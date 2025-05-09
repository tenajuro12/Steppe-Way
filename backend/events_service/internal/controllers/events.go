package controllers

import (
	"diplomaPorject/backend/events_service/internal/models"
	"diplomaPorject/backend/events_service/utils/db"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type EventRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	Location    string `json:"location"`
	Address     string `json:"address"`
	Link        string `json:"link"`
	Capacity    int    `json:"capacity"`
	Category    string `json:"category"`
	ImageURL    string `json:"image_url"`
}

func generateRandomFilename(originalFilename string) string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%d_%s", timestamp, filepath.Base(originalFilename))
}

func uploadImage(file io.Reader, filename string) (string, error) {
	uploadDir := "/app/uploads/events"

	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %v", err)
	}

	randomFilename := generateRandomFilename(filename)
	filePath := filepath.Join(uploadDir, randomFilename)

	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("failed to save file: %v", err)
	}

	return fmt.Sprintf("/uploads/events/%s", randomFilename), nil
}

func CreateEvent(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		log.Printf("Failed to parse form: %v", err)
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

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

	var imageURL string
	file, header, err := r.FormFile("image")
	if err == nil {
		defer file.Close()
		imageURL, err = uploadImage(file, header.Filename)
		if err != nil {
			log.Printf("Failed to upload image: %v", err)
			http.Error(w, "Failed to upload image", http.StatusInternalServerError)
			return
		}
	} else {
		log.Println("No image uploaded. Proceeding without image.")
	}

	title := r.FormValue("title")
	description := r.FormValue("description")
	location := r.FormValue("location")
	category := r.FormValue("category")
	address := r.FormValue("address")
	link := r.FormValue("link")
	capacityStr := r.FormValue("capacity")
	startDateStr := r.FormValue("start_date")
	endDateStr := r.FormValue("end_date")

	capacity, err := strconv.Atoi(capacityStr)
	if err != nil {
		log.Printf("Invalid capacity value: %v", err)
		http.Error(w, "Invalid capacity value", http.StatusBadRequest)
		return
	}

	startDate, err := time.Parse(time.RFC3339, startDateStr)
	if err != nil {
		log.Printf("Invalid start date format: %v", err)
		http.Error(w, "Invalid start date format (expected RFC3339)", http.StatusBadRequest)
		return
	}

	endDate, err := time.Parse(time.RFC3339, endDateStr)
	if err != nil {
		log.Printf("Invalid end date format: %v", err)
		http.Error(w, "Invalid end date format (expected RFC3339)", http.StatusBadRequest)
		return
	}

	event := models.Event{
		Title:       title,
		Description: description,
		StartDate:   startDate,
		EndDate:     endDate,
		Location:    location,
		Address:     address,
		Link:        link,
		Capacity:    capacity,
		Category:    category,
		ImageURL:    imageURL,
		AdminID:     adminID,
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
	event.Address = req.Address
	event.Link = req.Link
	
	if err := db.DB.Save(&event).Error; err != nil {
		http.Error(w, "Cant update event", http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(event)
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
