package routes

import (
	"favorites_service/internal/controller"
	"net/http"

	"github.com/gorilla/mux"
)

func SetupRoutes() http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Healthy"))
	}).Methods("GET")

	favHandler := handlers.NewFavoriteHandler()

	r.HandleFunc("/favorites", favHandler.AddFavorite).Methods("POST")
	r.HandleFunc("/favorites", favHandler.GetUserFavorites).Methods("GET")
	r.HandleFunc("/favorites/{id:[0-9]+}", favHandler.GetFavorite).Methods("GET")
	r.HandleFunc("/favorites/{type}/{id:[0-9]+}", favHandler.RemoveFavorite).Methods("DELETE")
	r.HandleFunc("/favorites/check/{type}/{id:[0-9]+}", favHandler.CheckFavorite).Methods("GET")

	return r
}
