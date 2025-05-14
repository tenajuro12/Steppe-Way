package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Incoming request to: %s", r.URL.Path)

		cookie, err := r.Cookie("session_token")
		if err != nil {
			log.Printf("No session_token cookie found: %v", err)
			http.Error(w, "Unauthorized - No session token", http.StatusUnauthorized)
			return
		}
		log.Printf("Found session token: %s", cookie.Value)

		authServiceURL := "http://auth-service:8082/validate-session"
		req, err := http.NewRequest("GET", authServiceURL, nil)
		if err != nil {
			log.Printf("Error creating validation request: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		req.Header.Set("Cookie", r.Header.Get("Cookie"))

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Error calling auth service: %v", err)
			http.Error(w, "Unauthorized - Auth service error", http.StatusUnauthorized)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("Auth service returned non-200 status: %d", resp.StatusCode)
			http.Error(w, fmt.Sprintf("Unauthorized - Auth service returned %d", resp.StatusCode), http.StatusUnauthorized)
			return
		}

		body, _ := io.ReadAll(resp.Body)
		var authResponse map[string]interface{}
		json.Unmarshal(body, &authResponse)

		userID, exists := authResponse["user_id"].(float64)
		if !exists {
			log.Println("Auth service response did not contain user_id")
			http.Error(w, "Unauthorized - Invalid session", http.StatusUnauthorized)
			return
		}

		username, usernameExists := authResponse["username"].(string)
		if !usernameExists {
			log.Println("Auth service response did not contain username")
			http.Error(w, "Unauthorized - Invalid session", http.StatusUnauthorized)
			return
		}

		isAdmin := false
		if adminValue, ok := authResponse["is_admin"].(bool); ok {
			isAdmin = adminValue
		}

		r.Header.Set("X-User-ID", strconv.Itoa(int(userID)))
		r.Header.Set("X-Username", username)
		r.Header.Set("X-Is-Admin", strconv.FormatBool(isAdmin))

		ctx := context.WithValue(r.Context(), "user_id", uint(userID))
		ctx = context.WithValue(ctx, "username", username)
		ctx = context.WithValue(ctx, "is_admin", isAdmin)
		r = r.WithContext(ctx)

		log.Printf("Authentication successful: user_id=%v username=%v is_admin=%v", int(userID), username, isAdmin)
		next.ServeHTTP(w, r)
	})
}

func AdminAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Incoming admin request to: %s", r.URL.Path)

		_, err := r.Cookie("session_token")
		if err != nil {
			log.Printf("No session_token cookie found: %v", err)
			http.Error(w, "Unauthorized - No session token", http.StatusUnauthorized)
			return
		}

		authServiceURL := "http://auth-service:8082/validate-admin"
		req, err := http.NewRequest("GET", authServiceURL, nil)
		if err != nil {
			log.Printf("Error creating admin validation request: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		req.Header.Set("Cookie", r.Header.Get("Cookie"))

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Error calling auth service for admin validation: %v", err)
			http.Error(w, "Unauthorized - Auth service error", http.StatusUnauthorized)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("Admin validation service returned non-200 status: %d", resp.StatusCode)
			http.Error(w, "Unauthorized - Not an admin", http.StatusUnauthorized)
			return
		}

		body, _ := io.ReadAll(resp.Body)
		var authResponse map[string]interface{}
		json.Unmarshal(body, &authResponse)

		adminID, exists := authResponse["admin_id"].(float64)
		if !exists {
			log.Println("Admin validation response did not contain admin_id")
			http.Error(w, "Unauthorized - Invalid admin session", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "admin_id", uint(adminID))
		r = r.WithContext(ctx)

		log.Printf("Admin authentication successful: admin_id=%v", int(adminID))
		next.ServeHTTP(w, r)
	})
}
