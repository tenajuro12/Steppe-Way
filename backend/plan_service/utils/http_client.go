package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"plan_service/internal/models"
)

func GetAttraction(attractionID uint) (*models.AttractionResponse, error) {
	resp, err := http.Get(fmt.Sprintf("http://attraction-service:8085/attractions/%d", attractionID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("attraction service returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var attraction models.AttractionResponse
	if err := json.Unmarshal(body, &attraction); err != nil {
		return nil, err
	}

	return &attraction, nil
}

func GetEvent(eventID uint) (*models.EventResponse, error) {
	resp, err := http.Get(fmt.Sprintf("http://events-service:8083/events/%d", eventID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("events service returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var event models.EventResponse
	if err := json.Unmarshal(body, &event); err != nil {
		return nil, err
	}

	return &event, nil
}

func GetFoodPlace(placeID uint) (*models.FoodPlaceResponse, error) {
	resp, err := http.Get(fmt.Sprintf("http://food-service:8090/places/%d", placeID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("food service returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var foodPlace models.FoodPlaceResponse
	if err := json.Unmarshal(body, &foodPlace); err != nil {
		return nil, err
	}

	return &foodPlace, nil
}

func GetAccommodation(accommodationID uint) (*models.AccommodationResponse, error) {
	resp, err := http.Get(fmt.Sprintf("http://accommodation-service:8089/accommodations/%d", accommodationID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("accommodation service returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var accommodation models.AccommodationResponse
	if err := json.Unmarshal(body, &accommodation); err != nil {
		return nil, err
	}

	return &accommodation, nil
}
