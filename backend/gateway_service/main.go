package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	middlewares "gateway_service/middleware"
	"github.com/gorilla/mux"
)

type ServiceConfig struct {
	URL   string
	Paths []string
	Auth  bool
}

var services = map[string]ServiceConfig{
	"blog": {
		URL:   "http://blogs-service:8081",
		Paths: []string{"/blogs"},
		Auth:  true,
	},
	"auth": {
		URL: "http://auth-service:8082",
		Paths: []string{
			"/login",
			"/register",
			"/profile",
			"/validate-admin",
			"/validate-session",
		},
		Auth: false,
	},
	"events": {
		URL: "http://events-service:8083",
		Paths: []string{
			"/admin/events",
			"/events",
			"/uploads/events",
		},
		Auth: false,
	},

	"attractions": {
		URL: "http://attraction-service:8085",
		Paths: []string{
			"/admin/attractions",
			"/attractions",
			"/uploads",
		},
		Auth: false,
	},
}

var pathAuthOverrides = map[string]bool{
	"/admin/events":      true,
	"/admin/attractions": true,
	"/attractions":       true,
}

func main() {
	r := mux.NewRouter()

	setupRoutes(r)

	handler := middlewares.CorsMiddleware(r)

	log.Println("Gateway service running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

func setupRoutes(r *mux.Router) {
	for _, config := range services {
		for _, path := range config.Paths {
			handler := createProxyHandler(config.URL)

			requiresAuth := config.Auth
			if override, exists := pathAuthOverrides[path]; exists {
				requiresAuth = override
			}

			if requiresAuth {
				handler = middlewares.AuthMiddleware(handler)
			}

			r.PathPrefix(path).Handler(handler)
		}
	}
}

func createProxyHandler(targetServiceURL string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		target, err := url.Parse(targetServiceURL)
		if err != nil {
			http.Error(w, "Invalid target URL", http.StatusInternalServerError)
			return
		}

		log.Printf("Proxying request to: %s%s", target.String(), r.URL.Path)

		proxy := httputil.NewSingleHostReverseProxy(target)

		proxy.ModifyResponse = func(response *http.Response) error {
			if response.StatusCode >= 300 && response.StatusCode < 400 {
				location := response.Header.Get("Location")
				if strings.Contains(location, target.Host) {
					response.Header.Set("Location", strings.Replace(location, target.Host, "localhost:8080", 1))
				}
			}
			return nil
		}

		r.Host = target.Host
		proxy.ServeHTTP(w, r)
	})
}
