package geojsonparsers

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
)

type Hex struct {
	Q int `json:"Q"`
	R int `json:"R"`
}

// IMPORTANT: These must match your engine.LatLngToHex logic exactly
const (
	Size      = 1000.0   // meters
	OriginLat = 35.8857 // The center point used in your engine
	OriginLon = 14.4031
)

type Visualizer struct {

}

func (v *Visualizer) Visualize() {
	// 1. Load baked hexes
	data, err := os.ReadFile("results/v2/land_mask.json")
	if err != nil {
		fmt.Println("Could not read land_mask.json:", err)
		return
	}

	var hexes []Hex
	if err := json.Unmarshal(data, &hexes); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	// 2. Build standard GeoJSON structure
	type Geometry struct {
		Type        string         `json:"type"`
		Coordinates [][][]float64 `json:"coordinates"` // Note the triple nesting for Polygon
	}
	type Feature struct {
		Type     string   `json:"type"`
		Geometry Geometry `json:"geometry"`
		Properties map[string]interface{} `json:"properties"`
	}
	type FeatureCollection struct {
		Type     string    `json:"type"`
		Features []Feature `json:"features"`
	}

	fc := FeatureCollection{
		Type:     "FeatureCollection",
		Features: make([]Feature, 0, len(hexes)),
	}

	for _, h := range hexes {
		// Convert Q,R back to Lat/Lon center
		// This uses the "Pointy Top" axial inverse math
		x := Size * 3.0/2.0 * float64(h.Q)
		y := Size * math.Sqrt(3.0) * (float64(h.R) + float64(h.Q)/2.0) 

		centerLat := OriginLat + (y / 111111.0)
		centerLon := OriginLon + (x / (111111.0 * math.Cos(OriginLat*math.Pi/180.0)))

		// Create 6 corners of the hexagon
		var ring [][]float64
		for i := 0; i < 7; i++ {
			// Start at 30 degrees for pointy top
			angle := (float64(i) * 60.0) * math.Pi / 180.0
			
			dx := Size * math.Cos(angle)
			dy := Size * math.Sin(angle)

			pLat := centerLat + (dy / 111111.0)
			pLon := centerLon + (dx / (111111.0 * math.Cos(centerLat*math.Pi/180.0)))
			
			// GeoJSON uses [Longitude, Latitude]
			ring = append(ring, []float64{pLon, pLat})
		}

		fc.Features = append(fc.Features, Feature{
			Type: "Feature",
			Properties: map[string]interface{}{
				"q": h.Q,
				"r": h.R,
			},
			Geometry: Geometry{
				Type:        "Polygon",
				Coordinates: [][][]float64{ring},
			},
		})
	}

	// 3. Save file
	result, _ := json.MarshalIndent(fc, "", "  ")
	os.WriteFile("results/v2/visual_germany_mask.geojson", result, 0644)
	fmt.Printf("Successfully generated visual_mask.geojson with %d hexes\n", len(hexes))
}
