package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"plan_service/internal/models"
	"plan_service/internal/services"

	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type PlanHandler struct {
	service services.PlanService
}

func responseWriter(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func errorResponse(w http.ResponseWriter, message string, status int) {
	responseWriter(w, map[string]string{"error": message}, status)
}

func GetUserID(r *http.Request) uint {
	s := r.Header.Get("X-User-ID")
	if id, err := strconv.ParseUint(s, 10, 32); err == nil {
		return uint(id)
	}
	return 0
}

func (h *PlanHandler) CreatePlan(w http.ResponseWriter, r *http.Request) {
	var plan models.Plan
	if err := json.NewDecoder(r.Body).Decode(&plan); err != nil {
		errorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	plan.UserID = GetUserID(r)
	if plan.UserID == 0 {
		errorResponse(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	if err := h.service.CreatePlan(&plan); err != nil {
		errorResponse(w, "Failed to create plan: "+err.Error(), http.StatusInternalServerError)
		return
	}

	responseWriter(w, plan, http.StatusCreated)
}

func (h *PlanHandler) GetPlan(w http.ResponseWriter, r *http.Request) {
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
		errorResponse(w, "Plan not found", http.StatusNotFound)
		return
	}

	items, err := h.service.GetPlanItems(uint(planID))
	if err != nil {
		errorResponse(w, "Failed to retrieve plan items", http.StatusInternalServerError)
		return
	}

	response := struct {
		models.Plan
		Items []models.PlanItem `json:"items"`
	}{
		Plan:  *plan,
		Items: items,
	}

	responseWriter(w, response, http.StatusOK)
}

func (h *PlanHandler) UpdatePlan(w http.ResponseWriter, r *http.Request) {
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

	existingPlan, err := h.service.GetPlan(uint(planID), userID)
	if err != nil {
		errorResponse(w, "Plan not found or access denied", http.StatusNotFound)
		return
	}

	var updates models.Plan
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		errorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	updates.ID = existingPlan.ID
	updates.UserID = existingPlan.UserID

	if err := h.service.UpdatePlan(&updates); err != nil {
		errorResponse(w, "Failed to update plan: "+err.Error(), http.StatusInternalServerError)
		return
	}

	responseWriter(w, updates, http.StatusOK)
}

func (h *PlanHandler) DeletePlan(w http.ResponseWriter, r *http.Request) {
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

	if err := h.service.DeletePlan(uint(planID), userID); err != nil {
		errorResponse(w, "Failed to delete plan: "+err.Error(), http.StatusInternalServerError)
		return
	}

	responseWriter(w, map[string]string{"message": "Plan deleted successfully"}, http.StatusOK)
}

func (h *PlanHandler) GetUserPlans(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)
	if userID == 0 {
		errorResponse(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	plans, err := h.service.GetUserPlans(userID)
	if err != nil {
		errorResponse(w, "Failed to retrieve plans: "+err.Error(), http.StatusInternalServerError)
		return
	}

	responseWriter(w, plans, http.StatusOK)
}

func (h *PlanHandler) AddItemToPlan(w http.ResponseWriter, r *http.Request) {
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

	_, err = h.service.GetPlan(uint(planID), userID)
	if err != nil {
		errorResponse(w, "Plan not found or access denied", http.StatusNotFound)
		return
	}

	var planItem models.PlanItem
	if err := json.NewDecoder(r.Body).Decode(&planItem); err != nil {
		errorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	planItem.PlanID = uint(planID)

	if err := h.service.AddItemToPlan(&planItem); err != nil {
		errorResponse(w, "Failed to add item to plan: "+err.Error(), http.StatusInternalServerError)
		return
	}

	responseWriter(w, planItem, http.StatusCreated)
}

func (h *PlanHandler) UpdatePlanItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	itemID, err := strconv.ParseUint(vars["itemId"], 10, 32)
	if err != nil {
		errorResponse(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	userID := GetUserID(r)
	if userID == 0 {
		errorResponse(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	var updates models.PlanItem
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		errorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	updates.ID = uint(itemID)

	if err := h.service.UpdatePlanItem(&updates); err != nil {
		errorResponse(w, "Failed to update plan item: "+err.Error(), http.StatusInternalServerError)
		return
	}

	responseWriter(w, updates, http.StatusOK)
}

func (h *PlanHandler) DeletePlanItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	itemID, err := strconv.ParseUint(vars["itemId"], 10, 32)
	if err != nil {
		errorResponse(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	userID := GetUserID(r)
	if userID == 0 {
		errorResponse(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	if err := h.service.DeletePlanItem(uint(itemID), userID); err != nil {
		errorResponse(w, "Failed to delete plan item: "+err.Error(), http.StatusInternalServerError)
		return
	}

	responseWriter(w, map[string]string{"message": "Item removed from plan"}, http.StatusOK)
}

func (h *PlanHandler) OptimizeRoute(w http.ResponseWriter, r *http.Request) {
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

	if err := h.service.OptimizeRoute(uint(planID), userID); err != nil {
		errorResponse(w, "Failed to optimize route: "+err.Error(), http.StatusInternalServerError)
		return
	}

	items, err := h.service.GetPlanItems(uint(planID))
	if err != nil {
		errorResponse(w, "Route optimized but failed to retrieve items", http.StatusInternalServerError)
		return
	}

	responseWriter(w, items, http.StatusOK)
}

func (h *PlanHandler) GetTemplates(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")

	templates, err := h.service.GetTemplates(category)
	if err != nil {
		errorResponse(w, "Failed to retrieve templates: "+err.Error(), http.StatusInternalServerError)
		return
	}

	responseWriter(w, templates, http.StatusOK)
}

func (h *PlanHandler) CreatePlanFromTemplate(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)
	if userID == 0 {
		errorResponse(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	bodyBytes, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	log.Printf("Raw request body: %s", string(bodyBytes))

	var request struct {
		TemplateID uint   `json:"template_id"`
		StartDate  string `json:"start_date"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Printf("Error decoding JSON: %v", err)
		errorResponse(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	startDate, err := time.Parse(time.RFC3339, request.StartDate)
	if err != nil {
		log.Printf("Error parsing date: %v", err)
		errorResponse(w, fmt.Sprintf("Invalid date format: %v", err), http.StatusBadRequest)
		return
	}

	plan, err := h.service.CreatePlanFromTemplate(request.TemplateID, userID, startDate)
	if err != nil {
		log.Printf("Error creating plan: %v", err)
		errorResponse(w, "Failed to create plan from template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	responseWriter(w, plan, http.StatusCreated)
}
func (h *PlanHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	var template models.PlanTemplate
	if err := json.NewDecoder(r.Body).Decode(&template); err != nil {
		errorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID := GetUserID(r)
	if userID == 0 {
		errorResponse(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	if err := h.service.CreateTemplate(&template); err != nil {
		errorResponse(w, "Failed to create template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	responseWriter(w, template, http.StatusCreated)
}

func (h *PlanHandler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	templateID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		errorResponse(w, "Invalid template ID", http.StatusBadRequest)
		return
	}

	template, err := h.service.GetTemplate(uint(templateID))
	if err != nil {
		errorResponse(w, "Template not found", http.StatusNotFound)
		return
	}

	responseWriter(w, template, http.StatusOK)
}

func (h *PlanHandler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	templateID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		errorResponse(w, "Invalid template ID", http.StatusBadRequest)
		return
	}

	userID := GetUserID(r)
	if userID == 0 {
		errorResponse(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	var updates models.PlanTemplate
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		errorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	updates.ID = uint(templateID)

	if err := h.service.UpdateTemplate(&updates); err != nil {
		errorResponse(w, "Failed to update template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	responseWriter(w, updates, http.StatusOK)
}

func (h *PlanHandler) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	templateID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		errorResponse(w, "Invalid template ID", http.StatusBadRequest)
		return
	}

	userID := GetUserID(r)
	if userID == 0 {
		errorResponse(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	if err := h.service.DeleteTemplate(uint(templateID)); err != nil {
		errorResponse(w, "Failed to delete template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	responseWriter(w, map[string]string{"message": "Template deleted successfully"}, http.StatusOK)
}

func (h *PlanHandler) GetTemplateItems(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	templateID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		errorResponse(w, "Invalid template ID", http.StatusBadRequest)
		return
	}

	items, err := h.service.GetTemplateItems(uint(templateID))
	if err != nil {
		errorResponse(w, "Failed to retrieve template items: "+err.Error(), http.StatusInternalServerError)
		return
	}

	responseWriter(w, items, http.StatusOK)
}

func (h *PlanHandler) AddItemToTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	templateID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		errorResponse(w, "Invalid template ID", http.StatusBadRequest)
		return
	}

	userID := GetUserID(r)
	if userID == 0 {
		errorResponse(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	var templateItem models.TemplateItem
	if err := json.NewDecoder(r.Body).Decode(&templateItem); err != nil {
		errorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	templateItem.TemplateID = uint(templateID)

	if err := h.service.AddItemToTemplate(&templateItem); err != nil {
		errorResponse(w, "Failed to add item to template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	responseWriter(w, templateItem, http.StatusCreated)
}

func (h *PlanHandler) UpdateTemplateItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	itemID, err := strconv.ParseUint(vars["itemId"], 10, 32)
	if err != nil {
		errorResponse(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	userID := GetUserID(r)
	if userID == 0 {
		errorResponse(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	var updates models.TemplateItem
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		errorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	updates.ID = uint(itemID)

	if err := h.service.UpdateTemplateItem(&updates); err != nil {
		errorResponse(w, "Failed to update template item: "+err.Error(), http.StatusInternalServerError)
		return
	}

	responseWriter(w, updates, http.StatusOK)
}

func (h *PlanHandler) DeleteTemplateItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	itemID, err := strconv.ParseUint(vars["itemId"], 10, 32)
	if err != nil {
		errorResponse(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	userID := GetUserID(r)
	if userID == 0 {
		errorResponse(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	if err := h.service.DeleteTemplateItem(uint(itemID)); err != nil {
		errorResponse(w, "Failed to delete template item: "+err.Error(), http.StatusInternalServerError)
		return
	}

	responseWriter(w, map[string]string{"message": "Template item deleted successfully"}, http.StatusOK)
}
