package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"plan_service/internal/models"
	"strconv"
	"strings"
	"time"
)

const (
	MaxLocationsPerRequest = 25
)

type Point struct {
	Lat     float64
	Lng     float64
	ItemID  uint
	Index   int
	Type    string
	Title   string
	Address string
}

type DistanceMatrixResponse struct {
	Status               string   `json:"status"`
	OriginAddresses      []string `json:"origin_addresses"`
	DestinationAddresses []string `json:"destination_addresses"`
	Rows                 []Row    `json:"rows"`
}

type Row struct {
	Elements []Element `json:"elements"`
}

type Element struct {
	Status   string   `json:"status"`
	Duration Duration `json:"duration"`
	Distance Distance `json:"distance"`
}

type Duration struct {
	Value int    `json:"value"`
	Text  string `json:"text"`
}

type Distance struct {
	Value int    `json:"value"`
	Text  string `json:"text"`
}

func parseLocation(location string) (float64, float64) {
	location = strings.TrimSpace(location)
	parts := strings.Split(location, ",")
	if len(parts) != 2 {
		log.Printf("Invalid location format: %s", location)
		return 0.0, 0.0
	}

	lat, err1 := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	lng, err2 := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)

	if err1 != nil || err2 != nil {
		log.Printf("Error parsing location: %s, errors: %v, %v", location, err1, err2)
		return 0.0, 0.0
	}

	if lat < -90 || lat > 90 || lng < -180 || lng > 180 {
		log.Printf("Invalid coordinates in range: %f, %f", lat, lng)
		return 0.0, 0.0
	}

	return lat, lng
}

func fetchRoadDistances(origin Point, destinations []Point) ([]float64, error) {
	if len(destinations) == 0 {
		return []float64{}, nil
	}

	destParams := make([]string, len(destinations))
	for i, dest := range destinations {
		destParams[i] = fmt.Sprintf("%f,%f", dest.Lat, dest.Lng)
	}

	url := fmt.Sprintf(
		"https://maps.googleapis.com/maps/api/distancematrix/json?origins=%f,%f&destinations=%s&mode=driving&key=%s",
		origin.Lat, origin.Lng, strings.Join(destParams, "|"), GoogleMapsApiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		log.Printf("Error making request to Google Maps API: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return nil, err
	}

	var result DistanceMatrixResponse
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("Error parsing JSON response: %v", err)
		return nil, err
	}

	if result.Status != "OK" {
		log.Printf("Google Maps API returned non-OK status: %s", result.Status)
		return nil, fmt.Errorf("API error: %s", result.Status)
	}

	distances := make([]float64, len(destinations))
	for i, element := range result.Rows[0].Elements {
		if element.Status == "OK" {

			distances[i] = float64(element.Distance.Value) / 1000.0
		} else {

			log.Printf("No route found from %s to %s (status: %s)",
				origin.Title, destinations[i].Title, element.Status)
			distances[i] = math.MaxFloat64
		}
	}

	return distances, nil
}

func calcStraightLineDistance(p1, p2 Point) float64 {
	const R = 6371

	lat1 := p1.Lat * math.Pi / 180
	lat2 := p2.Lat * math.Pi / 180

	deltaLat := (p2.Lat - p1.Lat) * math.Pi / 180
	deltaLng := (p2.Lng - p1.Lng) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*
			math.Sin(deltaLng/2)*math.Sin(deltaLng/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

func nearestNeighborWithRoadDistances(points []Point) ([]int, error) {
	n := len(points)
	if n <= 1 {
		return []int{0}, nil
	}

	log.Printf("Running nearest neighbor with road distances on %d points", n)

	path := make([]int, 0, n)
	visited := make([]bool, n)

	current := 0
	path = append(path, current)
	visited[current] = true
	log.Printf("Starting with point %d: %s", current, points[current].Title)

	for len(path) < n {
		currentPoint := points[current]

		unvisited := make([]Point, 0)
		for i, point := range points {
			if !visited[i] {
				unvisited = append(unvisited, point)
			}
		}

		if len(unvisited) == 0 {
			break
		}

		distances, err := fetchRoadDistances(currentPoint, unvisited)
		if err != nil {
			log.Printf("Error fetching road distances, falling back to straight-line: %v", err)

			minDist := math.MaxFloat64
			nextIdx := -1

			for i, point := range points {
				if !visited[i] {
					dist := calcStraightLineDistance(currentPoint, point)
					if dist < minDist {
						minDist = dist
						nextIdx = i
					}
				}
			}

			if nextIdx != -1 {
				current = nextIdx
				path = append(path, current)
				visited[current] = true
				log.Printf("Next nearest point (straight-line): %d: %s (distance: %.2f km)",
					current, points[current].Title, minDist)
			}
			continue
		}

		minDist := math.MaxFloat64
		var nextPoint Point
		nextIdx := -1

		for i, dist := range distances {
			if dist < minDist {
				minDist = dist
				nextPoint = unvisited[i]

				for j, p := range points {
					if p.ItemID == nextPoint.ItemID && !visited[j] {
						nextIdx = j
						break
					}
				}
			}
		}

		if nextIdx != -1 {
			current = nextIdx
			path = append(path, current)
			visited[current] = true
			log.Printf("Next nearest point (road distance): %d: %s (distance: %.2f km)",
				current, points[current].Title, minDist)
		} else {
			log.Printf("Warning: Could not find next nearest point, using fallback")

			for i, v := range visited {
				if !v {
					current = i
					path = append(path, current)
					visited[current] = true
					log.Printf("Fallback to next unvisited point: %d: %s",
						current, points[current].Title)
					break
				}
			}
		}
	}

	return path, nil
}

func OptimizeRoute(items []models.PlanItem) []models.PlanItem {
	if len(items) <= 1 {
		return items
	}

	validPoints := make([]Point, 0, len(items))
	validIndices := make([]int, 0, len(items))
	invalidItems := make([]models.PlanItem, 0)

	for i, item := range items {
		lat, lng := parseLocation(item.Location)
		if lat != 0.0 || lng != 0.0 {
			point := Point{
				Lat:     lat,
				Lng:     lng,
				ItemID:  item.ID,
				Index:   i,
				Type:    item.ItemType,
				Title:   item.Title,
				Address: item.Address,
			}
			validPoints = append(validPoints, point)
			validIndices = append(validIndices, i)
			log.Printf("Valid point %d: %s at (%f, %f)", i, item.Title, lat, lng)
		} else {
			log.Printf("Skipping item with invalid location: %s at %s", item.Title, item.Location)
			invalidItems = append(invalidItems, item)
		}
	}

	if len(validPoints) <= 1 {
		log.Printf("Not enough valid points to optimize (found %d)", len(validPoints))
		return items
	}

	optimizedIndices, err := nearestNeighborWithRoadDistances(validPoints)
	if err != nil {
		log.Printf("Error in road distance optimization: %v, falling back to straight-line", err)

		optimizedIndices = nearestNeighbor(validPoints)
	}

	log.Printf("Optimized indices: %v", optimizedIndices)

	result := make([]models.PlanItem, 0, len(items))

	for i, idx := range optimizedIndices {
		originalIdx := validIndices[idx]
		item := items[originalIdx]
		item.OrderIndex = i + 1
		result = append(result, item)
		log.Printf("Item %d (%s) now has order_index %d",
			item.ID, item.Title, item.OrderIndex)
	}

	for i, item := range invalidItems {
		item.OrderIndex = len(result) + i + 1
		result = append(result, item)
		log.Printf("Added invalid item %d (%s) with order_index %d",
			item.ID, item.Title, item.OrderIndex)
	}

	return result
}

func nearestNeighbor(points []Point) []int {
	n := len(points)
	if n <= 1 {
		return []int{0}
	}

	log.Printf("Running legacy nearest neighbor on %d points", n)

	visited := make([]bool, n)
	path := make([]int, n)

	current := 0
	path[0] = current
	visited[current] = true
	log.Printf("Starting with point %d: %s", current, points[current].Title)

	for i := 1; i < n; i++ {
		nextIdx := -1
		minDist := math.MaxFloat64

		for j := 0; j < n; j++ {
			if !visited[j] {
				dist := calcStraightLineDistance(points[current], points[j])
				log.Printf("Distance from %s to %s: %f km",
					points[current].Title, points[j].Title, dist)

				if dist < minDist {
					minDist = dist
					nextIdx = j
				}
			}
		}

		if nextIdx == -1 {
			log.Printf("Warning: Could not find next nearest point, using fallback")

			for j := 0; j < n; j++ {
				if !visited[j] {
					nextIdx = j
					break
				}
			}
		}

		log.Printf("Next nearest point is %d: %s (distance: %f km)",
			nextIdx, points[nextIdx].Title, minDist)

		current = nextIdx
		path[i] = current
		visited[current] = true
	}

	return path
}
