package controllers

import (
	"encoding/json"
	"fmt"
	"food_service/internal/models"
	"food_service/utils/db"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

func (c *FoodController) AddDish(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	placeIDStr := vars["id"]
	placeID, err := strconv.ParseUint(placeIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid place ID", http.StatusBadRequest)
		return
	}

	var place models.Place
	if err := db.DB.First(&place, placeID).Error; err != nil {
		http.Error(w, "Place not found", http.StatusNotFound)
		return
	}

	adminIDValue := r.Context().Value("admin_id")
	if adminIDValue == nil {
		http.Error(w, "Unauthorized - admin ID missing", http.StatusUnauthorized)
		return
	}
	adminID, ok := adminIDValue.(uint)
	if !ok || adminID != place.AdminID {
		http.Error(w, "Unauthorized - not the place owner", http.StatusUnauthorized)
		return
	}

	name := r.FormValue("name")
	description := r.FormValue("description")

	var price float64
	if priceStr := r.FormValue("price"); priceStr != "" {
		price, err = strconv.ParseFloat(priceStr, 64)
		if err != nil {
			http.Error(w, "Invalid price format", http.StatusBadRequest)
			return
		}
	}

	isSpecialty := false
	if specialtyStr := r.FormValue("is_specialty"); specialtyStr == "true" {
		isSpecialty = true
	}

	var cuisineID *uint
	if cuisineIDStr := r.FormValue("cuisine_id"); cuisineIDStr != "" {
		id, err := strconv.ParseUint(cuisineIDStr, 10, 32)
		if err == nil {
			var cuisine models.Cuisine
			if err := db.DB.First(&cuisine, id).Error; err == nil {
				idUint := uint(id)
				cuisineID = &idUint
			}
		}
	}

	dish := models.Dish{
		PlaceID:     uint(placeID),
		Name:        name,
		Description: description,
		Price:       price,
		IsSpecialty: isSpecialty,
		CuisineID:   cuisineID,
	}

	tx := db.DB.Begin()
	if err := tx.Create(&dish).Error; err != nil {
		tx.Rollback()
		http.Error(w, "Failed to create dish", http.StatusInternalServerError)
		return
	}

	files := r.MultipartForm.File["images"]
	if len(files) > 0 {
		imageURLs, err := uploadImages(files, "dishes")
		if err != nil {
			tx.Rollback()
			http.Error(w, fmt.Sprintf("Failed to upload images: %v", err), http.StatusInternalServerError)
			return
		}

		for _, url := range imageURLs {
			image := models.DishImage{
				DishID: dish.ID,
				URL:    url,
			}
			if err := tx.Create(&image).Error; err != nil {
				tx.Rollback()
				http.Error(w, "Failed to save dish image", http.StatusInternalServerError)
				return
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	if err := db.DB.Preload("Images").First(&dish, dish.ID).Error; err != nil {
		http.Error(w, "Failed to reload dish", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dish)
}

func (c *FoodController) UpdateDish(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	dishIDStr := vars["dish_id"]
	dishID, err := strconv.ParseUint(dishIDStr, 10, 32)
	if err != nil {
		log.Printf("Invalid dish ID: %v", err)
		http.Error(w, "Invalid dish ID", http.StatusBadRequest)
		return
	}

	log.Printf("Updating dish ID: %s", dishIDStr)

	var dish models.Dish
	if err := db.DB.First(&dish, dishID).Error; err != nil {
		log.Printf("Dish not found: %v", err)
		http.Error(w, "Dish not found", http.StatusNotFound)
		return
	}

	log.Printf("Found dish: ID=%d, Name=%s, PlaceID=%d", dish.ID, dish.Name, dish.PlaceID)

	var place models.Place
	if err := db.DB.First(&place, dish.PlaceID).Error; err != nil {
		log.Printf("Place not found: %v", err)
		http.Error(w, "Place not found", http.StatusNotFound)
		return
	}

	adminIDValue := r.Context().Value("admin_id")
	if adminIDValue == nil {
		http.Error(w, "Unauthorized - admin ID missing", http.StatusUnauthorized)
		return
	}
	adminID, ok := adminIDValue.(uint)
	if !ok || adminID != place.AdminID {
		http.Error(w, "Unauthorized - not the place owner", http.StatusUnauthorized)
		return
	}

	tx := db.DB.Begin()

	if name := r.FormValue("name"); name != "" {
		log.Printf("Updating name from '%s' to '%s'", dish.Name, name)
		dish.Name = name
	}
	if description := r.FormValue("description"); description != "" {
		dish.Description = description
	}
	if priceStr := r.FormValue("price"); priceStr != "" {
		price, err := strconv.ParseFloat(priceStr, 64)
		if err == nil {
			dish.Price = price
		}
	}
	if specialtyStr := r.FormValue("is_specialty"); specialtyStr != "" {
		dish.IsSpecialty = specialtyStr == "true"
	}
	if cuisineIDStr := r.FormValue("cuisine_id"); cuisineIDStr != "" {
		id, err := strconv.ParseUint(cuisineIDStr, 10, 32)
		if err == nil {
			var cuisine models.Cuisine
			if err := db.DB.First(&cuisine, id).Error; err == nil {
				idUint := uint(id)
				dish.CuisineID = &idUint
			}
		}
	}

	if deleteExisting := r.FormValue("delete_existing_images"); deleteExisting == "true" {
		log.Printf("Deleting existing images for dish ID %d", dish.ID)
		if err := tx.Where("dish_id = ?", dish.ID).Delete(&models.DishImage{}).Error; err != nil {
			log.Printf("Error deleting existing images: %v", err)
			tx.Rollback()
			http.Error(w, "Failed to delete existing images", http.StatusInternalServerError)
			return
		}
	}

	files := r.MultipartForm.File["images"]
	if len(files) > 0 {
		imageURLs, err := uploadImages(files, "dishes")
		if err != nil {
			log.Printf("Error uploading images: %v", err)
			tx.Rollback()
			http.Error(w, fmt.Sprintf("Failed to upload images: %v", err), http.StatusInternalServerError)
			return
		}

		for _, url := range imageURLs {
			image := models.DishImage{
				DishID: dish.ID,
				URL:    url,
			}
			if err := tx.Create(&image).Error; err != nil {
				log.Printf("Error saving image: %v", err)
				tx.Rollback()
				http.Error(w, "Failed to save dish image", http.StatusInternalServerError)
				return
			}
		}
	}

	log.Printf("Saving updated dish ID %d", dish.ID)
	if err := tx.Save(&dish).Error; err != nil {
		log.Printf("Error saving dish: %v", err)
		tx.Rollback()
		http.Error(w, "Failed to update dish", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("Error committing transaction: %v", err)
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	if err := db.DB.Preload("Images").First(&dish, dish.ID).Error; err != nil {
		log.Printf("Error reloading dish: %v", err)
		http.Error(w, "Failed to reload dish", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully updated dish ID %d", dish.ID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dish)
}
func (c *FoodController) DeleteDish(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dishIDStr := vars["dish_id"]
	dishID, err := strconv.ParseUint(dishIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid dish ID", http.StatusBadRequest)
		return
	}

	var dish models.Dish
	if err := db.DB.First(&dish, dishID).Error; err != nil {
		http.Error(w, "Dish not found", http.StatusNotFound)
		return
	}

	var place models.Place
	if err := db.DB.First(&place, dish.PlaceID).Error; err != nil {
		http.Error(w, "Place not found", http.StatusNotFound)
		return
	}

	adminIDValue := r.Context().Value("admin_id")
	if adminIDValue == nil {
		http.Error(w, "Unauthorized - admin ID missing", http.StatusUnauthorized)
		return
	}
	adminID, ok := adminIDValue.(uint)
	if !ok || adminID != place.AdminID {
		http.Error(w, "Unauthorized - not the place owner", http.StatusUnauthorized)
		return
	}

	if err := db.DB.Delete(&dish).Error; err != nil {
		http.Error(w, "Failed to delete dish", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Dish deleted successfully"})
}

func (c *FoodController) UploadDishImages(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	dishIDStr := vars["dish_id"]
	dishID, err := strconv.ParseUint(dishIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid dish ID", http.StatusBadRequest)
		return
	}

	var dish models.Dish
	if err := db.DB.First(&dish, dishID).Error; err != nil {
		http.Error(w, "Dish not found", http.StatusNotFound)
		return
	}

	var place models.Place
	if err := db.DB.First(&place, dish.PlaceID).Error; err != nil {
		http.Error(w, "Place not found", http.StatusNotFound)
		return
	}

	adminIDValue := r.Context().Value("admin_id")
	if adminIDValue == nil {
		http.Error(w, "Unauthorized - admin ID missing", http.StatusUnauthorized)
		return
	}
	adminID, ok := adminIDValue.(uint)
	if !ok || adminID != place.AdminID {
		http.Error(w, "Unauthorized - not the place owner", http.StatusUnauthorized)
		return
	}

	tx := db.DB.Begin()

	if deleteExisting := r.FormValue("delete_existing_images"); deleteExisting == "true" {
		if err := tx.Where("dish_id = ?", dishID).Delete(&models.DishImage{}).Error; err != nil {
			tx.Rollback()
			http.Error(w, "Failed to delete existing dish images", http.StatusInternalServerError)
			return
		}
	}

	files := r.MultipartForm.File["images"]
	if len(files) > 0 {
		imageURLs, err := uploadImages(files, "dishes")
		if err != nil {
			tx.Rollback()
			http.Error(w, fmt.Sprintf("Failed to upload images: %v", err), http.StatusInternalServerError)
			return
		}

		for _, url := range imageURLs {
			image := models.DishImage{
				DishID: uint(dishID),
				URL:    url,
			}
			if err := tx.Create(&image).Error; err != nil {
				tx.Rollback()
				http.Error(w, "Failed to save dish image", http.StatusInternalServerError)
				return
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	if err := db.DB.Preload("Images").First(&dish, dishID).Error; err != nil {
		http.Error(w, "Failed to reload dish", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(dish)
}
func (c *FoodController) ListDishesOfPlace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	placeID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		http.Error(w, "Invalid place ID", http.StatusBadRequest)
		return
	}

	var dishes []models.Dish
	if err := db.DB.Preload("Images").Where("place_id = ?", placeID).Find(&dishes).Error; err != nil {
		http.Error(w, "Failed to fetch dishes", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dishes)
}
