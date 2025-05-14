package routes

import (
	"accommodation_service/internal/controllers"
	"accommodation_service/internal/middleware"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func SetupRoutes() *mux.Router {
	r := mux.NewRouter()

	uploadDir := "/app/uploads/accommodations"
	log.Printf("Serving static files from: %s", uploadDir)
	r.PathPrefix("/uploads/accommodations/").Handler(
		http.StripPrefix("/uploads/accommodations/", http.FileServer(http.Dir(uploadDir))),
	).Methods("GET")

	accommodationController := controllers.AccommodationController{}

	r.HandleFunc("/accommodations", accommodationController.ListAccommodations).Methods("GET")
	r.HandleFunc("/accommodations/{id:[0-9]+}", accommodationController.GetAccommodation).Methods("GET")
	r.HandleFunc("/accommodations/{id:[0-9]+}/reviews", accommodationController.GetAccommodationReviews).Methods("GET")

	admin := r.PathPrefix("/admin/accommodations").Subrouter()
	admin.Use(middleware.AdminAuthMiddleware)
	admin.HandleFunc("", accommodationController.CreateAccommodation).Methods("POST")
	admin.HandleFunc("", accommodationController.ListAdminAccommodations).Methods("GET")
	admin.HandleFunc("/{id:[0-9]+}", accommodationController.UpdateAccommodation).Methods("PUT")
	admin.HandleFunc("/{id:[0-9]+}", accommodationController.DeleteAccommodation).Methods("DELETE")
	admin.HandleFunc("/room-types/{room_id:[0-9]+}", accommodationController.DeleteRoomType).Methods("DELETE")
	admin.HandleFunc("/{id:[0-9]+}/publish", accommodationController.PublishAccommodation).Methods("POST")
	admin.HandleFunc("/{id:[0-9]+}/unpublish", accommodationController.UnpublishAccommodation).Methods("POST")
	admin.HandleFunc("/room-types/{room_id:[0-9]+}/images", accommodationController.UploadRoomTypeImages).Methods("POST")
	user := r.PathPrefix("/user/accommodations").Subrouter()
	user.Use(middleware.AuthMiddleware)
	user.HandleFunc("/{id:[0-9]+}/reviews", accommodationController.AddReview).Methods("POST")
	user.HandleFunc("/reviews/{review_id:[0-9]+}", accommodationController.UpdateReview).Methods("PUT")
	user.HandleFunc("/reviews/{review_id:[0-9]+}", accommodationController.DeleteReview).Methods("DELETE")

	return r
}
