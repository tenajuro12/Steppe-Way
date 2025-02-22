package main

import (
	"diplomaPorject/backend/blogs_service/internal/routes"
	"diplomaPorject/backend/blogs_service/utils/db"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	// Connect to Database
	db.ConnectDB()

	// Initialize Router
	r := mux.NewRouter()

	// Register Routes
	routes.RegisterBlogRoutes(r)

	// Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	fmt.Println("Blog service running on port:", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
