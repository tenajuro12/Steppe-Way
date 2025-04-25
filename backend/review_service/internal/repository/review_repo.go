package repository

import (
	"review_service/internal/models"
	"review_service/utils"
)

func CreateReview(r *models.Review) error {
	return utils.DB.Create(r).Error
}

func GetAllReviews() ([]models.Review, error) {
	var list []models.Review
	err := utils.DB.Find(&list).Error
	return list, err
}

func GetReviewsByAttraction(attractionID uint) ([]models.Review, error) {
	var list []models.Review
	err := utils.DB.Where("attraction_id = ?", attractionID).Find(&list).Error
	return list, err
}

func UpdateReview(r *models.Review) error {
	return utils.DB.Save(r).Error
}

func DeleteReview(id uint) error {
	return utils.DB.Delete(&models.Review{}, id).Error
}
