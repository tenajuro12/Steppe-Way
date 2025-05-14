package routes

import (
	"food_service/internal/controllers"
	"food_service/internal/middleware"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func SetupRoutes() *mux.Router {
	r := mux.NewRouter()

	uploadDir := "/app/uploads/food"
	log.Printf("Serving static files from: %s", uploadDir)
	r.PathPrefix("/uploads/food/").Handler(
		http.StripPrefix("/uploads/food/", http.FileServer(http.Dir(uploadDir))),
	).Methods("GET")

	foodController := controllers.FoodController{}

	r.HandleFunc("/places", foodController.ListPlaces).Methods("GET")
	r.HandleFunc("/places/{id:[0-9]+}", foodController.GetPlace).Methods("GET")
	r.HandleFunc("/places/{id:[0-9]+}/reviews", foodController.GetPlaceReviews).Methods("GET")
	r.HandleFunc("/places/search", foodController.SearchPlaces).Methods("GET")
	r.HandleFunc("/dishes/search", foodController.SearchDishes).Methods("GET")
	r.HandleFunc("/cuisines", foodController.ListCuisines).Methods("GET")

	admin := r.PathPrefix("/admin/places").Subrouter()
	admin.Use(middleware.AdminAuthMiddleware)
	admin.HandleFunc("/{id:[0-9]+}/dishes/list", foodController.ListDishesOfPlace).Methods("GET")
	admin.HandleFunc("", foodController.CreatePlace).Methods("POST")
	admin.HandleFunc("", foodController.ListAdminPlaces).Methods("GET")
	admin.HandleFunc("/{id:[0-9]+}", foodController.UpdatePlace).Methods("PUT")
	admin.HandleFunc("/{id:[0-9]+}", foodController.DeletePlace).Methods("DELETE")
	admin.HandleFunc("/{id:[0-9]+}/publish", foodController.PublishPlace).Methods("POST")
	admin.HandleFunc("/{id:[0-9]+}/unpublish", foodController.UnpublishPlace).Methods("POST")
	admin.HandleFunc("/{id:[0-9]+}/dishes", foodController.AddDish).Methods("POST")
	admin.HandleFunc("/dishes/{dish_id:[0-9]+}", foodController.UpdateDish).Methods("PUT")
	admin.HandleFunc("/dishes/{dish_id:[0-9]+}", foodController.DeleteDish).Methods("DELETE")
	admin.HandleFunc("/dishes/{dish_id:[0-9]+}/images", foodController.UploadDishImages).Methods("POST")
	admin.HandleFunc("/cuisines", foodController.CreateCuisine).Methods("POST")
	admin.HandleFunc("/cuisines/{id:[0-9]+}", foodController.UpdateCuisine).Methods("PUT")
	admin.HandleFunc("/cuisines/{id:[0-9]+}", foodController.DeleteCuisine).Methods("DELETE")

	user := r.PathPrefix("/user/places").Subrouter()
	user.Use(middleware.AuthMiddleware)
	user.HandleFunc("/{id:[0-9]+}/reviews", foodController.AddReview).Methods("POST")
	user.HandleFunc("/reviews/{review_id:[0-9]+}", foodController.UpdateReview).Methods("PUT")
	user.HandleFunc("/reviews/{review_id:[0-9]+}", foodController.DeleteReview).Methods("DELETE")

	return r
}
