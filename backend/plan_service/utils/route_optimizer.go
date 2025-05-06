package utils

import (
	"log"
	"math"
	"plan_service/internal/models"
	"strconv"
	"strings"
)

type Point struct {
	Lat     float64
	Lng     float64
	ItemID  uint
	Type    string
	Title   string
	Address string
}

func calcDistance(p1, p2 Point) float64 {
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

func OptimizeRoute(items []models.PlanItem) []models.PlanItem {
	if len(items) <= 1 {
		return items
	}

	points := make([]Point, len(items))
	for i, item := range items {
		lat, lng := parseLocation(item.Location)
		points[i] = Point{
			Lat:     lat,
			Lng:     lng,
			ItemID:  item.ID,
			Type:    item.ItemType,
			Title:   item.Title,
			Address: item.Address,
		}

		// Debug locations
		log.Printf("Point %d: %s at (%f, %f)", i, item.Title, lat, lng)
	}

	optimizedIndices := nearestNeighbor(points)
	log.Printf("Optimized indices: %v", optimizedIndices)

	result := make([]models.PlanItem, len(items))
	for i, idx := range optimizedIndices {
		result[i] = items[idx]
		result[i].OrderIndex = i + 1
		log.Printf("Item %d (%s) now has order_index %d",
			result[i].ID, result[i].Title, result[i].OrderIndex)
	}

	return result
}

func parseLocation(location string) (float64, float64) {
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

	return lat, lng
}

func nearestNeighbor(points []Point) []int {
	n := len(points)
	if n <= 1 {
		return []int{0}
	}

	log.Printf("Running nearest neighbor on %d points", n)

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
				dist := calcDistance(points[current], points[j])
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
