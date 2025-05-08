package handlers

import (
	"encoding/json"
	"favorites_service/internal/models"
	"favorites_service/internal/service"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type FavoriteHandler struct {
	service services.FavoriteService
}

func NewFavoriteHandler() *FavoriteHandler {
	return &FavoriteHandler{
		service: services.NewFavoriteService(),
	}
}

func (h *FavoriteHandler) AddFavorite(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.FavoriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	favorite := models.Favorite{
		UserID:      userID,
		ItemID:      req.ItemID,
		ItemType:    req.ItemType,
		Title:       req.Title,
		ImageURL:    req.ImageURL,
		Description: req.Description,
		City:        req.City,
		Location:    req.Location,
		Date:        req.Date,
		Category:    req.Category,
	}

	if err := h.service.AddFavorite(&favorite); err != nil {
		http.Error(w, fmt.Sprintf("Failed to add favorite: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(favorite.ToResponse())
}

func (h *FavoriteHandler) GetUserFavorites(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	itemType := r.URL.Query().Get("type")

	favorites, err := h.service.GetUserFavorites(userID, itemType)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get favorites: %v", err), http.StatusInternalServerError)
		return
	}

	response := make([]models.FavoriteResponse, len(favorites))
	for i, fav := range favorites {
		response[i] = fav.ToResponse()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *FavoriteHandler) CheckFavorite(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	itemType := vars["type"]
	itemIDStr := vars["id"]

	itemID, err := strconv.ParseUint(itemIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	isFavorite, err := h.service.CheckFavorite(userID, uint(itemID), itemType)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to check favorite status: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"is_favorite": isFavorite})
}

func (h *FavoriteHandler) RemoveFavorite(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	itemType := vars["type"]
	itemIDStr := vars["id"]

	itemID, err := strconv.ParseUint(itemIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	if err := h.service.RemoveFavorite(userID, uint(itemID), itemType); err != nil {
		http.Error(w, fmt.Sprintf("Failed to remove favorite: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Favorite removed successfully"})
}

func (h *FavoriteHandler) GetFavorite(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	favoriteIDStr := vars["id"]

	favoriteID, err := strconv.ParseUint(favoriteIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid favorite ID", http.StatusBadRequest)
		return
	}

	favorite, err := h.service.GetFavorite(uint(favoriteID), userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get favorite: %v", err), http.StatusInternalServerError)
		return
	}

	if favorite == nil {
		http.Error(w, "Favorite not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(favorite.ToResponse())
}

func getUserID(r *http.Request) (uint, error) {
	// First, try to get from header
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr != "" {
		userID, err := strconv.ParseUint(userIDStr, 10, 32)
		if err != nil {
			return 0, err
		}
		return uint(userID), nil
	}
	
	userIDValue := r.Context().Value("user_id")
	if userIDValue == nil {
		return 0, fmt.Errorf("user ID not found")
	}

	userID, ok := userIDValue.(uint)
	if !ok {
		return 0, fmt.Errorf("invalid user ID format")
	}

	return userID, nil
}
