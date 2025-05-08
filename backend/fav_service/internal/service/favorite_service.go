package services

import (
	"errors"
	"favorites_service/internal/models"
	"favorites_service/utils/db"
	"fmt"

	"gorm.io/gorm"
)

type FavoriteService struct{}

func NewFavoriteService() FavoriteService {
	return FavoriteService{}
}

func (s FavoriteService) AddFavorite(favorite *models.Favorite) error {
	var existingFavorite models.Favorite
	result := db.DB.Where("user_id = ? AND item_id = ? AND item_type = ?",
		favorite.UserID, favorite.ItemID, favorite.ItemType).First(&existingFavorite)

	if result.Error == nil {
		*favorite = existingFavorite
		return nil
	}

	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return result.Error
	}

	if err := db.DB.Create(favorite).Error; err != nil {
		return err
	}

	return nil
}

func (s FavoriteService) GetUserFavorites(userID uint, itemType string) ([]models.Favorite, error) {
	var favorites []models.Favorite
	query := db.DB.Where("user_id = ?", userID)

	if itemType != "" {
		query = query.Where("item_type = ?", itemType)
	}

	if err := query.Order("created_at DESC").Find(&favorites).Error; err != nil {
		return nil, err
	}

	return favorites, nil
}

func (s FavoriteService) CheckFavorite(userID, itemID uint, itemType string) (bool, error) {
	var count int64
	err := db.DB.Model(&models.Favorite{}).
		Where("user_id = ? AND item_id = ? AND item_type = ?", userID, itemID, itemType).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (s FavoriteService) RemoveFavorite(userID, itemID uint, itemType string) error {
	result := db.DB.Where("user_id = ? AND item_id = ? AND item_type = ?",
		userID, itemID, itemType).Delete(&models.Favorite{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("favorite not found")
	}

	return nil
}

func (s FavoriteService) GetFavorite(favoriteID, userID uint) (*models.Favorite, error) {
	var favorite models.Favorite
	if err := db.DB.Where("id = ? AND user_id = ?", favoriteID, userID).First(&favorite).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &favorite, nil
}

func (s FavoriteService) GetFavoriteStatistics(userID uint) (map[string]int, error) {
	type Result struct {
		ItemType string
		Count    int
	}

	var results []Result
	if err := db.DB.Model(&models.Favorite{}).
		Select("item_type, COUNT(*) as count").
		Where("user_id = ?", userID).
		Group("item_type").
		Scan(&results).Error; err != nil {
		return nil, err
	}

	statistics := make(map[string]int)
	for _, result := range results {
		statistics[result.ItemType] = result.Count
	}

	return statistics, nil
}

func (s FavoriteService) RemoveFavoriteByID(favoriteID, userID uint) error {
	result := db.DB.Where("id = ? AND user_id = ?", favoriteID, userID).
		Delete(&models.Favorite{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("favorite not found")
	}

	return nil
}
