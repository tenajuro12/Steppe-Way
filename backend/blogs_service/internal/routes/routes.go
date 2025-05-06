package routes

import (
	"diplomaPorject/backend/blogs_service/internal/controllers"
	"diplomaPorject/backend/blogs_service/internal/middleware"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func RegisterBlogRoutes(r *mux.Router) {
	uploadDir := "/app/uploads"
	log.Printf("Serving blog uploads from: %s", uploadDir)
	r.PathPrefix("/uploads/").Handler(
		http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadDir))),
	).Methods("GET")

	r.HandleFunc("/blogs", controllers.GetBlogs).Methods("GET")
	r.HandleFunc("/blogs/{id:[0-9]+}", controllers.GetBlog).Methods("GET")
	r.HandleFunc("/blogs/{id:[0-9]+}/comments", controllers.GetComments).Methods("GET")
	r.HandleFunc("/internal/sync-username",
		controllers.SyncUsername).Methods("POST")

	blogs := r.PathPrefix("/blogs").Subrouter()
	blogs.Use(middleware.BlogsAuthMiddleware)

	blogs.HandleFunc("", controllers.CreateBlog).Methods("POST")
	blogs.HandleFunc("/{id:[0-9]+}", controllers.UpdateBlog).Methods("PUT")
	blogs.HandleFunc("/{id:[0-9]+}", controllers.DeleteBlog).Methods("DELETE")
	blogs.HandleFunc("/{id:[0-9]+}/like", controllers.LikeBlog).Methods("POST")
	blogs.HandleFunc("/{id:[0-9]+}/unlike", controllers.UnlikeBlog).Methods("POST")

	blogs.HandleFunc("/{id:[0-9]+}/comments", controllers.AddComment).Methods("POST")

	comments := r.PathPrefix("/comments").Subrouter()
	comments.Use(middleware.BlogsAuthMiddleware)
	comments.HandleFunc("/{comment_id:[0-9]+}", controllers.UpdateComment).Methods("PUT")
	comments.HandleFunc("/{comment_id:[0-9]+}", controllers.DeleteComment).Methods("DELETE")
}
