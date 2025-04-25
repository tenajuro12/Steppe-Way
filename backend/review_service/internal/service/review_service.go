package service

import (
	"review_service/internal/models"
	"review_service/internal/repository"
	"time"
)

type ReviewService interface {
	Create(*models.Review) error
	GetAll() ([]models.Review, error)
	GetByAttraction(uint) ([]models.Review, error)
	Update(*models.Review) error
	Delete(uint) error
}

type reviewService struct{}

func NewReviewService() ReviewService {
	return &reviewService{}
}

func (s *reviewService) Create(r *models.Review) error {
	r.CreatedAt = time.Now()
	r.UpdatedAt = time.Now()
	return repository.CreateReview(r)
}

func (s *reviewService) GetAll() ([]models.Review, error) {
	return repository.GetAllReviews()
}

func (s *reviewService) GetByAttraction(attractionID uint) ([]models.Review, error) {
	return repository.GetReviewsByAttraction(attractionID)
}

func (s *reviewService) Update(r *models.Review) error {
	r.UpdatedAt = time.Now()
	return repository.UpdateReview(r)
}

func (s *reviewService) Delete(id uint) error {
	return repository.DeleteReview(id)
}
