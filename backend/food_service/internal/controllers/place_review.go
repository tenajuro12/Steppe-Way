package controllers

import (
	"encoding/json"
	"fmt"
	"food_service/internal/models"
	"food_service/utils/db"
	"github.com/gorilla/mux"
	"log"
	"math"
	"net/http"
	"strconv"
)

func (c *FoodController) AddReview(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	placeIDStr := vars["id"]
	placeID, err := strconv.ParseUint(placeIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid place ID", http.StatusBadRequest)
		return
	}

	var place models.Place
	if err := db.DB.First(&place, placeID).Error; err != nil {
		http.Error(w, "Place not found", http.StatusNotFound)
		return
	}

	userID, err := getUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized - user ID missing", http.StatusUnauthorized)
		return
	}

	username, profileImg, err := getUserProfileInfo(userID)
	if err != nil {
		log.Printf("Warning: couldn't get user profile info: %v", err)
	}

	ratingStr := r.FormValue("rating")
	rating, err := strconv.Atoi(ratingStr)
	if err != nil || rating < 1 || rating > 5 {
		http.Error(w, "Invalid rating - must be between 1 and 5", http.StatusBadRequest)
		return
	}

	comment := r.FormValue("comment")

	review := models.FoodReview{
		PlaceID:    uint(placeID),
		UserID:     userID,
		Username:   username,
		ProfileImg: profileImg,
		Rating:     rating,
		Comment:    comment,
	}

	tx := db.DB.Begin()
	if err := tx.Create(&review).Error; err != nil {
		tx.Rollback()
		http.Error(w, "Failed to create review", http.StatusInternalServerError)
		return
	}

	files := r.MultipartForm.File["images"]
	if len(files) > 0 {
		imageURLs, err := uploadImages(files, "reviews")
		if err != nil {
			tx.Rollback()
			http.Error(w, fmt.Sprintf("Failed to upload images: %v", err), http.StatusInternalServerError)
			return
		}

		for _, url := range imageURLs {
			image := models.FoodReviewImage{
				ReviewID: review.ID,
				URL:      url,
			}
			if err := tx.Create(&image).Error; err != nil {
				tx.Rollback()
				http.Error(w, "Failed to save review image", http.StatusInternalServerError)
				return
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	if err := updateAverageRating(uint(placeID)); err != nil {
		log.Printf("Warning: failed to update average rating: %v", err)
	}

	if err := db.DB.Preload("Images").First(&review, review.ID).Error; err != nil {
		http.Error(w, "Failed to reload review", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(review)
}
func (c *FoodController) UpdateReview(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	reviewIDStr := vars["review_id"]
	reviewID, err := strconv.ParseUint(reviewIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid review ID", http.StatusBadRequest)
		return
	}

	var review models.FoodReview
	if err := db.DB.First(&review, reviewID).Error; err != nil {
		http.Error(w, "Review not found", http.StatusNotFound)
		return
	}

	userID, err := getUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized - user ID missing", http.StatusUnauthorized)
		return
	}

	if userID != review.UserID {
		http.Error(w, "Unauthorized - not the review owner", http.StatusUnauthorized)
		return
	}

	if ratingStr := r.FormValue("rating"); ratingStr != "" {
		rating, err := strconv.Atoi(ratingStr)
		if err != nil || rating < 1 || rating > 5 {
			http.Error(w, "Invalid rating - must be between 1 and 5", http.StatusBadRequest)
			return
		}
		review.Rating = rating
	}

	if comment := r.FormValue("comment"); comment != "" {
		review.Comment = comment
	}

	tx := db.DB.Begin()

	files := r.MultipartForm.File["images"]
	if len(files) > 0 {
		if deleteExisting := r.FormValue("delete_existing_images"); deleteExisting == "true" {
			if err := tx.Where("review_id = ?", review.ID).Delete(&models.FoodReviewImage{}).Error; err != nil {
				tx.Rollback()
				http.Error(w, "Failed to delete existing images", http.StatusInternalServerError)
				return
			}
		}

		imageURLs, err := uploadImages(files, "reviews")
		if err != nil {
			tx.Rollback()
			http.Error(w, fmt.Sprintf("Failed to upload images: %v", err), http.StatusInternalServerError)
			return
		}

		for _, url := range imageURLs {
			image := models.FoodReviewImage{
				ReviewID: review.ID,
				URL:      url,
			}
			if err := tx.Create(&image).Error; err != nil {
				tx.Rollback()
				http.Error(w, "Failed to save review image", http.StatusInternalServerError)
				return
			}
		}
	}

	if err := tx.Save(&review).Error; err != nil {
		tx.Rollback()
		http.Error(w, "Failed to update review", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit().Error; err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	if err := updateAverageRating(review.PlaceID); err != nil {
		log.Printf("Warning: failed to update average rating: %v", err)
	}

	if err := db.DB.Preload("Images").First(&review, review.ID).Error; err != nil {
		http.Error(w, "Failed to reload review", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(review)
}

func (c *FoodController) DeleteReview(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	reviewIDStr := vars["review_id"]
	reviewID, err := strconv.ParseUint(reviewIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid review ID", http.StatusBadRequest)
		return
	}

	var review models.FoodReview
	if err := db.DB.First(&review, reviewID).Error; err != nil {
		http.Error(w, "Review not found", http.StatusNotFound)
		return
	}

	userID, err := getUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized - user ID missing", http.StatusUnauthorized)
		return
	}

	if userID != review.UserID {
		var place models.Place
		if err := db.DB.First(&place, review.PlaceID).Error; err != nil {
			http.Error(w, "Unauthorized - not the review owner", http.StatusUnauthorized)
			return
		}

		adminIDValue := r.Context().Value("admin_id")
		if adminIDValue == nil {
			http.Error(w, "Unauthorized - not the review owner", http.StatusUnauthorized)
			return
		}
		adminID, ok := adminIDValue.(uint)
		if !ok || adminID != place.AdminID {
			http.Error(w, "Unauthorized - not the review owner or place owner", http.StatusUnauthorized)
			return
		}
	}

	placeID := review.PlaceID

	if err := db.DB.Delete(&review).Error; err != nil {
		http.Error(w, "Failed to delete review", http.StatusInternalServerError)
		return
	}

	if err := updateAverageRating(placeID); err != nil {
		log.Printf("Warning: failed to update average rating: %v", err)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Review deleted successfully"})
}

func (c *FoodController) GetPlaceReviews(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var reviews []models.FoodReview
	query := db.DB.Preload("Images").Where("place_id = ?", id)

	page := 1
	pageSize := 20
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if pageInt, err := strconv.Atoi(pageStr); err == nil && pageInt > 0 {
			page = pageInt
		}
	}
	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		if pageSizeInt, err := strconv.Atoi(pageSizeStr); err == nil && pageSizeInt > 0 {
			pageSize = pageSizeInt
		}
	}

	var totalCount int64
	query.Model(&models.FoodReview{}).Count(&totalCount)

	offset := (page - 1) * pageSize
	query = query.Offset(offset).Limit(pageSize)

	query = query.Order("created_at desc")

	if err := query.Find(&reviews).Error; err != nil {
		http.Error(w, "Failed to fetch reviews", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"reviews": reviews,
		"pagination": map[string]interface{}{
			"page":        page,
			"page_size":   pageSize,
			"total":       totalCount,
			"total_pages": int(math.Ceil(float64(totalCount) / float64(pageSize))),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
