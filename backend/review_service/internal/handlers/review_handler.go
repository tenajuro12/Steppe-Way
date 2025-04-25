package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"review_service/internal/models"
	"review_service/internal/service"
	"review_service/utils"
	"strconv"
	"strings"
	"time"
)

var reviewService = service.NewReviewService()

func ReviewRouter(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		CreateReview(w, r)
	case http.MethodGet:
		GetReviews(w, r)
	}
}

func ReviewByIDRouter(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/reviews/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodPut:
		UpdateReview(w, r, uint(id))
	case http.MethodDelete:
		DeleteReview(w, r, uint(id))
	}
}

func CreateReview(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)
	attractionID, _ := strconv.Atoi(r.FormValue("attraction_id"))
	rating, _ := strconv.Atoi(r.FormValue("rating"))
	comment := r.FormValue("comment")

	username := r.Header.Get("X-Username")
	userIDStr := r.Header.Get("X-User-ID")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var imageURL string
	file, handler, err := r.FormFile("image")
	if err == nil {
		defer file.Close()
		os.MkdirAll("uploads", os.ModePerm)
		filename := fmt.Sprintf("uploads/%d_%s", time.Now().UnixNano(), handler.Filename)
		dst, _ := os.Create(filename)
		defer dst.Close()
		io.Copy(dst, file)
		imageURL = "/" + filename
	}

	review := models.Review{
		AttractionID: uint(attractionID),
		UserID:       uint(userID),
		Rating:       rating,
		Comment:      comment,
		Username:     username,
		ImageURL:     imageURL,
	}

	if err := reviewService.Create(&review); err != nil {
		http.Error(w, "Could not create review", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(review)
}

func GetReviews(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("attraction_id")
	if idStr != "" {
		id, _ := strconv.Atoi(idStr)
		reviews, _ := reviewService.GetByAttraction(uint(id))
		json.NewEncoder(w).Encode(reviews)
		return
	}

	reviews, _ := reviewService.GetAll()
	json.NewEncoder(w).Encode(reviews)
}

func UpdateReview(w http.ResponseWriter, r *http.Request, id uint) {
	var input models.Review
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// 🧠 Загружаем существующий отзыв из БД
	var existing models.Review
	if err := utils.DB.First(&existing, id).Error; err != nil {
		http.Error(w, "Review not found", http.StatusNotFound)
		return
	}

	// 🔧 Обновляем только нужные поля
	existing.Comment = input.Comment
	existing.Rating = input.Rating
	existing.UpdatedAt = time.Now()

	// 💾 Сохраняем
	if err := utils.DB.Save(&existing).Error; err != nil {
		http.Error(w, "Failed to update", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existing)
}

func DeleteReview(w http.ResponseWriter, r *http.Request, id uint) {
	if err := reviewService.Delete(id); err != nil {
		http.Error(w, "Failed to delete", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
