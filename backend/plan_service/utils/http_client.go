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
