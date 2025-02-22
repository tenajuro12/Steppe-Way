package main

import (
	"authorization_service/internal/routes"
	"authorization_service/utils/db"
	"fmt"
	"log"
	"net/http"
)

func main() {
	db.ConnectDB()
	if db.DB == nil {
		log.Fatal("Database connection is nil!")
	}

	r := routes.SetupRoutes()

	fmt.Println("Server running on port:", 8082)
	log.Fatal(http.ListenAndServe(":"+"8082", r))
}
