package routes

import (
	"diplomaPorject/backend/blogs_service/internal/controllers"
	"diplomaPorject/backend/blogs_service/internal/middleware"
	"github.com/gorilla/mux"
)

func RegisterBlogRoutes(r *mux.Router) {
	r.HandleFunc("/blogs", controllers.GetBlogs).Methods("GET")
	r.HandleFunc("/blogs/{id:[0-9]+}", controllers.GetBlog).Methods("GET")

	protectedRoutes := r.PathPrefix("/blogs").Subrouter()
	protectedRoutes.Use(middleware.BlogsAuthMiddleware)
	protectedRoutes.HandleFunc("", controllers.CreateBlog).Methods("POST")
	protectedRoutes.HandleFunc("/{id:[0-9]+}", controllers.UpdateBlog).Methods("PUT")
	protectedRoutes.HandleFunc("/{id:[0-9]+}", controllers.DeleteBlog).Methods("DELETE")
	protectedRoutes.HandleFunc("/{id:[0-9]+}/like", controllers.LikeBlog).Methods("POST")
	protectedRoutes.HandleFunc("/{id:[0-9]+}/comment", controllers.AddComment).Methods("POST")
}
