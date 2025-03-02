package controllers

import (
	"authorization_service/internal/model"
	"authorization_service/utils/db"
	"authorization_service/utils/hashing"
	utils "authorization_service/utils/session"
	"bytes"
	"encoding/json"
	"errors"
	"gorm.io/gorm"
	"log"
	"net/http"
	"time"
)

const DefaultProfileImageURL = "../../../backend/uploads/users"

func Register(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Error with decoding", http.StatusBadRequest)
		return
	}

	hashedPassword, err := hashing.HashPassword(creds.Password)
	if err != nil {
		http.Error(w, "Error with hashing password", http.StatusInternalServerError)
		return
	}

	user := model.User{
		Username: creds.Username,
		Email:    creds.Email,
		Password: hashedPassword,
	}

	if err := db.DB.Create(&user).Error; err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	createDefaultProfile(user.ID, user.Username, user.Email)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
}
func createDefaultProfile(userID uint, username, email string) {
	profileServiceURL := "http://profile-service:8084/profiles"

	profileData := map[string]interface{}{
		"user_id":     userID,
		"username":    username,
		"email":       email,
		"bio":         "I just joined!",
		"profile_img": DefaultProfileImageURL,
	}

	data, err := json.Marshal(profileData)
	if err != nil {
		log.Printf("[Auth Service] Error marshaling profile data: %v", err)
		return
	}

	log.Printf("[Auth Service] Sending profile creation request: %s", string(data)) // <-- NEW LOG

	resp, err := http.Post(profileServiceURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Printf("[Auth Service] Error calling profile service: %v", err) // <-- NEW LOG
		return
	}
	defer resp.Body.Close()

	log.Printf("[Auth Service] Profile service returned status: %d", resp.StatusCode) // <-- NEW LOG

	if resp.StatusCode != http.StatusCreated {
		log.Printf("[Auth Service] Profile service returned an unexpected status: %d", resp.StatusCode)
	}
}

func Login(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	json.NewDecoder(r.Body).Decode(&creds)
	var user model.User
	if err := db.DB.Where("email=?", creds.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	if err := hashing.CheckPassword(user.Password, creds.Password); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := utils.CreateSession(w, r, user.ID); err != nil {
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Login successful"})

}

func Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Error(w, "No session found", http.StatusUnauthorized)
		return
	}

	if err := db.DB.Where("token = ?", cookie.Value).Delete(&model.Session{}).Error; err != nil {
		http.Error(w, "Failed to delete session", http.StatusInternalServerError)
		return
	}

	cookie.MaxAge = -1
	http.SetCookie(w, cookie)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Logged out successfully"))
}

func GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, authenticated := utils.GetSessionUserID(r)
	if !authenticated {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var user model.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(user)
}

func ValidateSession(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var session model.Session
	if err := db.DB.Where("token = ? AND expires_at > ?",
		cookie.Value, time.Now()).First(&session).Error; err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]uint{
		"user_id": session.UserID,
	})
}
func ValidateAdmin(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Error(w, "No authorization token", http.StatusUnauthorized)
		return
	}

	var session model.Session
	if err := db.DB.Where("token = ?", cookie.Value).First(&session).Error; err != nil {
		http.Error(w, "Unauthorized token", http.StatusUnauthorized)
		return
	}

	var user model.User
	if err := db.DB.First(&user, session.UserID).Error; err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	if !user.IsAdmin {
		http.Error(w, "Forbidden", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]uint{
		"admin_id": user.ID,
	})
	w.WriteHeader(http.StatusOK)
}
