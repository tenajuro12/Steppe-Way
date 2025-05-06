package controllers

import (
	"crypto/rand"
	"diplomaPorject/backend/blogs_service/internal/models"
	"diplomaPorject/backend/blogs_service/utils"
	"diplomaPorject/backend/blogs_service/utils/db"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

func generateRandomFilename(originalFilename string) string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes) + filepath.Ext(originalFilename)
}

func uploadImages(files []*multipart.FileHeader) ([]string, error) {
	var imageURLs []string
	uploadDir := "/app/uploads"

	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %v", err)
	}

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %v", err)
		}
		defer file.Close()

		randomFilename := generateRandomFilename(fileHeader.Filename)
		filePath := filepath.Join(uploadDir, randomFilename)

		dst, err := os.Create(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to create file: %v", err)
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			return nil, fmt.Errorf("failed to save file: %v", err)
		}

		imageURLs = append(imageURLs, "/uploads/"+randomFilename)
	}
	return imageURLs, nil
}

func CreateBlog(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value("user_id").(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	username, _ := utils.GetUsername(userID) // <- НОВАЯ строка

	title := r.FormValue("title")
	content := r.FormValue("content")
	category := r.FormValue("category")

	files := r.MultipartForm.File["images"]
	imageURLs, err := uploadImages(files)
	if err != nil {
		http.Error(w, "Image upload failed", http.StatusInternalServerError)
		return
	}

	blog := models.Blog{
		Title:    title,
		Content:  content,
		Category: category,
		UserID:   userID,
		Username: username,
	}

	if err := db.DB.Create(&blog).Error; err != nil {
		http.Error(w, "Failed to create blog", http.StatusInternalServerError)
		return
	}

	for _, u := range imageURLs {
		db.DB.Create(&models.BlogImage{BlogID: blog.ID, URL: u})
	}

	db.DB.Preload("Images").First(&blog, blog.ID)
	json.NewEncoder(w).Encode(blog)
}

func GetBlogs(w http.ResponseWriter, r *http.Request) {
	var blogs []models.Blog

	query := db.DB.Preload("Comments").Preload("Images") // <- добавлено

	if category := r.URL.Query().Get("category"); category != "" {
		query = query.Where("category = ?", category)
	}

	page := 1
	pageSize := 10
	if pageParam := r.URL.Query().Get("page"); pageParam != "" {
		if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
			page = p
		}
	}

	offset := (page - 1) * pageSize

	var totalCount int64
	query.Model(&models.Blog{}).Count(&totalCount)

	if err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&blogs).Error; err != nil {
		http.Error(w, "Failed to fetch blogs", http.StatusInternalServerError)
		return
	}

	response := struct {
		Blogs    []models.Blog `json:"blogs"`
		Total    int64         `json:"total"`
		Page     int           `json:"page"`
		PageSize int           `json:"page_size"`
	}{
		Blogs:    blogs,
		Total:    totalCount,
		Page:     page,
		PageSize: pageSize,
	}

	json.NewEncoder(w).Encode(response)
}

func GetBlog(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid blog ID", http.StatusBadRequest)
		return
	}

	var blog models.Blog
	if err := db.DB.Preload("Comments").Preload("Images").First(&blog, id).Error; err != nil {
		http.Error(w, "Blog not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(blog)
}

func UpdateBlog(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	userIDValue := r.Context().Value("user_id")
	if userIDValue == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, ok := userIDValue.(uint)
	if !ok {
		http.Error(w, "Invalid user ID", http.StatusInternalServerError)
		return
	}

	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid blog ID", http.StatusBadRequest)
		return
	}

	var blog models.Blog
	if err := db.DB.Preload("Images").First(&blog, id).Error; err != nil {
		http.Error(w, "Blog not found", http.StatusNotFound)
		return
	}

	if blog.UserID != userID {
		http.Error(w, "Unauthorized to update this blog", http.StatusForbidden)
		return
	}

	title := r.FormValue("title")
	content := r.FormValue("content")
	category := r.FormValue("category")

	if title == "" || content == "" {
		http.Error(w, "Title and content are required", http.StatusBadRequest)
		return
	}

	blog.Title = title
	blog.Content = content
	blog.Category = category

	db.DB.Where("blog_id = ?", blog.ID).Delete(&models.BlogImage{})

	files := r.MultipartForm.File["images"]
	imageURLs, err := uploadImages(files)
	if err != nil {
		http.Error(w, fmt.Sprintf("Image upload failed: %v", err), http.StatusInternalServerError)
		return
	}

	for _, url := range imageURLs {
		image := models.BlogImage{
			BlogID: blog.ID,
			URL:    url,
		}
		db.DB.Create(&image)
	}

	if err := db.DB.Save(&blog).Error; err != nil {
		http.Error(w, "Failed to update blog", http.StatusInternalServerError)
		return
	}

	db.DB.Preload("Images").First(&blog, blog.ID)
	json.NewEncoder(w).Encode(blog)
}

func DeleteBlog(w http.ResponseWriter, r *http.Request) {
	userIDValue := r.Context().Value("user_id")
	if userIDValue == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, ok := userIDValue.(uint)
	if !ok {
		http.Error(w, "Invalid user ID", http.StatusInternalServerError)
		return
	}

	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid blog ID", http.StatusBadRequest)
		return
	}

	var existingBlog models.Blog
	if err := db.DB.First(&existingBlog, id).Error; err != nil {
		http.Error(w, "Blog not found", http.StatusNotFound)
		return
	}

	if existingBlog.UserID != userID {
		http.Error(w, "Unauthorized to delete this blog", http.StatusForbidden)
		return
	}

	if err := db.DB.Delete(&models.Blog{}, id).Error; err != nil {
		log.Printf("Error deleting blog: %v", err)
		http.Error(w, "Failed to delete blog", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func SyncUsername(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID   uint   `json:"user_id"`
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	db.DB.Model(&models.Blog{}).
		Where("user_id = ?", req.UserID).
		Update("username", req.Username)
	db.DB.Model(&models.Comment{}).
		Where("user_id = ?", req.UserID).
		Update("username", req.Username)
	w.WriteHeader(http.StatusNoContent)
}
