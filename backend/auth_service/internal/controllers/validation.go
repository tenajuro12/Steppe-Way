package controllers

import (
	"authorization_service/internal/model"
	"authorization_service/utils/db"
	"encoding/json"
	"net/http"
	"time"
)

func ValidateSession(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var session model.Session
	if err := db.DB.Where("token = ? AND expires_at > ?", cookie.Value, time.Now()).First(&session).Error; err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 👇 Загрузим пользователя по session.UserID
	var user model.User
	if err := db.DB.First(&user, session.UserID).Error; err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// ✅ Вернём всё, что нужно gateway'ю
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id":  user.ID,
		"username": user.Username,
		"is_admin": user.IsAdmin,
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
