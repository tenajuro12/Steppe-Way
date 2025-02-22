package controllers

import (
	"diplomaPorject/backend/blogs_service/internal/models"
	"diplomaPorject/backend/blogs_service/utils/db"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

func CreateBlog(w http.ResponseWriter, r *http.Request) {
	userIDValue := r.Context().Value("user_id")
	if userIDValue == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, ok := userIDValue.(uint)
	if !ok {
		http.Error(w, "Invalid user ID", http.StatusInternalServerError)
		return
	}

	var blog models.Blog
	if err := json.NewDecoder(r.Body).Decode(&blog); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	blog.UserID = userID

	if err := validateBlog(&blog); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := db.DB.Create(&blog).Error; err != nil {
		log.Printf("Error creating blog: %v", err)
		http.Error(w, "Failed to create blog", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(blog)
}

func GetBlogs(w http.ResponseWriter, r *http.Request) {
	var blogs []models.Blog

	query := db.DB.Preload("Comments")

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
	query.Model(&models.Blog{}).Count(&totalCount)

	if err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&blogs).Error; err != nil {
		http.Error(w, "Failed to fetch blogs", http.StatusInternalServerError)
		return
	}

	response := struct {
		Blogs    []models.Blog `json:"blogs"`
		Total    int64         `json:"total"`
		Page     int           `json:"page"`
		PageSize int           `json:"page_size"`
	}{
		Blogs:    blogs,
		Total:    totalCount,
		Page:     page,
		PageSize: pageSize,
	}

	json.NewEncoder(w).Encode(response)
}

func GetBlog(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid blog ID", http.StatusBadRequest)
		return
	}

	var blog models.Blog
	if err := db.DB.Preload("Comments").First(&blog, id).Error; err != nil {
		http.Error(w, "Blog not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(blog)
}

func UpdateBlog(w http.ResponseWriter, r *http.Request) {
	userIDValue := r.Context().Value("user_id")
	if userIDValue == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, ok := userIDValue.(uint)
	if !ok {
		http.Error(w, "Invalid user ID", http.StatusInternalServerError)
		return
	}

	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid blog ID", http.StatusBadRequest)
		return
	}

	var existingBlog models.Blog
	if err := db.DB.First(&existingBlog, id).Error; err != nil {
		http.Error(w, "Blog not found", http.StatusNotFound)
		return
	}

	if existingBlog.UserID != userID {
		http.Error(w, "Unauthorized to update this blog", http.StatusForbidden)
		return
	}

	var updatedBlog models.Blog
	if err := json.NewDecoder(r.Body).Decode(&updatedBlog); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	updatedBlog.ID = existingBlog.ID
	updatedBlog.UserID = existingBlog.UserID

	if err := validateBlog(&updatedBlog); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := db.DB.Save(&updatedBlog).Error; err != nil {
		log.Printf("Error updating blog: %v", err)
		http.Error(w, "Failed to update blog", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(updatedBlog)
}

func DeleteBlog(w http.ResponseWriter, r *http.Request) {
	userIDValue := r.Context().Value("user_id")
	if userIDValue == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, ok := userIDValue.(uint)
	if !ok {
		http.Error(w, "Invalid user ID", http.StatusInternalServerError)
		return
	}

	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid blog ID", http.StatusBadRequest)
		return
	}

	var existingBlog models.Blog
	if err := db.DB.First(&existingBlog, id).Error; err != nil {
		http.Error(w, "Blog not found", http.StatusNotFound)
		return
	}

	if existingBlog.UserID != userID {
		http.Error(w, "Unauthorized to delete this blog", http.StatusForbidden)
		return
	}

	if err := db.DB.Delete(&models.Blog{}, id).Error; err != nil {
		log.Printf("Error deleting blog: %v", err)
		http.Error(w, "Failed to delete blog", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func validateBlog(blog *models.Blog) error {
	if blog.Title == "" {
		return fmt.Errorf("title is required")
	}
	if len(blog.Title) > 255 {
		return fmt.Errorf("title is too long")
	}
	if blog.Content == "" {
		return fmt.Errorf("content is required")
	}
	return nil
}

func validateComment(comment *models.Comment) error {
	if comment.Content == "" {
		return fmt.Errorf("comment content is required")
	}
	if len(comment.Content) > 500 {
		return fmt.Errorf("comment is too long")
	}
	return nil
}
