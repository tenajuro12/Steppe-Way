package controllers

import (
	"accommodation_service/internal/model"
	"accommodation_service/utils/db"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

type AccommodationController struct{}

func generateRandomFilename(originalFilename string) string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes) + filepath.Ext(originalFilename)
}

func uploadImages(files []*multipart.FileHeader) ([]string, error) {
	var imageURLs []string
	uploadDir := "/app/uploads/accommodations"

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

		imageURLs = append(imageURLs, "/uploads/accommodations/"+randomFilename)
	}
	return imageURLs, nil
}

func (c *AccommodationController) CreateAccommodation(w http.ResponseWriter, r *http.Request) {
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
	website := r.FormValue("website")
	location := r.FormValue("location")
	accommodationType := r.FormValue("type")

	var amenities models.StringArray
	if amenitiesStr := r.FormValue("amenities"); amenitiesStr != "" {
		if err := json.Unmarshal([]byte(amenitiesStr), &amenities); err != nil {
			http.Error(w, "Invalid amenities format", http.StatusBadRequest)
			return
		}
	}

	files := r.MultipartForm.File["images"]
	imageURLs, err := uploadImages(files)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to upload images: %v", err), http.StatusInternalServerError)
		return
	}

	accommodation := models.Accommodation{
		Name:        name,
		Description: description,
		City:        city,
		Address:     address,
		Website:     website,
		Location:    location,
		Type:        accommodationType,
		AdminID:     adminID,
		Amenities:   amenities,
	}

	tx := db.DB.Begin()
	if err := tx.Create(&accommodation).Error; err != nil {
		tx.Rollback()
		http.Error(w, "Failed to create accommodation", http.StatusInternalServerError)
		return
	}

	for _, url := range imageURLs {
		image := models.AccommodationImage{
			AccommodationID: accommodation.ID,
			URL:             url,
		}
		if err := tx.Create(&image).Error; err != nil {
			tx.Rollback()
			http.Error(w, "Failed to save image", http.StatusInternalServerError)
			return
		}
	}

	roomTypes := r.FormValue("room_types")
	if roomTypes != "" {
		log.Printf("Received room_types JSON: %s", roomTypes)

		var roomTypesRaw []map[string]interface{}
		if err := json.Unmarshal([]byte(roomTypes), &roomTypesRaw); err != nil {
			log.Printf("Error unmarshaling room types: %v", err)
			tx.Rollback()
			http.Error(w, "Invalid room types format", http.StatusBadRequest)
			return
		}

		for _, rawRoom := range roomTypesRaw {
			var roomType models.RoomType
			roomType.AccommodationID = accommodation.ID

			if name, ok := rawRoom["Name"].(string); ok {
				roomType.Name = name
			} else if name, ok := rawRoom["name"].(string); ok {
				roomType.Name = name
			}

			if desc, ok := rawRoom["Description"].(string); ok {
				roomType.Description = desc
			} else if desc, ok := rawRoom["description"].(string); ok {
				roomType.Description = desc
			}

			if price, ok := rawRoom["Price"].(float64); ok {
				roomType.Price = price
			} else if price, ok := rawRoom["price"].(float64); ok {
				roomType.Price = price
			}

			if maxGuests, ok := rawRoom["MaxGuests"].(float64); ok {
				roomType.MaxGuests = int(maxGuests)
			} else if maxGuests, ok := rawRoom["max_guests"].(float64); ok {
				roomType.MaxGuests = int(maxGuests)
			}

			if roomType.MaxGuests <= 0 {
				roomType.MaxGuests = 1
			}

			if bedType, ok := rawRoom["BedType"].(string); ok {
				roomType.BedType = bedType
			} else if bedType, ok := rawRoom["bed_type"].(string); ok {
				roomType.BedType = bedType
			}

			if roomType.BedType == "" {
				roomType.BedType = "Single"
			}

			if amenities, ok := rawRoom["Amenities"].([]interface{}); ok {
				for _, a := range amenities {
					if amenity, ok := a.(string); ok {
						roomType.Amenities = append(roomType.Amenities, amenity)
					}
				}
			} else if amenities, ok := rawRoom["amenities"].([]interface{}); ok {
				for _, a := range amenities {
					if amenity, ok := a.(string); ok {
						roomType.Amenities = append(roomType.Amenities, amenity)
					}
				}
			}

			if err := tx.Create(&roomType).Error; err != nil {
				tx.Rollback()
				http.Error(w, "Failed to create room type", http.StatusInternalServerError)
				return
			}

			log.Printf("Created room type: id=%d, name=%s, maxGuests=%d, bedType=%s",
				roomType.ID, roomType.Name, roomType.MaxGuests, roomType.BedType)
		}
	}
	if err := tx.Commit().Error; err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	if err := db.DB.Preload("Images").Preload("RoomTypes.Images").First(&accommodation, accommodation.ID).Error; err != nil {
		http.Error(w, "Failed to reload accommodation", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(accommodation)
}
func (c *AccommodationController) GetAccommodation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var accommodation models.Accommodation
	if err := db.DB.Preload("Images").Preload("RoomTypes.Images").Preload("Reviews").Preload("Reviews.Images").First(&accommodation, id).Error; err != nil {
		http.Error(w, "Accommodation not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(accommodation)
}

func (c *AccommodationController) UpdateAccommodation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	var accommodation models.Accommodation
	if err := db.DB.First(&accommodation, id).Error; err != nil {
		http.Error(w, "Accommodation not found", http.StatusNotFound)
		return
	}

	adminIDValue := r.Context().Value("admin_id")
	if adminIDValue == nil {
		http.Error(w, "Unauthorized - admin ID missing", http.StatusUnauthorized)
		return
	}
	adminID, ok := adminIDValue.(uint)
	if !ok || adminID != accommodation.AdminID {
		http.Error(w, "Unauthorized - not the accommodation owner", http.StatusUnauthorized)
		return
	}

	if name := r.FormValue("name"); name != "" {
		accommodation.Name = name
	}
	if description := r.FormValue("description"); description != "" {
		accommodation.Description = description
	}
	if city := r.FormValue("city"); city != "" {
		accommodation.City = city
	}
	if address := r.FormValue("address"); address != "" {
		accommodation.Address = address
	}
	if website := r.FormValue("website"); website != "" {
		accommodation.Website = website
	}
	if location := r.FormValue("location"); location != "" {
		accommodation.Location = location
	}
	if accommodationType := r.FormValue("type"); accommodationType != "" {
		accommodation.Type = accommodationType
	}
	if amenitiesStr := r.FormValue("amenities"); amenitiesStr != "" {
		var amenities models.StringArray
		if err := json.Unmarshal([]byte(amenitiesStr), &amenities); err != nil {
			http.Error(w, "Invalid amenities format", http.StatusBadRequest)
			return
		}
		accommodation.Amenities = amenities
	}

	tx := db.DB.Begin()

	files := r.MultipartForm.File["images"]
	if len(files) > 0 {
		imageURLs, err := uploadImages(files)
		if err != nil {
			tx.Rollback()
			http.Error(w, fmt.Sprintf("Failed to upload images: %v", err), http.StatusInternalServerError)
			return
		}

		if deleteExisting := r.FormValue("delete_existing_images"); deleteExisting == "true" {
			if err := tx.Where("accommodation_id = ?", accommodation.ID).Delete(&models.AccommodationImage{}).Error; err != nil {
				tx.Rollback()
				http.Error(w, "Failed to delete existing images", http.StatusInternalServerError)
				return
			}
		}

		for _, url := range imageURLs {
			image := models.AccommodationImage{
				AccommodationID: accommodation.ID,
				URL:             url,
			}
			if err := tx.Create(&image).Error; err != nil {
				tx.Rollback()
				http.Error(w, "Failed to save image", http.StatusInternalServerError)
				return
			}
		}
	}

	roomTypesStr := r.FormValue("room_types")
	if deletedRoomTypeIdsStr := r.FormValue("deleted_room_type_ids"); deletedRoomTypeIdsStr != "" {
		var deletedRoomTypeIds []uint
		if err := json.Unmarshal([]byte(deletedRoomTypeIdsStr), &deletedRoomTypeIds); err != nil {
			log.Printf("Error parsing deleted room type IDs: %v", err)
		} else {
			log.Printf("Deleting room types: %v", deletedRoomTypeIds)
			for _, roomTypeID := range deletedRoomTypeIds {
				var roomType models.RoomType
				if err := tx.First(&roomType, roomTypeID).Error; err != nil {
					log.Printf("Room type %d not found: %v", roomTypeID, err)
					continue
				}

				if roomType.AccommodationID != accommodation.ID {
					log.Printf("Room type %d does not belong to accommodation %d", roomTypeID, accommodation.ID)
					continue
				}

				if err := tx.Delete(&models.RoomType{}, roomTypeID).Error; err != nil {
					log.Printf("Error deleting room type %d: %v", roomTypeID, err)
				} else {
					log.Printf("Deleted room type %d", roomTypeID)
				}
			}
		}
	}
	if roomTypesStr != "" {
		var roomTypes []models.RoomType
		if err := json.Unmarshal([]byte(roomTypesStr), &roomTypes); err != nil {
			tx.Rollback()
			http.Error(w, "Invalid room types format", http.StatusBadRequest)
			return
		}

		updateStrategy := r.FormValue("room_types_update_strategy")
		if updateStrategy == "replace" {
			if err := tx.Where("accommodation_id = ?", accommodation.ID).Delete(&models.RoomType{}).Error; err != nil {
				tx.Rollback()
				http.Error(w, "Failed to delete existing room types", http.StatusInternalServerError)
				return
			}

			for i := range roomTypes {
				roomTypes[i].AccommodationID = accommodation.ID
				roomTypes[i].ID = 0
				if err := tx.Create(&roomTypes[i]).Error; err != nil {
					tx.Rollback()
					http.Error(w, "Failed to create new room type", http.StatusInternalServerError)
					return
				}
			}
		} else {
			for _, roomType := range roomTypes {
				if roomType.ID > 0 {
					var existingRoomType models.RoomType
					if err := tx.First(&existingRoomType, roomType.ID).Error; err != nil {
						tx.Rollback()
						http.Error(w, "Room type not found", http.StatusBadRequest)
						return
					}

					if existingRoomType.AccommodationID != accommodation.ID {
						tx.Rollback()
						http.Error(w, "Room type does not belong to this accommodation", http.StatusBadRequest)
						return
					}

					existingRoomType.Name = roomType.Name
					existingRoomType.Description = roomType.Description
					existingRoomType.Price = roomType.Price
					existingRoomType.MaxGuests = roomType.MaxGuests
					existingRoomType.BedType = roomType.BedType
					existingRoomType.Amenities = roomType.Amenities

					if err := tx.Save(&existingRoomType).Error; err != nil {
						tx.Rollback()
						http.Error(w, "Failed to update room type", http.StatusInternalServerError)
						return
					}
				} else {
					roomType.AccommodationID = accommodation.ID
					if err := tx.Create(&roomType).Error; err != nil {
						tx.Rollback()
						http.Error(w, "Failed to create room type", http.StatusInternalServerError)
						return
					}
				}
			}
		}
	}

	if err := tx.Save(&accommodation).Error; err != nil {
		tx.Rollback()
		http.Error(w, "Failed to update accommodation", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit().Error; err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	if err := db.DB.Preload("Images").Preload("RoomTypes.Images").First(&accommodation, accommodation.ID).Error; err != nil {
		http.Error(w, "Failed to reload accommodation", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(accommodation)
}

func (c *AccommodationController) DeleteAccommodation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var accommodation models.Accommodation
	if err := db.DB.First(&accommodation, id).Error; err != nil {
		http.Error(w, "Accommodation not found", http.StatusNotFound)
		return
	}

	adminIDValue := r.Context().Value("admin_id")
	if adminIDValue == nil {
		http.Error(w, "Unauthorized - admin ID missing", http.StatusUnauthorized)
		return
	}
	adminID, ok := adminIDValue.(uint)
	if !ok || adminID != accommodation.AdminID {
		http.Error(w, "Unauthorized - not the accommodation owner", http.StatusUnauthorized)
		return
	}

	if err := db.DB.Delete(&accommodation).Error; err != nil {
		http.Error(w, "Failed to delete accommodation", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Accommodation deleted successfully"})
}

func (c *AccommodationController) ListAccommodations(w http.ResponseWriter, r *http.Request) {
	var accommodations []models.Accommodation
	query := db.DB.Preload("Images").Preload("RoomTypes.Images").Preload("RoomTypes")

	if city := r.URL.Query().Get("city"); city != "" {
		query = query.Where("city LIKE ?", "%"+city+"%")
	}

	if accommodationType := r.URL.Query().Get("type"); accommodationType != "" {
		query = query.Where("type = ?", accommodationType)
	}

	if minPrice := r.URL.Query().Get("min_price"); minPrice != "" {
		minPriceFloat, err := strconv.ParseFloat(minPrice, 64)
		if err == nil {
			query = query.Where("id IN (SELECT accommodation_id FROM room_types WHERE price >= ?)", minPriceFloat)
		}
	}
	if maxPrice := r.URL.Query().Get("max_price"); maxPrice != "" {
		maxPriceFloat, err := strconv.ParseFloat(maxPrice, 64)
		if err == nil {
			query = query.Where("id IN (SELECT accommodation_id FROM room_types WHERE price <= ?)", maxPriceFloat)
		}
	}

	if maxGuests := r.URL.Query().Get("max_guests"); maxGuests != "" {
		maxGuestsInt, err := strconv.Atoi(maxGuests)
		if err == nil {
			query = query.Where("id IN (SELECT accommodation_id FROM room_types WHERE max_guests >= ?)", maxGuestsInt)
		}
	}

	if amenitiesStr := r.URL.Query().Get("amenities"); amenitiesStr != "" {
		amenities := strings.Split(amenitiesStr, ",")
		for _, amenity := range amenities {
			query = query.Where("amenities LIKE ?", "%"+amenity+"%")
		}
	}

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

				var results []struct {
					ID       uint
					Distance float64
				}

				var allAccommodations []models.Accommodation
				if err := db.DB.Find(&allAccommodations).Error; err != nil {
					http.Error(w, "Failed to fetch accommodations", http.StatusInternalServerError)
					return
				}

				for _, acc := range allAccommodations {
					parts := strings.Split(acc.Location, ",")
					if len(parts) != 2 {
						continue
					}
					accLat, err1 := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
					accLng, err2 := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
					if err1 != nil || err2 != nil {
						continue
					}

					distance := calculateDistance(latFloat, lngFloat, accLat, accLng)
					if distance <= maxDistance {
						results = append(results, struct {
							ID       uint
							Distance float64
						}{ID: acc.ID, Distance: distance})
					}
				}

				var ids []uint
				for _, result := range results {
					ids = append(ids, result.ID)
				}

				if len(ids) > 0 {
					query = query.Where("id IN ?", ids)
				} else {
					json.NewEncoder(w).Encode([]models.Accommodation{})
					return
				}
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
	query.Model(&models.Accommodation{}).Count(&totalCount)

	offset := (page - 1) * pageSize
	query = query.Offset(offset).Limit(pageSize)

	query = query.Order("created_at desc")

	query = query.Where("is_published = ?", true)

	if err := query.Find(&accommodations).Error; err != nil {
		http.Error(w, "Failed to fetch accommodations", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"accommodations": accommodations,
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

func (c *AccommodationController) ListAdminAccommodations(w http.ResponseWriter, r *http.Request) {
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

	var accommodations []models.Accommodation
	query := db.DB.Preload("Images").Preload("RoomTypes.Images").Preload("RoomTypes").Where("admin_id = ?", adminID)

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
	query.Model(&models.Accommodation{}).Count(&totalCount)

	offset := (page - 1) * pageSize
	query = query.Offset(offset).Limit(pageSize)

	query = query.Order("created_at desc")

	if err := query.Find(&accommodations).Error; err != nil {
		http.Error(w, "Failed to fetch accommodations", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"accommodations": accommodations,
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

func (c *AccommodationController) PublishAccommodation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var accommodation models.Accommodation
	if err := db.DB.First(&accommodation, id).Error; err != nil {
		http.Error(w, "Accommodation not found", http.StatusNotFound)
		return
	}

	adminIDValue := r.Context().Value("admin_id")
	if adminIDValue == nil {
		http.Error(w, "Unauthorized - admin ID missing", http.StatusUnauthorized)
		return
	}
	adminID, ok := adminIDValue.(uint)
	if !ok || adminID != accommodation.AdminID {
		http.Error(w, "Unauthorized - not the accommodation owner", http.StatusUnauthorized)
		return
	}

	accommodation.IsPublished = true
	if err := db.DB.Save(&accommodation).Error; err != nil {
		http.Error(w, "Failed to publish accommodation", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Accommodation published successfully"})
}

func (c *AccommodationController) UnpublishAccommodation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var accommodation models.Accommodation
	if err := db.DB.First(&accommodation, id).Error; err != nil {
		http.Error(w, "Accommodation not found", http.StatusNotFound)
		return
	}

	adminIDValue := r.Context().Value("admin_id")
	if adminIDValue == nil {
		http.Error(w, "Unauthorized - admin ID missing", http.StatusUnauthorized)
		return
	}
	adminID, ok := adminIDValue.(uint)
	if !ok || adminID != accommodation.AdminID {
		http.Error(w, "Unauthorized - not the accommodation owner", http.StatusUnauthorized)
		return
	}

	accommodation.IsPublished = false
	if err := db.DB.Save(&accommodation).Error; err != nil {
		http.Error(w, "Failed to unpublish accommodation", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Accommodation unpublished successfully"})
}

func (c *AccommodationController) AddReview(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	accommodationID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid accommodation ID", http.StatusBadRequest)
		return
	}

	var accommodation models.Accommodation
	if err := db.DB.First(&accommodation, accommodationID).Error; err != nil {
		http.Error(w, "Accommodation not found", http.StatusNotFound)
		return
	}

	userIDValue := r.Context().Value("user_id")
	if userIDValue == nil {
		http.Error(w, "Unauthorized - user ID missing", http.StatusUnauthorized)
		return
	}
	userID, ok := userIDValue.(uint)
	if !ok {
		http.Error(w, "Internal Server Error - Invalid User ID", http.StatusInternalServerError)
		return
	}

	usernameValue := r.Context().Value("username")
	username := ""
	if usernameValue != nil {
		if usernameStr, ok := usernameValue.(string); ok {
			username = usernameStr
		}
	}

	ratingStr := r.FormValue("rating")
	rating, err := strconv.Atoi(ratingStr)
	if err != nil || rating < 1 || rating > 5 {
		http.Error(w, "Invalid rating - must be between 1 and 5", http.StatusBadRequest)
		return
	}

	comment := r.FormValue("comment")

	review := models.AccommodationReview{
		AccommodationID: uint(accommodationID),
		UserID:          userID,
		Username:        username,
		Rating:          rating,
		Comment:         comment,
	}

	tx := db.DB.Begin()
	if err := tx.Create(&review).Error; err != nil {
		tx.Rollback()
		http.Error(w, "Failed to create review", http.StatusInternalServerError)
		return
	}

	files := r.MultipartForm.File["images"]
	if len(files) > 0 {
		imageURLs, err := uploadImages(files)
		if err != nil {
			tx.Rollback()
			http.Error(w, fmt.Sprintf("Failed to upload images: %v", err), http.StatusInternalServerError)
			return
		}

		for _, url := range imageURLs {
			image := models.ReviewImage{
				ReviewID: review.ID,
				URL:      url,
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

	if err := db.DB.Preload("Images").First(&review, review.ID).Error; err != nil {
		http.Error(w, "Failed to reload review", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(review)
}

func (c *AccommodationController) UpdateReview(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	reviewID, err := strconv.ParseUint(vars["review_id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid review ID", http.StatusBadRequest)
		return
	}

	var review models.AccommodationReview
	if err := db.DB.First(&review, reviewID).Error; err != nil {
		http.Error(w, "Review not found", http.StatusNotFound)
		return
	}

	userIDValue := r.Context().Value("user_id")
	if userIDValue == nil {
		http.Error(w, "Unauthorized - user ID missing", http.StatusUnauthorized)
		return
	}
	userID, ok := userIDValue.(uint)
	if !ok || userID != review.UserID {
		http.Error(w, "Unauthorized - not the review owner", http.StatusUnauthorized)
		return
	}

	if ratingStr := r.FormValue("rating"); ratingStr != "" {
		rating, err := strconv.Atoi(ratingStr)
		if err != nil || rating < 1 || rating > 5 {
			http.Error(w, "Invalid rating - must be between 1 and 5", http.StatusBadRequest)
			return
		}
		review.Rating = rating
	}

	if comment := r.FormValue("comment"); comment != "" {
		review.Comment = comment
	}

	tx := db.DB.Begin()

	files := r.MultipartForm.File["images"]
	if len(files) > 0 {

		if deleteExisting := r.FormValue("delete_existing_images"); deleteExisting == "true" {
			if err := tx.Where("review_id = ?", review.ID).Delete(&models.ReviewImage{}).Error; err != nil {
				tx.Rollback()
				http.Error(w, "Failed to delete existing images", http.StatusInternalServerError)
				return
			}
		}

		imageURLs, err := uploadImages(files)
		if err != nil {
			tx.Rollback()
			http.Error(w, fmt.Sprintf("Failed to upload images: %v", err), http.StatusInternalServerError)
			return
		}

		for _, url := range imageURLs {
			image := models.ReviewImage{
				ReviewID: review.ID,
				URL:      url,
			}
			if err := tx.Create(&image).Error; err != nil {
				tx.Rollback()
				http.Error(w, "Failed to save image", http.StatusInternalServerError)
				return
			}
		}
	}

	if err := tx.Save(&review).Error; err != nil {
		tx.Rollback()
		http.Error(w, "Failed to update review", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit().Error; err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	if err := db.DB.Preload("Images").First(&review, review.ID).Error; err != nil {
		http.Error(w, "Failed to reload review", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(review)
}

func (c *AccommodationController) DeleteReview(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	reviewID, err := strconv.ParseUint(vars["review_id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid review ID", http.StatusBadRequest)
		return
	}

	var review models.AccommodationReview
	if err := db.DB.First(&review, reviewID).Error; err != nil {
		http.Error(w, "Review not found", http.StatusNotFound)
		return
	}

	userIDValue := r.Context().Value("user_id")
	if userIDValue == nil {
		http.Error(w, "Unauthorized - user ID missing", http.StatusUnauthorized)
		return
	}
	userID, ok := userIDValue.(uint)
	if !ok || userID != review.UserID {

		var accommodation models.Accommodation
		if err := db.DB.First(&accommodation, review.AccommodationID).Error; err != nil {
			http.Error(w, "Unauthorized - not the review owner", http.StatusUnauthorized)
			return
		}

		adminIDValue := r.Context().Value("admin_id")
		if adminIDValue == nil {
			http.Error(w, "Unauthorized - not the review owner", http.StatusUnauthorized)
			return
		}
		adminID, ok := adminIDValue.(uint)
		if !ok || adminID != accommodation.AdminID {
			http.Error(w, "Unauthorized - not the review owner or accommodation owner", http.StatusUnauthorized)
			return
		}
	}

	if err := db.DB.Delete(&review).Error; err != nil {
		http.Error(w, "Failed to delete review", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Review deleted successfully"})
}

func (c *AccommodationController) GetAccommodationReviews(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var reviews []models.AccommodationReview
	query := db.DB.Preload("Images").Where("accommodation_id = ?", id)

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
	query.Model(&models.AccommodationReview{}).Count(&totalCount)

	offset := (page - 1) * pageSize
	query = query.Offset(offset).Limit(pageSize)

	query = query.Order("created_at desc")

	if err := query.Find(&reviews).Error; err != nil {
		http.Error(w, "Failed to fetch reviews", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"reviews": reviews,
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

func (c *AccommodationController) UploadRoomTypeImages(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	roomTypeID, err := strconv.ParseUint(vars["room_id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid room type ID", http.StatusBadRequest)
		return
	}

	var roomType models.RoomType
	if err := db.DB.First(&roomType, roomTypeID).Error; err != nil {
		http.Error(w, "Room type not found", http.StatusNotFound)
		return
	}

	var accommodation models.Accommodation
	if err := db.DB.First(&accommodation, roomType.AccommodationID).Error; err != nil {
		http.Error(w, "Accommodation not found", http.StatusNotFound)
		return
	}

	adminIDValue := r.Context().Value("admin_id")
	if adminIDValue == nil {
		http.Error(w, "Unauthorized - admin ID missing", http.StatusUnauthorized)
		return
	}
	adminID, ok := adminIDValue.(uint)
	if !ok || adminID != accommodation.AdminID {
		http.Error(w, "Unauthorized - not the accommodation owner", http.StatusUnauthorized)
		return
	}

	tx := db.DB.Begin()

	if deleteExisting := r.FormValue("delete_existing_images"); deleteExisting == "true" {
		if err := tx.Where("room_type_id = ?", roomTypeID).Delete(&models.RoomImage{}).Error; err != nil {
			tx.Rollback()
			http.Error(w, "Failed to delete existing room images", http.StatusInternalServerError)
			return
		}
	}

	files := r.MultipartForm.File["images"]
	if len(files) > 0 {
		imageURLs, err := uploadImages(files)
		if err != nil {
			tx.Rollback()
			http.Error(w, fmt.Sprintf("Failed to upload images: %v", err), http.StatusInternalServerError)
			return
		}

		for _, url := range imageURLs {
			image := models.RoomImage{
				RoomTypeID: uint(roomTypeID),
				URL:        url,
			}
			if err := tx.Create(&image).Error; err != nil {
				tx.Rollback()
				http.Error(w, "Failed to save room image", http.StatusInternalServerError)
				return
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	if err := db.DB.Preload("Images").First(&roomType, roomTypeID).Error; err != nil {
		http.Error(w, "Failed to reload room type", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(roomType)
}

func (c *AccommodationController) DeleteRoomType(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomTypeID, err := strconv.ParseUint(vars["room_id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid room type ID", http.StatusBadRequest)
		return
	}

	adminIDValue := r.Context().Value("admin_id")
	if adminIDValue == nil {
		http.Error(w, "Unauthorized - admin ID missing", http.StatusUnauthorized)
		return
	}
	adminID, ok := adminIDValue.(uint)
	if !ok {
		http.Error(w, "Internal Server Error - Invalid Admin ID", http.StatusInternalServerError)
		return
	}

	var roomType models.RoomType
	if err := db.DB.First(&roomType, roomTypeID).Error; err != nil {
		http.Error(w, "Room type not found", http.StatusNotFound)
		return
	}

	var accommodation models.Accommodation
	if err := db.DB.First(&accommodation, roomType.AccommodationID).Error; err != nil {
		http.Error(w, "Accommodation not found", http.StatusNotFound)
		return
	}

	if accommodation.AdminID != adminID {
		http.Error(w, "Unauthorized - not the accommodation owner", http.StatusUnauthorized)
		return
	}

	if err := db.DB.Delete(&roomType).Error; err != nil {
		http.Error(w, "Failed to delete room type", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Room type deleted successfully"})
}
