package middlewares

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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

		body, _ := ioutil.ReadAll(resp.Body)
		var authResponse map[string]interface{}
		json.Unmarshal(body, &authResponse)

		userID, exists := authResponse["user_id"].(float64)
		if !exists {
			log.Println("Auth service response did not contain user_id")
			http.Error(w, "Unauthorized - Invalid session", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "user_id", uint(userID))
		r = r.WithContext(ctx)

		log.Printf("Authentication successful, forwarding to blogs service")
		next.ServeHTTP(w, r)
	})
}
