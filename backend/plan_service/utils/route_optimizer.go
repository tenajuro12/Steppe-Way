package utils

import (
	"math"
	"plan_service/internal/models"
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
	if len(items) <= 2 {
		return items
	}

	points := make([]Point, len(items))
	for i, item := range items {
		lat, lng := parseLocation(item.Location)
		points[i] = Point{
			Lat:     lat,
			Lng:     lng,
			ItemID:  item.ItemID,
			Type:    item.ItemType,
			Title:   item.Title,
			Address: item.Address,
		}
	}

	optimizedIndices := nearestNeighbor(points)

	result := make([]models.PlanItem, len(items))
	for i, idx := range optimizedIndices {
		result[i] = items[idx]
		result[i].OrderIndex = i + 1
	}

	return result
}

func parseLocation(location string) (float64, float64) {
	// Dummy implementation - would be replaced with actual geocoding
	return 0.0, 0.0
}

func nearestNeighbor(points []Point) []int {
	n := len(points)
	visited := make([]bool, n)
	path := make([]int, n)

	current := 0
	path[0] = current
	visited[current] = true

	for i := 1; i < n; i++ {
		nextIdx := -1
		minDist := math.MaxFloat64

		for j := 0; j < n; j++ {
			if !visited[j] {
				dist := calcDistance(points[current], points[j])
				if dist < minDist {
					minDist = dist
					nextIdx = j
				}
			}
		}

		current = nextIdx
		path[i] = current
		visited[current] = true
	}

	return path
}
