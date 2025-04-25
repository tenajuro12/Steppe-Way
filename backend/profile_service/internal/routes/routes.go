package routes

import (
	"github.com/gorilla/mux"
	"net/http"
	"profile_service/internal/controllers"
)

func SetupRoutes(r *mux.Router) {
	r.HandleFunc("/user/profiles", controllers.CreateProfile).Methods("POST")
	r.HandleFunc("/user/profiles/{user_id}", controllers.GetProfile).Methods("GET")
	r.HandleFunc("/user/profiles/{user_id}", controllers.UpdateProfile).Methods("PATCH")
	r.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))
}
