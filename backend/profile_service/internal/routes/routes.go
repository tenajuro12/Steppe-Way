package routes

import (
	"github.com/gorilla/mux"
	"profile_service/internal/controllers"
)

func SetupRoutes(r *mux.Router) {
	r.HandleFunc("/profiles", controllers.CreateProfile).Methods("POST")
	r.HandleFunc("/profiles/{user_id}", controllers.GetProfile).Methods("GET")
	r.HandleFunc("/profiles/{user_id}", controllers.UpdateProfile).Methods("PATCH")

}
