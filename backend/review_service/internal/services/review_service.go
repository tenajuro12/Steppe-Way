package services

import (
	"errors"
	"review_service/internal/clients"
	"review_service/internal/models"
	"review_service/internal/repositories"
)

type ReviewService interface {
	CreateReview(review *models.Review) (*models.Review, error)
	GetReviewsByAttractionID(attractionID uint) ([]models.Review, error)
	GetReviewByID(reviewID uint) (*models.Review, error)
	GetReviewsByUserID(userID uint) ([]models.Review, error)
	UpdateReview(reviewID uint, review *models.Review) (*models.Review, error)
	DeleteReview(reviewID uint) error
	GetAttractionAverageRating(attractionID uint) (float64, error)
	GetUsernameByID(userID uint) (string, error)
}

type reviewService struct {
	reviewRepo    repositories.ReviewRepository
	profileClient clients.ProfileClient
}

func NewReviewService(reviewRepo repositories.ReviewRepository, profileClient clients.ProfileClient) ReviewService {
	return &reviewService{
		reviewRepo:    reviewRepo,
		profileClient: profileClient,
	}
}

func (r *reviewService) CreateReview(review *models.Review) (*models.Review, error) {
	if review.Rating < 1 || review.Rating > 5 {
		return nil, errors.New("rating must be between 1 and 5")
	}
	return r.reviewRepo.Create(review)
}

func (r *reviewService) GetReviewsByAttractionID(attractionID uint) ([]models.Review, error) {
	return r.reviewRepo.FindByAttractionId(attractionID)
}
func (r *reviewService) GetReviewByID(reviewID uint) (*models.Review, error) {
	return r.reviewRepo.FindById(reviewID)
}

func (r *reviewService) GetReviewsByUserID(userID uint) ([]models.Review, error) {
	return r.reviewRepo.FindByUserId(userID)
}

func (r *reviewService) UpdateReview(reviewID uint, review *models.Review) (*models.Review, error) {
	// Validation
	if review.Rating < 1 || review.Rating > 5 {
		return nil, errors.New("rating must be between 1 and 5")
	}

	return r.reviewRepo.Update(review)
}

func (r *reviewService) DeleteReview(reviewID uint) error {
	return r.reviewRepo.Delete(reviewID)
}
func (r *reviewService) GetAttractionAverageRating(attractionID uint) (float64, error) {
	return r.reviewRepo.GetAverageRatingByAttractionID(attractionID)
}

func (r *reviewService) GetUsernameByID(userID uint) (string, error) {
	profile, err := r.profileClient.GetProfileByUserID(userID)
	if err != nil {
		return "", err
	}
	return profile.Username, nil
}
