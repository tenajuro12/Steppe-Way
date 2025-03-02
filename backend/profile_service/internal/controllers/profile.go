package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
	"log"
	"net/http"
	"profile_service/internal/db"
	"profile_service/internal/models"
	"strconv"
)

const DefaultProfileImageURL = "../../../backend/uploads/users"

func CreateProfile(w http.ResponseWriter, r *http.Request) {
	var p models.Profile
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if p.UserID == 0 {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	var existing models.Profile
	err := db.DB.Where("user_id = ?", p.UserID).First(&existing).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		http.Error(w, "Error checking for existing profile", http.StatusInternalServerError)
		return
	}
	if err == nil {
		http.Error(w, "Profile already exists for this user", http.StatusConflict)
		return
	}

	if p.Bio == "" {
		p.Bio = "I just joined!"
	}
	if p.ProfileImg == "" {
		p.ProfileImg = DefaultProfileImageURL
	}

	if err := db.DB.Create(&p).Error; err != nil {
		log.Printf("Failed to create profile for user_id %d: %v", p.UserID, err)
		http.Error(w, "Failed to create profile", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}

func GetProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["user_id"]
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var profile models.Profile
	if err := db.DB.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

func UpdateProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["user_id"]
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var input struct {
		Username   string `json:"username"`
		Email      string `json:"email"`
		Bio        string `json:"bio"`
		ProfileImg string `json:"profile_img"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if input.Username != "" || input.Email != "" {
		authUpdate := make(map[string]string)
		if input.Username != "" {
			authUpdate["username"] = input.Username
		}
		if input.Email != "" {
			authUpdate["email"] = input.Email
		}
		payloadBytes, err := json.Marshal(authUpdate)
		if err != nil {
			http.Error(w, "Failed to marshal auth update payload", http.StatusInternalServerError)
			return
		}

		authURL := fmt.Sprintf("http://auth-service:8082/users/%d", userID)
		req, err := http.NewRequest("PATCH", authURL, bytes.NewBuffer(payloadBytes))
		if err != nil {
			http.Error(w, "Failed to create request to auth service", http.StatusInternalServerError)
			return
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "Failed to update user info in auth service", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			http.Error(w, "Auth service update failed", http.StatusInternalServerError)
			return
		}
	}

	var profile models.Profile
	if err := db.DB.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}

	if input.Username != "" {
		profile.Username = input.Username
	}
	if input.Email != "" {
		profile.Email = input.Email
	}
	if input.Bio != "" {
		profile.Bio = input.Bio
	}
	if input.ProfileImg != "" {
		profile.ProfileImg = input.ProfileImg
	}

	if err := db.DB.Save(&profile).Error; err != nil {
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}
