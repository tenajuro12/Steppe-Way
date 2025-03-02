package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"profile_service/internal/db"
	"profile_service/internal/routes"
)

func main() {
	db.InitDB()

	r := mux.NewRouter()
	routes.SetupRoutes(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8084"
	}

	log.Printf("Profile Service running on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
