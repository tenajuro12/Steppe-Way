package routes

import (
	"diplomaPorject/backend/events_service/internal/controllers"
	"diplomaPorject/backend/events_service/internal/middleware"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func SetupRoutes() *mux.Router {
	r := mux.NewRouter()
	uploadDir := "/app/uploads/events"

	log.Printf("Serving static files from: %s", uploadDir)

	r.PathPrefix("/uploads/events/").Handler(
		http.StripPrefix("/uploads/events/", http.FileServer(http.Dir(uploadDir))),
	).Methods("GET")

	admin := r.PathPrefix("/admin/events").Subrouter()
	admin.Use(middleware.AdminAuthMiddleware)
	admin.HandleFunc("", controllers.CreateEvent).Methods("POST")
	admin.HandleFunc("", controllers.ListEvents).Methods("GET")
	admin.HandleFunc("/{id}", controllers.UpdateEvent).Methods("PUT")
	admin.HandleFunc("/{id}", controllers.DeleteEvent).Methods("DELETE")
	admin.HandleFunc("/{id}/publish", controllers.PublishEvent).Methods("POST")
	admin.HandleFunc("/{id}/unpublish", controllers.UnpublishEvent).Methods("POST")

	r.HandleFunc("/events", controllers.ListPublishedEvents).Methods("GET")
	r.HandleFunc("/events/{id}", controllers.GetEvent).Methods("GET")

	return r
}
