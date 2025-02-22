package main

import (
	routes1 "diplomaPorject/backend/events_service/internal/routes"
	"diplomaPorject/backend/events_service/utils/db"
	"log"
	"net/http"
)

func main() {
	db.ConnectDB()
	router := routes1.SetupRoutes()
	log.Println("Events service running on port 8083...")
	log.Fatal(http.ListenAndServe(":8083", router))
}
