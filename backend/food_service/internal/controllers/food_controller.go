package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"food_service/internal/models"
	"food_service/utils/db"
	"io"
	"log"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

type FoodController struct{}

func generateRandomFilename(originalFilename string) string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes) + filepath.Ext(originalFilename)
}

func uploadImages(files []*multipart.FileHeader, subdir string) ([]string, error) {
	var imageURLs []string
	uploadDir := fmt.Sprintf("/app/uploads/food/%s", subdir)

	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %v", err)
	}

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %v", err)
		}
		defer file.Close()

		randomFilename := generateRandomFilename(fileHeader.Filename)
		filePath := filepath.Join(uploadDir, randomFilename)

		dst, err := os.Create(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to create file: %v", err)
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			return nil, fmt.Errorf("failed to save file: %v", err)
		}

		imageURLs = append(imageURLs, fmt.Sprintf("/uploads/food/%s/%s", subdir, randomFilename))
	}
	return imageURLs, nil
}

func (c *FoodController) CreatePlace(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	adminIDValue := r.Context().Value("admin_id")
	if adminIDValue == nil {
		http.Error(w, "Unauthorized - admin ID missing", http.StatusUnauthorized)
		return
	}
	adminID, ok := adminIDValue.(uint)
	if !ok {
		http.Error(w, "Internal Server Error - Invalid Admin Id", http.StatusInternalServerError)
		return
	}

	name := r.FormValue("name")
	description := r.FormValue("description")
	city := r.FormValue("city")
	address := r.FormValue("address")
	placeType := r.FormValue("type")
	priceRange := r.FormValue("price_range")
	website := r.FormValue("website")
	phone := r.FormValue("phone")
	location := r.FormValue("location")

	place := models.Place{
		Name:        name,
		Description: description,
		City:        city,
		Address:     address,
		Type:        placeType,
		PriceRange:  priceRange,
		Website:     website,
		Phone:       phone,
		Location:    location,
		AdminID:     adminID,
	}

	tx := db.DB.Begin()

	if err := tx.Create(&place).Error; err != nil {
		tx.Rollback()
		http.Error(w, "Failed to create food place", http.StatusInternalServerError)
		return
	}

	cuisineIDs := r.Form["cuisine_ids"]
	if len(cuisineIDs) > 0 {
		for _, idStr := range cuisineIDs {
			id, err := strconv.ParseUint(idStr, 10, 32)
			if err != nil {
				continue
			}

			var cuisine models.Cuisine
			if err := tx.First(&cuisine, id).Error; err != nil {
				continue
			}

			if err := tx.Model(&place).Association("Cuisines").Append(&cuisine); err != nil {
				tx.Rollback()
				http.Error(w, "Failed to associate cuisine", http.StatusInternalServerError)
				return
			}
		}
	}

	files := r.MultipartForm.File["images"]
	if len(files) > 0 {
		imageURLs, err := uploadImages(files, "places")
		if err != nil {
			tx.Rollback()
			http.Error(w, fmt.Sprintf("Failed to upload images: %v", err), http.StatusInternalServerError)
			return
		}

		for _, url := range imageURLs {
			image := models.PlaceImage{
				PlaceID: place.ID,
				URL:     url,
			}
			if err := tx.Create(&image).Error; err != nil {
				tx.Rollback()
				http.Error(w, "Failed to save image", http.StatusInternalServerError)
				return
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	if err := db.DB.Preload("Images").Preload("Cuisines").First(&place, place.ID).Error; err != nil {
		http.Error(w, "Failed to reload place data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(place)
}

func (c *FoodController) GetPlace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var place models.Place
	if err := db.DB.Preload("Images").Preload("Cuisines").Preload("Dishes").Preload("Dishes.Images").Preload("Reviews").Preload("Reviews.Images").First(&place, id).Error; err != nil {
		log.Printf("Error fetching place with ID %s: %v", id, err)
		http.Error(w, "Place not found", http.StatusNotFound)
		return
	}

	if place.Dishes == nil {
		place.Dishes = []models.Dish{}
	}
	if place.Reviews == nil {
		place.Reviews = []models.FoodReview{}
	}
	if place.Cuisines == nil {
		place.Cuisines = []models.Cuisine{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(place)
}

func (c *FoodController) UpdatePlace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	var place models.Place
	if err := db.DB.First(&place, id).Error; err != nil {
		http.Error(w, "Place not found", http.StatusNotFound)
		return
	}

	adminIDValue := r.Context().Value("admin_id")
	if adminIDValue == nil {
		http.Error(w, "Unauthorized - admin ID missing", http.StatusUnauthorized)
		return
	}
	adminID, ok := adminIDValue.(uint)
	if !ok || adminID != place.AdminID {
		http.Error(w, "Unauthorized - not the place owner", http.StatusUnauthorized)
		return
	}

	if name := r.FormValue("name"); name != "" {
		place.Name = name
	}
	if description := r.FormValue("description"); description != "" {
		place.Description = description
	}
	if city := r.FormValue("city"); city != "" {
		place.City = city
	}
	if address := r.FormValue("address"); address != "" {
		place.Address = address
	}
	if placeType := r.FormValue("type"); placeType != "" {
		place.Type = placeType
	}
	if priceRange := r.FormValue("price_range"); priceRange != "" {
		place.PriceRange = priceRange
	}
	if website := r.FormValue("website"); website != "" {
		place.Website = website
	}
	if phone := r.FormValue("phone"); phone != "" {
		place.Phone = phone
	}
	if location := r.FormValue("location"); location != "" {
		place.Location = location
	}

	tx := db.DB.Begin()

	if cuisineIDs := r.Form["cuisine_ids"]; len(cuisineIDs) > 0 {
		if r.FormValue("update_cuisines") == "replace" {
			if err := tx.Model(&place).Association("Cuisines").Clear(); err != nil {
				tx.Rollback()
				http.Error(w, "Failed to clear cuisines", http.StatusInternalServerError)
				return
			}

			for _, idStr := range cuisineIDs {
				id, err := strconv.ParseUint(idStr, 10, 32)
				if err != nil {
					continue
				}

				var cuisine models.Cuisine
				if err := tx.First(&cuisine, id).Error; err != nil {
					continue
				}

				if err := tx.Model(&place).Association("Cuisines").Append(&cuisine); err != nil {
					tx.Rollback()
					http.Error(w, "Failed to associate cuisine", http.StatusInternalServerError)
					return
				}
			}
		}
	}

	files := r.MultipartForm.File["images"]
	if len(files) > 0 {
		if deleteExisting := r.FormValue("delete_existing_images"); deleteExisting == "true" {
			if err := tx.Where("place_id = ?", place.ID).Delete(&models.PlaceImage{}).Error; err != nil {
				tx.Rollback()
				http.Error(w, "Failed to delete existing images", http.StatusInternalServerError)
				return
			}
		}

		imageURLs, err := uploadImages(files, "places")
		if err != nil {
			tx.Rollback()
			http.Error(w, fmt.Sprintf("Failed to upload images: %v", err), http.StatusInternalServerError)
			return
		}

		for _, url := range imageURLs {
			image := models.PlaceImage{
				PlaceID: place.ID,
				URL:     url,
			}
			if err := tx.Create(&image).Error; err != nil {
				tx.Rollback()
				http.Error(w, "Failed to save image", http.StatusInternalServerError)
				return
			}
		}
	}

	if err := tx.Save(&place).Error; err != nil {
		tx.Rollback()
		http.Error(w, "Failed to update place", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit().Error; err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	if err := db.DB.Preload("Images").Preload("Cuisines").First(&place, place.ID).Error; err != nil {
		http.Error(w, "Failed to reload place data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(place)
}

func (c *FoodController) DeletePlace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var place models.Place
	if err := db.DB.First(&place, id).Error; err != nil {
		http.Error(w, "Place not found", http.StatusNotFound)
		return
	}

	adminIDValue := r.Context().Value("admin_id")
	if adminIDValue == nil {
		http.Error(w, "Unauthorized - admin ID missing", http.StatusUnauthorized)
		return
	}
	adminID, ok := adminIDValue.(uint)
	if !ok || adminID != place.AdminID {
		http.Error(w, "Unauthorized - not the place owner", http.StatusUnauthorized)
		return
	}

	if err := db.DB.Delete(&place).Error; err != nil {
		http.Error(w, "Failed to delete place", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Place deleted successfully"})
}

func (c *FoodController) ListPlaces(w http.ResponseWriter, r *http.Request) {
	query := db.DB.Preload("Images").Preload("Cuisines")

	if city := r.URL.Query().Get("city"); city != "" {
		query = query.Where("city LIKE ?", "%"+city+"%")
	}

	if placeType := r.URL.Query().Get("type"); placeType != "" {
		query = query.Where("type = ?", placeType)
	}

	if cuisine := r.URL.Query().Get("cuisine"); cuisine != "" {
		query = query.Joins("JOIN place_cuisines ON places.id = place_cuisines.place_id").
			Joins("JOIN cuisines ON cuisines.id = place_cuisines.cuisine_id").
			Where("cuisines.name LIKE ?", "%"+cuisine+"%")
	}

	if minRating := r.URL.Query().Get("min_rating"); minRating != "" {
		rating, err := strconv.ParseFloat(minRating, 64)
		if err == nil {
			query = query.Where("average_rating >= ?", rating)
		}
	}

	if priceRange := r.URL.Query().Get("price_range"); priceRange != "" {
		query = query.Where("price_range = ?", priceRange)
	}

	var places []models.Place
	if lat := r.URL.Query().Get("lat"); lat != "" {
		if lng := r.URL.Query().Get("lng"); lng != "" {
			latFloat, latErr := strconv.ParseFloat(lat, 64)
			lngFloat, lngErr := strconv.ParseFloat(lng, 64)
			if latErr == nil && lngErr == nil {
				maxDistance := 10.0
				if distance := r.URL.Query().Get("distance"); distance != "" {
					if distFloat, err := strconv.ParseFloat(distance, 64); err == nil {
						maxDistance = distFloat
					}
				}

				var placesWithLocation []models.Place
				if err := query.Where("location IS NOT NULL AND location != ''").Find(&placesWithLocation).Error; err != nil {
					http.Error(w, "Failed to fetch places", http.StatusInternalServerError)
					return
				}

				var placesWithDistance []models.PlaceWithDistance
				for _, place := range placesWithLocation {
					parts := strings.Split(place.Location, ",")
					if len(parts) != 2 {
						continue
					}
					placeLat, err1 := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
					placeLng, err2 := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
					if err1 != nil || err2 != nil {
						continue
					}

					distance := calculateDistance(latFloat, lngFloat, placeLat, placeLng)
					if distance <= maxDistance {
						placesWithDistance = append(placesWithDistance, models.PlaceWithDistance{
							Place:    place,
							Distance: math.Round(distance*10) / 10,
						})
					}
				}

				sort.Slice(placesWithDistance, func(i, j int) bool {
					return placesWithDistance[i].Distance < placesWithDistance[j].Distance
				})

				for _, pwd := range placesWithDistance {
					places = append(places, pwd.Place)
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(places)
				return
			}
		}
	}

	page := 1
	pageSize := 20
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if pageInt, err := strconv.Atoi(pageStr); err == nil && pageInt > 0 {
			page = pageInt
		}
	}
	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		if pageSizeInt, err := strconv.Atoi(pageSizeStr); err == nil && pageSizeInt > 0 {
			pageSize = pageSizeInt
		}
	}

	var totalCount int64
	query.Model(&models.Place{}).Where("is_published = ?", true).Count(&totalCount)

	offset := (page - 1) * pageSize
	query = query.Offset(offset).Limit(pageSize)

	query = query.Order("average_rating DESC, name ASC")

	query = query.Where("is_published = ?", true)

	if err := query.Find(&places).Error; err != nil {
		http.Error(w, "Failed to fetch places", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"places": places,
		"pagination": map[string]interface{}{
			"page":        page,
			"page_size":   pageSize,
			"total":       totalCount,
			"total_pages": int(math.Ceil(float64(totalCount) / float64(pageSize))),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (c *FoodController) PublishPlace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var place models.Place
	if err := db.DB.First(&place, id).Error; err != nil {
		http.Error(w, "Place not found", http.StatusNotFound)
		return
	}

	adminIDValue := r.Context().Value("admin_id")
	if adminIDValue == nil {
		http.Error(w, "Unauthorized - admin ID missing", http.StatusUnauthorized)
		return
	}
	adminID, ok := adminIDValue.(uint)
	if !ok || adminID != place.AdminID {
		http.Error(w, "Unauthorized - not the place owner", http.StatusUnauthorized)
		return
	}

	place.IsPublished = true
	if err := db.DB.Save(&place).Error; err != nil {
		http.Error(w, "Failed to publish place", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Place published successfully"})
}

func (c *FoodController) UnpublishPlace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var place models.Place
	if err := db.DB.First(&place, id).Error; err != nil {
		http.Error(w, "Place not found", http.StatusNotFound)
		return
	}

	adminIDValue := r.Context().Value("admin_id")
	if adminIDValue == nil {
		http.Error(w, "Unauthorized - admin ID missing", http.StatusUnauthorized)
		return
	}
	adminID, ok := adminIDValue.(uint)
	if !ok || adminID != place.AdminID {
		http.Error(w, "Unauthorized - not the place owner", http.StatusUnauthorized)
		return
	}

	place.IsPublished = false
	if err := db.DB.Save(&place).Error; err != nil {
		http.Error(w, "Failed to unpublish place", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Place unpublished successfully"})
}

func (c *FoodController) ListAdminPlaces(w http.ResponseWriter, r *http.Request) {
	adminIDValue := r.Context().Value("admin_id")
	if adminIDValue == nil {
		http.Error(w, "Unauthorized - admin ID missing", http.StatusUnauthorized)
		return
	}
	adminID, ok := adminIDValue.(uint)
	if !ok {
		http.Error(w, "Internal Server Error - Invalid Admin Id", http.StatusInternalServerError)
		return
	}

	var places []models.Place
	query := db.DB.Preload("Images").Preload("Cuisines").Where("admin_id = ?", adminID)

	page := 1
	pageSize := 20
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if pageInt, err := strconv.Atoi(pageStr); err == nil && pageInt > 0 {
			page = pageInt
		}
	}
	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		if pageSizeInt, err := strconv.Atoi(pageSizeStr); err == nil && pageSizeInt > 0 {
			pageSize = pageSizeInt
		}
	}

	var totalCount int64
	query.Model(&models.Place{}).Count(&totalCount)

	offset := (page - 1) * pageSize
	query = query.Offset(offset).Limit(pageSize)

	query = query.Order("created_at desc")

	if err := query.Find(&places).Error; err != nil {
		http.Error(w, "Failed to fetch places", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"places": places,
		"pagination": map[string]interface{}{
			"page":        page,
			"page_size":   pageSize,
			"total":       totalCount,
			"total_pages": int(math.Ceil(float64(totalCount) / float64(pageSize))),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (c *FoodController) SearchPlaces(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Search query is required", http.StatusBadRequest)
		return
	}

	var places []models.Place
	searchQuery := "%" + query + "%"

	db.DB.Preload("Images").Preload("Cuisines").
		Where("is_published = ? AND (name LIKE ? OR description LIKE ? OR city LIKE ?)",
			true, searchQuery, searchQuery, searchQuery).
		Order("average_rating DESC").
		Limit(20).
		Find(&places)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(places)
}

func (c *FoodController) SearchDishes(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Search query is required", http.StatusBadRequest)
		return
	}

	searchQuery := "%" + query + "%"

	type DishWithPlace struct {
		Dish  models.Dish  `json:"dish"`
		Place models.Place `json:"place"`
	}

	var results []DishWithPlace

	rows, err := db.DB.Table("dishes").
		Select("dishes.*, places.id as place_id, places.name as place_name, places.city as place_city, places.address as place_address").
		Joins("JOIN places ON dishes.place_id = places.id").
		Where("dishes.name LIKE ? AND places.is_published = ?", searchQuery, true).
		Order("places.average_rating DESC").
		Limit(50).
		Rows()

	if err != nil {
		http.Error(w, "Failed to search dishes", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var dish models.Dish
		var place models.Place
		var placeName, placeCity, placeAddress string

		if err := db.DB.ScanRows(rows, &dish); err != nil {
			continue
		}

		rows.Scan(nil, nil, nil, nil, nil, nil, nil, nil, &place.ID, &placeName, &placeCity, &placeAddress)

		place.Name = placeName
		place.City = placeCity
		place.Address = placeAddress

		var dishImages []models.DishImage
		db.DB.Where("dish_id = ?", dish.ID).Find(&dishImages)
		dish.Images = dishImages

		var placeImages []models.PlaceImage
		db.DB.Where("place_id = ?", place.ID).Find(&placeImages)
		place.Images = placeImages

		results = append(results, DishWithPlace{
			Dish:  dish,
			Place: place,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func getUserProfileInfo(userID uint) (string, string, error) {
	defaultUsername := "User"
	defaultProfileImg := ""

	profileServiceURL := fmt.Sprintf("http://profile-service:8084/user/profiles/%d", userID)

	resp, err := http.Get(profileServiceURL)
	if err != nil {
		log.Printf("Error fetching user profile: %v", err)
		return defaultUsername, defaultProfileImg, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Profile service returned non-200 status: %d", resp.StatusCode)
		return defaultUsername, defaultProfileImg, fmt.Errorf("profile service error: %d", resp.StatusCode)
	}

	var profileData struct {
		Username   string `json:"username"`
		ProfileImg string `json:"profile_img"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&profileData); err != nil {
		log.Printf("Error decoding profile response: %v", err)
		return defaultUsername, defaultProfileImg, err
	}

	return profileData.Username, profileData.ProfileImg, nil
}

func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371

	lat1Rad := lat1 * math.Pi / 180
	lon1Rad := lon1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lon2Rad := lon2 * math.Pi / 180

	dLat := lat2Rad - lat1Rad
	dLon := lon2Rad - lon1Rad
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	distance := R * c

	return distance
}

func getUserID(r *http.Request) (uint, error) {
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

func updateAverageRating(placeID uint) error {
	var count int64
	var sum float64

	err := db.DB.Model(&models.FoodReview{}).
		Where("place_id = ?", placeID).
		Count(&count).Error
	if err != nil {
		return err
	}

	if count == 0 {
		return db.DB.Model(&models.Place{}).
			Where("id = ?", placeID).
			Update("average_rating", 0).Error
	}

	err = db.DB.Model(&models.FoodReview{}).
		Where("place_id = ?", placeID).
		Select("COALESCE(SUM(rating), 0)").
		Scan(&sum).Error
	if err != nil {
		return err
	}

	average := math.Round((sum/float64(count))*10) / 10

	return db.DB.Model(&models.Place{}).
		Where("id = ?", placeID).
		Update("average_rating", average).Error
}
