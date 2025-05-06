package controllers

import (
	"diplomaPorject/backend/blogs_service/internal/models"
	"diplomaPorject/backend/blogs_service/utils"
	"diplomaPorject/backend/blogs_service/utils/db"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

func AddComment(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

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
	blogID, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid blog ID", http.StatusBadRequest)
		return
	}

	content := r.FormValue("content")
	if content == "" {
		http.Error(w, "Comment content is required", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["images"]
	imageURLs, err := uploadImages(files)
	if err != nil {
		http.Error(w, fmt.Sprintf("Image upload failed: %v", err), http.StatusInternalServerError)
		return
	}
	username, err := utils.GetUsername(userID)
	if err != nil {
		username = "unknown"
	}

	comment := models.Comment{
		Content:  content,
		BlogID:   uint(blogID),
		UserID:   userID,
		Username: username,
	}

	if err := db.DB.Create(&comment).Error; err != nil {
		http.Error(w, "Failed to add comment", http.StatusInternalServerError)
		return
	}

	for _, url := range imageURLs {
		cimg := models.CommentImage{CommentID: comment.ID, URL: url}
		db.DB.Create(&cimg)
	}

	db.DB.Preload("Images").First(&comment, comment.ID)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(comment)
}

func GetComments(w http.ResponseWriter, r *http.Request) {
	blogID, _ := strconv.Atoi(mux.Vars(r)["id"])
	var comments []models.Comment
	if err := db.DB.
		Where("blog_id = ?", blogID).
		Preload("Images").
		Order("created_at ASC").
		Find(&comments).Error; err != nil {
		http.Error(w, "Failed to fetch comments", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(comments)
}

func UpdateComment(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	commentID, _ := strconv.Atoi(mux.Vars(r)["comment_id"])

	userID, _ := r.Context().Value("user_id").(uint)

	var comment models.Comment
	if err := db.DB.Preload("Images").First(&comment, commentID).Error; err != nil {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}
	if comment.UserID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if txt := r.FormValue("content"); txt != "" {
		comment.Content = txt
	}

	if files := r.MultipartForm.File["images"]; len(files) > 0 {
		db.DB.Where("comment_id = ?", comment.ID).Delete(&models.CommentImage{})
		urls, err := uploadImages(files)
		if err != nil {
			http.Error(w, "Image upload error", http.StatusInternalServerError)
			return
		}
		for _, u := range urls {
			db.DB.Create(&models.CommentImage{CommentID: comment.ID, URL: u})
		}
	}

	db.DB.Save(&comment)
	db.DB.Preload("Images").First(&comment, comment.ID)
	json.NewEncoder(w).Encode(comment)
}

func DeleteComment(w http.ResponseWriter, r *http.Request) {
	commentID, _ := strconv.Atoi(mux.Vars(r)["comment_id"])
	userID, _ := r.Context().Value("user_id").(uint)

	var comment models.Comment
	if err := db.DB.First(&comment, commentID).Error; err != nil {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}
	if comment.UserID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	db.DB.Where("comment_id = ?", comment.ID).Delete(&models.CommentImage{})
	db.DB.Delete(&comment)
	w.WriteHeader(http.StatusNoContent)
}
