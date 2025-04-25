package main

import (
	"log"
	"net/http"
	"review_service/internal/handlers"
	"review_service/utils"
)

func main() {
	utils.ConnectDB()

	http.HandleFunc("/reviews", handlers.ReviewRouter)
	http.HandleFunc("/reviews/", handlers.ReviewByIDRouter)
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

	log.Println("Review service running on :8086")
	log.Fatal(http.ListenAndServe(":8086", nil))
}
