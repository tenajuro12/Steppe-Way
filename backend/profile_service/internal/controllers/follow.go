package controllers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"profile_service/internal/db"
	"profile_service/internal/models"
	"strconv"
)

func FollowUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	followeeID, err := strconv.Atoi(vars["user_id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	followerID, ok := r.Context().Value("user_id").(uint)
	if !ok || followerID == 0 {
		http.Error(w, "Unauthorized - No authenticated user", http.StatusUnauthorized)
		return
	}

	if followerID == uint(followeeID) {
		http.Error(w, "You cannot follow yourself", http.StatusBadRequest)
		return
	}

	follow := models.Follow{FollowerID: uint(followerID), FolloweeID: uint(followeeID)}

	var existing models.Follow
	if err := db.DB.Where("follower_id = ? AND followee_id = ?", followerID, followeeID).First(&existing).Error; err == nil {
		http.Error(w, "You are already following this user", http.StatusConflict)
		return
	}
	if err := db.DB.Create(&follow).Error; err != nil {
		http.Error(w, "Failed to follow user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Successfully followed user"})

}
