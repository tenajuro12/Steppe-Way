package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
	"io"
	"log"
	"net/http"
	"os"
	"profile_service/internal/db"
	"profile_service/internal/models"
	"profile_service/utils"
	"strconv"
)

const DefaultProfileImageURL = "/uploads/users/default_user.jpg"

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
	userID := vars["user_id"]

	var profile models.Profile
	if err := db.DB.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Profile not found", http.StatusNotFound)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	email := r.FormValue("email")
	bio := r.FormValue("bio")

	if username != "" {
		profile.Username = username
	}
	if email != "" {
		profile.Email = email
	}
	if bio != "" {
		profile.Bio = bio
	}

	file, handler, err := r.FormFile("image")
	if err == nil {
		defer file.Close()
		filename := strconv.FormatInt(int64(profile.UserID), 10) + "_" + handler.Filename
		savePath := "uploads/users/" + filename

		dst, err := createOrReplaceFile(savePath)
		if err != nil {
			http.Error(w, "Unable to save image", http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		_, err = dst.ReadFrom(file)
		if err != nil {
			http.Error(w, "Error saving file", http.StatusInternalServerError)
			return
		}

		profile.ProfileImg = "/" + savePath
	}

	if err := db.DB.Save(&profile).Error; err != nil {
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	if username != "" || email != "" {
		if err := updateAuthService(userID, username, email); err != nil {
			http.Error(w, "Auth service update failed", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	go utils.NotifyBlogsService(profile.UserID, profile.Username)
	json.NewEncoder(w).Encode(profile)
}

func createOrReplaceFile(path string) (*os.File, error) {
	_ = os.MkdirAll("uploads/users", os.ModePerm)
	return os.Create(path)
}

func updateAuthService(userID, username, email string) error {
	authServiceURL := "http://auth-service:8082/update-user"

	updateData := map[string]string{
		"user_id": userID,
	}

	if username != "" {
		updateData["username"] = username
	}
	if email != "" {
		updateData["email"] = email
	}

	data, _ := json.Marshal(updateData)

	req, err := http.NewRequest("PATCH", authServiceURL, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Printf("auth-service PATCH response %d: %s", resp.StatusCode, string(body))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.New("auth-service update failed")
	}

	return nil
}
