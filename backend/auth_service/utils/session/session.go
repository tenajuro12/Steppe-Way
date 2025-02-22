package utils

import (
	"authorization_service/internal/model"
	"authorization_service/utils/db"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"
)

func generateSessionToken() string {
	bytes := make([]byte, 32)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func CreateSession(w http.ResponseWriter, r *http.Request, userID uint) error {
	sessionToken := generateSessionToken()
	expiration := time.Now().Add(24 * time.Hour)

	// Save session to database
	session := model.Session{
		UserID:    userID,
		Token:     sessionToken,
		ExpiresAt: expiration,
	}
	if err := db.DB.Create(&session).Error; err != nil {
		return err
	}

	// Set session token in a secure HttpOnly cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Expires:  expiration,
		HttpOnly: true,
		Path:     "/",
	})
	return nil
}

func GetSessionUserID(r *http.Request) (uint, bool) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return 0, false
	}

	var session model.Session
	if err := db.DB.Where("token = ? AND expires_at > ?",
		cookie.Value, time.Now()).First(&session).Error; err != nil {
		return 0, false
	}

	return session.UserID, true
}

func DestroySession(w http.ResponseWriter, r *http.Request) error {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return nil // No session to destroy
	}

	// Remove session from database
	if err := db.DB.Where("token = ?", cookie.Value).Delete(&model.Session{}).Error; err != nil {
		return err
	}

	// Clear the cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Path:     "/",
	})
	return nil
}
