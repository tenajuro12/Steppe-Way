package main

import (
	"favorites_service/internal/routes"
	"favorites_service/utils/db"
	"log"
	"net/http"
)

func main() {
	db.ConnectDB()
	router := routes.SetupRoutes()
	log.Println("Favorite service running on port 8088...")
	log.Fatal(http.ListenAndServe(":8088", router))
}
