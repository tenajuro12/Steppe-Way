package services

import (
	"errors"
	"log"
	"plan_service/internal/models"
	"plan_service/utils"
	database "plan_service/utils/db"
	"time"
)

type PlanService struct {
}

func (s *PlanService) CreatePlan(plan *models.Plan) error {
	return database.DB.Create(plan).Error
}

func (s *PlanService) GetPlan(planID uint, userID uint) (*models.Plan, error) {
	var plan models.Plan
	result := database.DB.Where("id = ? AND (user_id = ? OR is_public = ?)", planID, userID, true).First(&plan)
	if result.Error != nil {
		return nil, result.Error
	}
	return &plan, nil
}
func (s *PlanService) UpdatePlan(plan *models.Plan) error {
	return database.DB.Save(plan).Error
}

func (s *PlanService) DeletePlan(planID uint, userID uint) error {
	result := database.DB.Where("id = ? AND user_id = ?", planID, userID).Delete(&models.Plan{})
	if result.RowsAffected == 0 {
		return errors.New("plan not found or user not authorized")
	}
	return result.Error
}

func (s *PlanService) GetUserPlans(userID uint) ([]models.Plan, error) {
	var plans []models.Plan
	result := database.DB.Where("user_id = ?", userID).Find(&plans)
	return plans, result.Error
}

func (s *PlanService) AddItemToPlan(planItem *models.PlanItem) error {
	var plan models.Plan
	if err := database.DB.First(&plan, planItem.PlanID).Error; err != nil {
		return err
	}

	var maxOrder int
	database.DB.Model(&models.PlanItem{}).Where("plan_id = ?", planItem.PlanID).
		Select("COALESCE(MAX(order_index), 0)").Scan(&maxOrder)

	planItem.OrderIndex = maxOrder + 1

	// Set default scheduled date to plan's start date if not provided
	if planItem.ScheduledFor.IsZero() {
		planItem.ScheduledFor = plan.StartDate
	}

	if planItem.ItemType == "attraction" {
		attraction, err := utils.GetAttraction(planItem.ItemID)
		if err != nil {
			return err
		}
		planItem.Title = attraction.Title
		planItem.Description = attraction.Description
		planItem.Location = attraction.Location
		planItem.Address = attraction.Address
	} else if planItem.ItemType == "event" {
		event, err := utils.GetEvent(planItem.ItemID)
		if err != nil {
			return err
		}
		planItem.Title = event.Title
		planItem.Description = event.Description
		planItem.Location = event.Location
		planItem.ScheduledFor = event.StartDate

		duration := event.EndDate.Sub(event.StartDate)
		planItem.Duration = int(duration.Minutes())
	}

	return database.DB.Create(planItem).Error
}

func (s *PlanService) UpdatePlanItem(planItem *models.PlanItem) error {
	return database.DB.Save(planItem).Error
}

func (s *PlanService) DeletePlanItem(itemID uint, userID uint) error {
	var planItem models.PlanItem
	if err := database.DB.First(&planItem, itemID).Error; err != nil {
		return err
	}

	var plan models.Plan
	if err := database.DB.Where("id = ? AND user_id = ?", planItem.PlanID, userID).First(&plan).Error; err != nil {
		return errors.New("not authorized to modify this plan")
	}

	return database.DB.Delete(&models.PlanItem{}, itemID).Error
}

func (s *PlanService) GetPlanItems(planID uint) ([]models.PlanItem, error) {
	var items []models.PlanItem
	result := database.DB.Where("plan_id = ?", planID).Order("order_index").Find(&items)
	return items, result.Error
}

func (s *PlanService) OptimizeRoute(planID uint, userID uint) error {
	var plan models.Plan
	if err := database.DB.Where("id = ? AND user_id = ?", planID, userID).First(&plan).Error; err != nil {
		return errors.New("plan not found or user not authorized")
	}

	var items []models.PlanItem
	if err := database.DB.Where("plan_id = ?", planID).Find(&items).Error; err != nil {
		return err
	}

	log.Printf("Before optimization: %+v", items)

	optimizedItems := utils.OptimizeRoute(items)

	log.Printf("After optimization: %+v", optimizedItems)

	tx := database.DB.Begin()
	for _, item := range optimizedItems {
		if err := tx.Model(&models.PlanItem{}).Where("id = ?", item.ID).Update("order_index", item.OrderIndex).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

func (s *PlanService) GetTemplates(category string) ([]models.PlanTemplate, error) {
	var templates []models.PlanTemplate
	query := database.DB.Where("is_public = ?", true)

	if category != "" {
		query = query.Where("category = ?", category)
	}

	result := query.Find(&templates)
	return templates, result.Error
}

func (s *PlanService) CreatePlanFromTemplate(templateID uint, userID uint, startDate time.Time) (*models.Plan, error) {
	var template models.PlanTemplate
	if err := database.DB.First(&template, templateID).Error; err != nil {
		return nil, err
	}

	plan := models.Plan{
		Title:       template.Title,
		Description: template.Description,
		StartDate:   startDate,
		EndDate:     startDate.AddDate(0, 0, template.Duration),
		UserID:      userID,
		City:        template.City,
	}

	if err := database.DB.Create(&plan).Error; err != nil {
		return nil, err
	}

	var templateItems []models.TemplateItem
	if err := database.DB.Where("template_id = ?", templateID).Order("day_number, order_in_day").Find(&templateItems).Error; err != nil {
		return nil, err
	}

	for _, tItem := range templateItems {
		planItem := models.PlanItem{
			PlanID:       plan.ID,
			ItemType:     tItem.ItemType,
			ItemID:       tItem.ItemID,
			Title:        tItem.Title,
			Description:  tItem.Description,
			Location:     tItem.Location,
			ScheduledFor: startDate.AddDate(0, 0, tItem.DayNumber-1),
			Duration:     tItem.Duration,
			OrderIndex:   tItem.OrderInDay,
		}

		if err := database.DB.Create(&planItem).Error; err != nil {
			return nil, err
		}
	}

	return &plan, nil
}
