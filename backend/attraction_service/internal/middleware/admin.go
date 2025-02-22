package middleware

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

type AdminResponse struct {
	AdminID uint `json:"admin_id"`
}

func AdminAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Admin authentication for: %s", r.URL.Path)

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
			log.Printf("Admin validation failed. Status: %d", resp.StatusCode)
			http.Error(w, "Unauthorized - Admin access denied", http.StatusUnauthorized)
			return
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error reading response body: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		var adminResp AdminResponse
		if err := json.Unmarshal(body, &adminResp); err != nil {
			log.Printf("Error parsing admin response: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), "admin_id", adminResp.AdminID)

		r = r.WithContext(ctx)

		log.Printf("Admin authentication successful. Admin ID: %d", adminResp.AdminID)
		next.ServeHTTP(w, r)
	})
}
