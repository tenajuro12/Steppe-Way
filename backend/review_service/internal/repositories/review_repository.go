package repositories

import (
	"fmt"
	"gorm.io/gorm"
	"review_service/internal/model"
)

type ReviewRepository interface {
	Create(review *model.Review) (*model.Review, error)
	FindById(id uint) (*model.Review, error)
	FindByAttractionId(id uint) ([]model.Review, error)
	FindByUserId(id uint) ([]model.Review, error)
	Update(review *model.Review) (*model.Review, error)
	Delete(id uint) error
	GetAverageRatingByAttractionID(attractionID uint) (float64, error)
}

type reviewRepository struct {
	db *gorm.DB
}

func NewReviewRepository(db *gorm.DB) reviewRepository {
	return reviewRepository{db: db}
}

func (r *reviewRepository) Create(review *model.Review) (*model.Review, error) {

	if err := r.db.Create(&review); err != nil {
		return nil, fmt.Errorf("failed to create review: %w", err)
	}
	return review, nil
}

func (r *reviewRepository) FindByID(id uint) (*model.Review, error) {
	var review model.Review
	if err := r.db.First(&review, id).Error; err != nil {
		return nil, fmt.Errorf("failed to find review: %w", err)
	}
	return &review, nil
}

func (r *reviewRepository) FindByAttractionID(attractionID uint) ([]model.Review, error) {
	var reviews []model.Review
	if err := r.db.Where("attraction_id = ?", attractionID).Find(&reviews).Error; err != nil {
		return nil, err
	}
	return reviews, nil
}

func (r *reviewRepository) FindByUserID(userID uint) ([]model.Review, error) {
	var reviews []model.Review
	if err := r.db.Where("user_id = ?", userID).Find(&reviews).Error; err != nil {
		return nil, err
	}
	return reviews, nil
}

func (r *reviewRepository) Update(review *model.Review) (*model.Review, error) {
	if err := r.db.Save(review).Error; err != nil {
		return nil, err
	}
	return review, nil
}

func (r *reviewRepository) Delete(id uint) error {
	return r.db.Delete(&model.Review{}, id).Error
}
func (r *reviewRepository) GetAverageRatingByAttractionID(attractionID uint) (float64, error) {
	var result struct {
		AverageRating float64
	}

	err := r.db.Model(&model.Review{}).
		Select("COALESCE(AVG(rating), 0) as average_rating").
		Where("attraction_id = ?", attractionID).
		Scan(&result).Error

	if err != nil {
		return 0, err
	}

	return result.AverageRating, nil
}
