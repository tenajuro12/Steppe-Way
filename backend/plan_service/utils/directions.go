package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"plan_service/internal/models"
	"strings"
	"time"
)

const (
	GoogleMapsDirectionsAPI = "https://maps.googleapis.com/maps/api/directions/json"
)

func getDirectionsAPIKey() string {
	apiKey := os.Getenv("GOOGLE_MAPS_API_KEY")
	if apiKey == "" {
		log.Println("Warning: GOOGLE_MAPS_API_KEY environment variable is not set")
	}
	return apiKey
}

type DirectionsResponse struct {
	Status            string             `json:"status"`
	Routes            []Route            `json:"routes"`
	GeocodedWaypoints []GeocodedWaypoint `json:"geocoded_waypoints"`
	ErrorMessage      string             `json:"error_message,omitempty"`
}

type GeocodedWaypoint struct {
	GeocoderStatus string   `json:"geocoder_status"`
	PlaceID        string   `json:"place_id"`
	Types          []string `json:"types"`
}

type Route struct {
	Summary          string       `json:"summary"`
	Legs             []Leg        `json:"legs"`
	OverviewPolyline Polyline     `json:"overview_polyline"`
	Bounds           LatLngBounds `json:"bounds"`
	Warnings         []string     `json:"warnings"`
	WaypointOrder    []int        `json:"waypoint_order"`
}

type Leg struct {
	Steps             []Step    `json:"steps"`
	Distance          TextValue `json:"distance"`
	Duration          TextValue `json:"duration"`
	DurationInTraffic TextValue `json:"duration_in_traffic,omitempty"`
	StartLocation     LatLng    `json:"start_location"`
	EndLocation       LatLng    `json:"end_location"`
	StartAddress      string    `json:"start_address"`
	EndAddress        string    `json:"end_address"`
}

type Step struct {
	TravelMode       string    `json:"travel_mode"`
	StartLocation    LatLng    `json:"start_location"`
	EndLocation      LatLng    `json:"end_location"`
	Polyline         Polyline  `json:"polyline"`
	Duration         TextValue `json:"duration"`
	Distance         TextValue `json:"distance"`
	HtmlInstructions string    `json:"html_instructions"`
	Maneuver         string    `json:"maneuver,omitempty"`
	Steps            []Step    `json:"steps,omitempty"`
}

type TextValue struct {
	Text  string `json:"text"`
	Value int    `json:"value"`
}

type Polyline struct {
	Points string `json:"points"`
}

type LatLng struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type LatLngBounds struct {
	Northeast LatLng `json:"northeast"`
	Southwest LatLng `json:"southwest"`
}

type DirectionsResult struct {
	Status       string            `json:"status"`
	Routes       []SimplifiedRoute `json:"routes"`
	ErrorMessage string            `json:"error_message,omitempty"`
}

type SimplifiedRoute struct {
	Summary         string           `json:"summary"`
	Distance        string           `json:"distance"`
	Duration        string           `json:"duration"`
	StartAddress    string           `json:"start_address"`
	EndAddress      string           `json:"end_address"`
	Steps           []SimplifiedStep `json:"steps"`
	EncodedPolyline string           `json:"encoded_polyline"`
	Warnings        []string         `json:"warnings"`
}

type SimplifiedStep struct {
	Instruction   string `json:"instruction"`
	Distance      string `json:"distance"`
	Duration      string `json:"duration"`
	StartLocation LatLng `json:"start_location"`
	EndLocation   LatLng `json:"end_location"`
	TravelMode    string `json:"travel_mode"`
	Maneuver      string `json:"maneuver,omitempty"`
}

func GetDirections(origin, destination string, waypoints []string, mode string) (*DirectionsResult, error) {
	if mode == "" {
		mode = "driving"
	}

	apiKey := getDirectionsAPIKey()
	if apiKey == "" {
		return nil, fmt.Errorf("Google Maps API key is not configured")
	}

	url := fmt.Sprintf(
		"%s?origin=%s&destination=%s&mode=%s&key=%s",
		GoogleMapsDirectionsAPI,
		origin,
		destination,
		mode,
		apiKey,
	)

	if len(waypoints) > 0 {
		waypointsStr := strings.Join(waypoints, "|")
		url = fmt.Sprintf("%s&waypoints=optimize:true|%s", url, waypointsStr)
	}

	url = fmt.Sprintf("%s&alternatives=false&language=en&units=metric", url)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		log.Printf("Error making request to Google Maps Directions API: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading Directions API response body: %v", err)
		return nil, err
	}

	var apiResponse DirectionsResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		log.Printf("Error parsing Directions API JSON response: %v", err)
		return nil, err
	}

	if apiResponse.Status != "OK" {
		log.Printf("Google Maps Directions API returned non-OK status: %s, message: %s",
			apiResponse.Status, apiResponse.ErrorMessage)
		return &DirectionsResult{
			Status:       apiResponse.Status,
			ErrorMessage: apiResponse.ErrorMessage,
		}, nil
	}

	result := &DirectionsResult{
		Status: apiResponse.Status,
		Routes: make([]SimplifiedRoute, 0, len(apiResponse.Routes)),
	}

	for _, route := range apiResponse.Routes {
		simplifiedRoute := SimplifiedRoute{
			Summary:         route.Summary,
			Warnings:        route.Warnings,
			EncodedPolyline: route.OverviewPolyline.Points,
			Steps:           []SimplifiedStep{},
		}

		for _, leg := range route.Legs {

			if simplifiedRoute.StartAddress == "" {
				simplifiedRoute.StartAddress = leg.StartAddress
			}
			simplifiedRoute.EndAddress = leg.EndAddress

			if simplifiedRoute.Distance == "" {
				simplifiedRoute.Distance = leg.Distance.Text
			}
			if simplifiedRoute.Duration == "" {
				simplifiedRoute.Duration = leg.Duration.Text
			}

			for _, step := range leg.Steps {
				simplifiedStep := SimplifiedStep{
					Instruction:   step.HtmlInstructions,
					Distance:      step.Distance.Text,
					Duration:      step.Duration.Text,
					StartLocation: step.StartLocation,
					EndLocation:   step.EndLocation,
					TravelMode:    step.TravelMode,
					Maneuver:      step.Maneuver,
				}
				simplifiedRoute.Steps = append(simplifiedRoute.Steps, simplifiedStep)
			}
		}

		result.Routes = append(result.Routes, simplifiedRoute)
	}

	return result, nil
}

func GetDirectionsForPlanItems(items []models.PlanItem, mode string) (*DirectionsResult, error) {
	if len(items) < 2 {
		return nil, fmt.Errorf("at least two plan items with valid locations are required")
	}

	validItems := make([]models.PlanItem, 0)
	for _, item := range items {
		if item.Location != "" {
			validItems = append(validItems, item)
		}
	}

	if len(validItems) < 2 {
		return nil, fmt.Errorf("at least two plan items with valid locations are required")
	}

	origin := validItems[0].Location
	destination := validItems[len(validItems)-1].Location

	waypoints := make([]string, 0)
	for i := 1; i < len(validItems)-1; i++ {
		waypoints = append(waypoints, validItems[i].Location)
	}

	return GetDirections(origin, destination, waypoints, mode)
}
