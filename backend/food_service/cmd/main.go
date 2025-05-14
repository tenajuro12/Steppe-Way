package main

import (
	"food_service/routes"
	"food_service/utils/db"
	"log"
	"net/http"
	"os"
)

func main() {
	db.ConnectDB()

	router := routes.SetupRoutes()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}

	log.Printf("Food service running on port %s", port)

	log.Fatal(http.ListenAndServe(":"+port, router))
}
