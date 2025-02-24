package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func BlogsAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Blog authentication for: %s", r.URL.Path)

		authServiceURL := os.Getenv("AUTH_SERVICE_URL")
		if authServiceURL == "" {
			authServiceURL = "http://auth-service:8082"
		}
		validateURL := fmt.Sprintf("%s/validate-session", authServiceURL)

		req, err := http.NewRequest("GET", validateURL, nil)
		if err != nil {
			log.Printf("Error creating session validation request: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		req.Header.Set("Cookie", r.Header.Get("Cookie"))

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Error calling auth service for session validation: %v", err)
			http.Error(w, "Unauthorized - Auth service error", http.StatusUnauthorized)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("Session validation failed. Status: %d", resp.StatusCode)
			http.Error(w, "Unauthorized - Invalid session", http.StatusUnauthorized)
			return
		}

		// Read and parse the response body to get user ID
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error reading response body: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		var userResp struct {
			UserID uint `json:"user_id"`
		}
		if err := json.Unmarshal(body, &userResp); err != nil {
			log.Printf("Error parsing user response: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), "user_id", userResp.UserID)
		r = r.WithContext(ctx)

		log.Printf("Blog authentication successful. User ID: %d", userResp.UserID)
		next.ServeHTTP(w, r)
	})
}
