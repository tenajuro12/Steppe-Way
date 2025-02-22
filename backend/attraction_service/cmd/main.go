package main

import (
	"diplomaPorject/backend/attraction/internal/routes"
	"diplomaPorject/backend/attraction/utils/db"
	"log"
	"net/http"
)

func main() {
	db.ConnectDB()
	router := routes.SetupRoutes()
	log.Println("Attraction service running on port 8085...")
	log.Fatal(http.ListenAndServe(":8085", router))
}
