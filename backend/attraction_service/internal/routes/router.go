package routes

import (
	"diplomaPorject/backend/attraction/internal/controllers"
	"diplomaPorject/backend/attraction/internal/middleware"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func SetupRoutes() *mux.Router {
	r := mux.NewRouter()

	uploadDir := "/app/uploads"

	log.Printf("Serving static files from: %s", uploadDir)

	r.PathPrefix("/uploads/").Handler(
		http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadDir))),
	).Methods("GET")

	admin := r.PathPrefix("/admin/attractions").Subrouter()
	admin.Use(middleware.AdminAuthMiddleware)
	admin.HandleFunc("", controllers.CreateAttraction).Methods("POST")
	admin.HandleFunc("", controllers.ListAttractions).Methods("GET")
	admin.HandleFunc("/{id}", controllers.GetAttraction).Methods("GET")
	admin.HandleFunc("/{id}", controllers.UpdateAttraction).Methods("PUT")
	admin.HandleFunc("/{id}", controllers.DeleteAttraction).Methods("DELETE")
	admin.HandleFunc("/{id}/publish", controllers.PublishAttraction).Methods("POST")
	admin.HandleFunc("/{id}/unpublish", controllers.UnpublishAttraction).Methods("POST")

	r.HandleFunc("/attractions", controllers.ListPublishedAttractions).Methods("GET")
	return r
}
