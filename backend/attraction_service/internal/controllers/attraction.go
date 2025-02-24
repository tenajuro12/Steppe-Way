package controllers

import (
	"crypto/rand"
	"diplomaPorject/backend/attraction/internal/models"
	"diplomaPorject/backend/attraction/utils/db"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

func generateRandomFilename(originalFilename string) string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes) + filepath.Ext(originalFilename)
}

type AttractionRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	City        string `json:"city"`
	Address     string `json:"address"`
	Location    string `json:"location"`
	ImageURL    string `json:"image_url"`
}

func CreateAttraction(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	adminIDValue := r.Context().Value("admin_id")
	if adminIDValue == nil {
		http.Error(w, "Unauthorized - admin ID missing", http.StatusUnauthorized)
		return
	}
	adminID, ok := adminIDValue.(uint)
	if !ok {
		http.Error(w, "Internal Server Error - Invalid Admin Id", http.StatusInternalServerError)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "No image file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	imageURL, err := uploadImage(file, header.Filename)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to upload image: %v", err), http.StatusInternalServerError)
		return
	}

	attraction := models.Attraction{
		Title:       r.FormValue("title"),
		Description: r.FormValue("description"),
		City:        r.FormValue("city"),
		Location:    r.FormValue("location"),
		Address:     r.FormValue("address"),
		AdminID:     adminID,
		ImageURL:    imageURL,
	}

	if err := db.DB.Create(&attraction).Error; err != nil {
		http.Error(w, "Failed to create attraction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(attraction)
}
func uploadImage(file io.Reader, filename string) (string, error) {
	uploadDir := "/app/uploads" // Matches Docker volume mount path

	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %v", err)
	}

	randomFilename := generateRandomFilename(filename)
	filePath := filepath.Join(uploadDir, randomFilename) // Use filePath for saving

	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("failed to save file: %v", err)
	}

	return fmt.Sprintf("/uploads/%s", randomFilename), nil
}

func GetAttraction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	var attraction models.Attraction
	if err := db.DB.First(&attraction, id).Error; err != nil {
		http.Error(w, "Not found event", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(attraction)
}

func UpdateAttraction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req AttractionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error with decoding", http.StatusBadRequest)
		return
	}

	var attraction models.Attraction
	if err := db.DB.First(&attraction, id).Error; err != nil {
		http.Error(w, "Error with searching attraction", http.StatusBadRequest)
		return
	}

	attraction.Title = req.Title
	attraction.Description = req.Description
	attraction.City = req.City
	attraction.Address = req.Address
	attraction.Location = req.Location
	attraction.ImageURL = req.ImageURL

	if err := db.DB.Save(&attraction).Error; err != nil {
		http.Error(w, "Can't update attraction", http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(&attraction)
}
func DeleteAttraction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if err := db.DB.Delete(&models.Attraction{}, id).Error; err != nil {
		http.Error(w, "Cant delete attraction", http.StatusBadRequest)
	}
	w.WriteHeader(http.StatusNoContent)
}

func ListAttractions(w http.ResponseWriter, r *http.Request) {
	var attractions []models.Attraction
	query := db.DB
	if err := query.Find(&attractions).Error; err != nil {
		http.Error(w, "Failed to fetch attractions", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(attractions)
}

func PublishAttraction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := db.DB.Model(&models.Attraction{}).Where("id = ?", id).Update("is_published", true).Error; err != nil {
		http.Error(w, "Failed to publish attractions", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func UnpublishAttraction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := db.DB.Model(&models.Attraction{}).Where("id = ?", id).Update("is_published", false).Error; err != nil {
		http.Error(w, "Failed to unpublish event", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func ListPublishedAttractions(w http.ResponseWriter, r *http.Request) {
	var attractions []models.Attraction
	query := db.DB.Where("is_published = ?", true)
	page := 1
	pageSize := 10

	if pageParam := r.URL.Query().Get("page"); pageParam != "" {
		if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
			page = p
		}
	}

	offset := (page - 1) * pageSize

	var totalCount int64
	query.Model(&models.Attraction{}).Count(&totalCount)
	if err := query.
		Order("title ASC").
		Offset(offset).
		Limit(pageSize).
		Find(&attractions).Error; err != nil {
		http.Error(w, "Failed to fetch attractions", http.StatusInternalServerError)
		return
	}

	response := struct {
		Attractions []models.Attraction `json:"attractions"` // ✅ Изменено с "events" на "attractions"
		Total       int64               `json:"total"`
		Page        int                 `json:"page"`
		PageSize    int                 `json:"page_size"`
		TotalPages  int                 `json:"total_pages"`
	}{
		Attractions: attractions,
		Total:       totalCount,
		Page:        page,
		PageSize:    pageSize,
		TotalPages:  int(math.Ceil(float64(totalCount) / float64(pageSize))),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
