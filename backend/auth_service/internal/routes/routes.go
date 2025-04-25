package routes

import (
	"authorization_service/internal/controllers"
	"authorization_service/middleware"
	"github.com/gorilla/mux"
)

func SetupRoutes() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/register", controllers.Register).Methods("POST")
	r.HandleFunc("/login", controllers.Login).Methods("POST")
	r.HandleFunc("/validate-session", controllers.ValidateSession).Methods("GET")
	r.HandleFunc("/validate-admin", controllers.ValidateAdmin).Methods("GET")

	protected := r.PathPrefix("/").Subrouter()
	protected.Use(middleware.AuthMiddleware)
	protected.HandleFunc("/profile", controllers.GetProfile).Methods("GET")
	protected.HandleFunc("/profile", controllers.Logout).Methods("POST")
	r.HandleFunc("/update-user", controllers.UpdateUser).Methods("PATCH")

	return r
}
