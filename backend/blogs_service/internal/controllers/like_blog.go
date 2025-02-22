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

func LikeBlog(w http.ResponseWriter, r *http.Request) {
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

	var existingLike models.BlogLike
	if result := db.DB.Where("user_id = ? AND blog_id = ?", userID, blog.ID).First(&existingLike); result.Error == nil {
		http.Error(w, "You have already liked this blog", http.StatusBadRequest)
		return
	}

	like := models.BlogLike{
		UserID: userID,
		BlogID: blog.ID,
	}
	if err := db.DB.Create(&like).Error; err != nil {
		log.Printf("Error creating like: %v", err)
		http.Error(w, "Failed to like blog", http.StatusInternalServerError)
		return
	}

	blog.Likes++
	if err := db.DB.Save(&blog).Error; err != nil {
		log.Printf("Error updating blog likes: %v", err)
	}

	json.NewEncoder(w).Encode(blog)
}
