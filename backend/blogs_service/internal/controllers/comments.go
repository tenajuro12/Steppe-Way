package controllers

import (
	"diplomaPorject/backend/blogs_service/internal/models"
	"diplomaPorject/backend/blogs_service/utils/db"
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

func AddComment(w http.ResponseWriter, r *http.Request) {
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

	var blog models.Blog
	if err := db.DB.First(&blog, id).Error; err != nil {
		http.Error(w, "Blog not found", http.StatusNotFound)
		return
	}

	var comment models.Comment
	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	comment.BlogID = uint(id)
	comment.UserID = userID

	if err := validateComment(&comment); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := db.DB.Create(&comment).Error; err != nil {
		log.Printf("Error adding comment: %v", err)
		http.Error(w, "Failed to add comment", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(comment)
}
