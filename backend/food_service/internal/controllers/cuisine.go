package controllers

import (
	"encoding/json"
	"food_service/internal/models"
	"food_service/utils/db"
	"github.com/gorilla/mux"
	"net/http"
)

func (c *FoodController) CreateCuisine(w http.ResponseWriter, r *http.Request) {
	var cuisine models.Cuisine
	if err := json.NewDecoder(r.Body).Decode(&cuisine); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if cuisine.Name == "" {
		http.Error(w, "Cuisine name is required", http.StatusBadRequest)
		return
	}

	if err := db.DB.Create(&cuisine).Error; err != nil {
		http.Error(w, "Failed to create cuisine", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(cuisine)
}

func (c *FoodController) ListCuisines(w http.ResponseWriter, r *http.Request) {
	var cuisines []models.Cuisine
	if err := db.DB.Order("name ASC").Find(&cuisines).Error; err != nil {
		http.Error(w, "Failed to fetch cuisines", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cuisines)
}

func (c *FoodController) UpdateCuisine(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var cuisine models.Cuisine
	if err := db.DB.First(&cuisine, id).Error; err != nil {
		http.Error(w, "Cuisine not found", http.StatusNotFound)
		return
	}

	var updateData models.Cuisine
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if updateData.Name != "" {
		cuisine.Name = updateData.Name
	}
	if updateData.Description != "" {
		cuisine.Description = updateData.Description
	}
	if updateData.Origin != "" {
		cuisine.Origin = updateData.Origin
	}

	if err := db.DB.Save(&cuisine).Error; err != nil {
		http.Error(w, "Failed to update cuisine", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cuisine)
}

func (c *FoodController) DeleteCuisine(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var cuisine models.Cuisine
	if err := db.DB.First(&cuisine, id).Error; err != nil {
		http.Error(w, "Cuisine not found", http.StatusNotFound)
		return
	}

	var count int64
	db.DB.Model(&models.Dish{}).Where("cuisine_id = ?", id).Count(&count)
	if count > 0 {
		http.Error(w, "Cannot delete cuisine that is used in dishes", http.StatusBadRequest)
		return
	}

	if err := db.DB.Delete(&cuisine).Error; err != nil {
		http.Error(w, "Failed to delete cuisine", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Cuisine deleted successfully"})
}
