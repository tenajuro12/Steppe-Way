package handlers

import (
	"fmt"
	"net/http"
	"plan_service/internal/models"
	"plan_service/utils"
	database "plan_service/utils/db"
	"strconv"

	"github.com/gorilla/mux"
)

func (h *PlanHandler) GetPlanDirections(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	planID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		errorResponse(w, "Invalid plan ID", http.StatusBadRequest)
		return
	}

	userID := GetUserID(r)
	if userID == 0 {
		errorResponse(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	plan, err := h.service.GetPlan(uint(planID), userID)
	if err != nil {
		errorResponse(w, "Plan not found or access denied", http.StatusNotFound)
		return
	}

	items, err := h.service.GetPlanItems(uint(planID))
	if err != nil {
		errorResponse(w, "Failed to retrieve plan items", http.StatusInternalServerError)
		return
	}

	if len(items) < 2 {
		errorResponse(w, "Plan must have at least two items with locations to generate directions", http.StatusBadRequest)
		return
	}

	mode := r.URL.Query().Get("mode")
	if mode == "" {
		mode = "driving"
	}

	validModes := map[string]bool{
		"driving":   true,
		"walking":   true,
		"bicycling": true,
		"transit":   true,
	}
	if !validModes[mode] {
		errorResponse(w, "Invalid transportation mode. Use: driving, walking, bicycling, or transit", http.StatusBadRequest)
		return
	}

	var startItemID, endItemID uint64
	var startItem, endItem int = 0, len(items) - 1

	if startIDParam := r.URL.Query().Get("start_item_id"); startIDParam != "" {
		startItemID, err = strconv.ParseUint(startIDParam, 10, 32)
		if err != nil {
			errorResponse(w, "Invalid start item ID", http.StatusBadRequest)
			return
		}

		found := false
		for i, item := range items {
			if item.ID == uint(startItemID) {
				startItem = i
				found = true
				break
			}
		}
		if !found {
			errorResponse(w, "Start item not found in plan", http.StatusBadRequest)
			return
		}
	}

	if endIDParam := r.URL.Query().Get("end_item_id"); endIDParam != "" {
		endItemID, err = strconv.ParseUint(endIDParam, 10, 32)
		if err != nil {
			errorResponse(w, "Invalid end item ID", http.StatusBadRequest)
			return
		}

		found := false
		for i, item := range items {
			if item.ID == uint(endItemID) {
				endItem = i
				found = true
				break
			}
		}
		if !found {
			errorResponse(w, "End item not found in plan", http.StatusBadRequest)
			return
		}
	}

	if startItem > endItem {
		errorResponse(w, "Start item must come before end item in the plan order", http.StatusBadRequest)
		return
	}

	segmentItems := items[startItem : endItem+1]

	directionsResult, err := utils.GetDirectionsForPlanItems(segmentItems, mode)
	if err != nil {
		errorResponse(w, fmt.Sprintf("Failed to get directions: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	if directionsResult.Status != "OK" {
		if directionsResult.ErrorMessage != "" {
			errorResponse(w, fmt.Sprintf("Directions API error: %s", directionsResult.ErrorMessage), http.StatusBadRequest)
		} else {
			errorResponse(w, fmt.Sprintf("Directions API error status: %s", directionsResult.Status), http.StatusBadRequest)
		}
		return
	}

	response := struct {
		PlanID        uint                    `json:"plan_id"`
		PlanTitle     string                  `json:"plan_title"`
		Directions    *utils.DirectionsResult `json:"directions"`
		Items         []models.PlanItem       `json:"items"`
		TotalDistance string                  `json:"total_distance"`
		TotalDuration string                  `json:"total_duration"`
	}{
		PlanID:     plan.ID,
		PlanTitle:  plan.Title,
		Directions: directionsResult,
		Items:      segmentItems,
	}

	if len(directionsResult.Routes) > 0 {
		response.TotalDistance = directionsResult.Routes[0].Distance
		response.TotalDuration = directionsResult.Routes[0].Duration
	}

	responseWriter(w, response, http.StatusOK)
}

func (h *PlanHandler) GetDirectionsBetweenItems(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	fromItemID, err := strconv.ParseUint(vars["fromItemId"], 10, 32)
	if err != nil {
		errorResponse(w, "Invalid from item ID", http.StatusBadRequest)
		return
	}

	toItemID, err := strconv.ParseUint(vars["toItemId"], 10, 32)
	if err != nil {
		errorResponse(w, "Invalid to item ID", http.StatusBadRequest)
		return
	}

	userID := GetUserID(r)
	if userID == 0 {
		errorResponse(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	var fromItem models.PlanItem
	if err := database.DB.First(&fromItem, fromItemID).Error; err != nil {
		errorResponse(w, "From item not found", http.StatusNotFound)
		return
	}

	_, err = h.service.GetPlan(fromItem.PlanID, userID)
	if err != nil {
		errorResponse(w, "Plan access denied", http.StatusUnauthorized)
		return
	}

	var toItem models.PlanItem
	if err := database.DB.First(&toItem, toItemID).Error; err != nil {
		errorResponse(w, "To item not found", http.StatusNotFound)
		return
	}

	if fromItem.Location == "" || toItem.Location == "" {
		errorResponse(w, "Both items must have valid locations", http.StatusBadRequest)
		return
	}

	mode := r.URL.Query().Get("mode")
	if mode == "" {
		mode = "driving"
	}

	validModes := map[string]bool{
		"driving":   true,
		"walking":   true,
		"bicycling": true,
		"transit":   true,
	}
	if !validModes[mode] {
		errorResponse(w, "Invalid transportation mode. Use: driving, walking, bicycling, or transit", http.StatusBadRequest)
		return
	}

	directionsResult, err := utils.GetDirections(fromItem.Location, toItem.Location, nil, mode)
	if err != nil {
		errorResponse(w, fmt.Sprintf("Failed to get directions: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	if directionsResult.Status != "OK" {
		if directionsResult.ErrorMessage != "" {
			errorResponse(w, fmt.Sprintf("Directions API error: %s", directionsResult.ErrorMessage), http.StatusBadRequest)
		} else {
			errorResponse(w, fmt.Sprintf("Directions API error status: %s", directionsResult.Status), http.StatusBadRequest)
		}
		return
	}

	response := struct {
		FromItem   models.PlanItem         `json:"from_item"`
		ToItem     models.PlanItem         `json:"to_item"`
		Directions *utils.DirectionsResult `json:"directions"`
	}{
		FromItem:   fromItem,
		ToItem:     toItem,
		Directions: directionsResult,
	}

	responseWriter(w, response, http.StatusOK)
}
