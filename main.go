package main

import (
	"encoding/json"
	"fmt"
	geojsonparsers "malta-dispatch/geojsonParsers"
	"malta-dispatch/internal/engine"
	"malta-dispatch/internal/store"
	"math/rand"
	"os"
	"time"
)

func main() {
	geojsonParser := geojsonparsers.NewGeoJsonParser2()
	landMask, err := geojsonParser.LoadLandMaskFromGeoJSON("results/v2/visual_mask.geojson")
	if err!=nil {
		fmt.Errorf("Error reading land mask due to %v", err)
	}

	drivers, err := geojsonParser.LoadPointsFromGeoJson("results/v2/less_drivers.geojson")

	if err!=nil {
		fmt.Errorf("Error reading driver points due to %v", err)
	}

	registry := store.NewRegistry()
	for i, p := range drivers {
        driverID := fmt.Sprintf("driver_%d", i)
        // This will snap the raw Lat/Lon to your 300m Hex grid
        registry.HandleDriverUpdate(driverID, p.Lat, p.Lon, landMask)
    }
	fmt.Printf("Registry populated with %d drivers across active cells.\n", len(registry.Drivers))

    customerPos := engine.LatLngToHex(35.891354, 14.440711, 1000.0)

    driversNearby := registry.FindNearby(customerPos, 5)

    if len(driversNearby) != 0 {
        fmt.Printf("Found %d Drivers.\n", len(driversNearby))
    }

	driverId, distance := registry.FindNearestNeighbours(customerPos, 5)
	fmt.Printf("Driver %s is %d km near you, dispatching.\n", driverId, distance)
}

func generateData() {
	// Seed random for different results each run
	rand.Seed(time.Now().UnixNano())

	geoJsonParser := geojsonparsers.NewGeoJsonParser2()
	mp, err := geoJsonParser.ReadGeoJson("results/v2/malta.geojson")
	if err != nil {
		fmt.Printf("Error whilst reading geojson: %v\n", err)
		return
	}

	minLon, minLat, maxLon, maxLat := geoJsonParser.DetermineLimits(mp)
	
	registry := store.NewRegistry()
	landMask := geoJsonParser.DetermineLandHexes(mp, minLon, minLat, maxLon, maxLat)
	var driverCoordinates []engine.Point

	// We only want to export drivers that actually landed on the LandMask
	for i := 0; i < 1000; i++ {
		driverId := fmt.Sprintf("Taxi-%03d", i)
		lat, lon := generateRandomCoordinates(minLon, minLat, maxLon, maxLat)
		
		// The registry handles the land check internally
		validHex := registry.HandleDriverUpdate(driverId, lat, lon, landMask)
		if (validHex) {
			driverCoordinates = append(driverCoordinates, engine.Point{
                Lat: lat,
                Lon: lon,
            })
		}
	}

	// Extract successful driver locations from registry for export
	exportDriverCoordinatesToGeoJson(driverCoordinates, "results/v2/drivers_sim.geojson")

	fmt.Println("Simulation complete. Check drivers_sim.geojson")

}

func generateRandomCoordinates(minLon, minLat, maxLon, maxLat float64) (float64, float64) {
	// rand.Float64() returns 0.0 to 1.0
	lon := minLon + rand.Float64()*(maxLon-minLon)
	lat := minLat + rand.Float64()*(maxLat-minLat)
	return lat, lon // Return Lat, Lon to match your registry's expected input
}

func exportDriverCoordinatesToGeoJson(coordinates []engine.Point, filename string) {
    type Feature struct {
        Type       string                 `json:"type"`
        Geometry   map[string]interface{} `json:"geometry"`
        Properties map[string]interface{} `json:"properties"`
    }

    type FeatureCollection struct {
        Type     string    `json:"type"`
        Features []Feature `json:"features"`
    }

    featureCollection := FeatureCollection{
        Type: "FeatureCollection",
    }

    // Use 'index, value' in range
    for _, pt := range coordinates {
        feature := Feature{
            Type: "Feature",
            Geometry: map[string]interface{}{
                "type": "Point",
                // IMPORTANT: geojson.io expects [Longitude, Latitude]
                "coordinates": []float64{pt.Lon, pt.Lat},
            },
            Properties: map[string]interface{}{
                "marker-color":  "#ff0000",
                "marker-symbol": "taxi",
            },
        }
        featureCollection.Features = append(featureCollection.Features, feature)
    }

    data, err := json.MarshalIndent(featureCollection, "", "  ")
    if err != nil {
        fmt.Printf("Error marshaling JSON: %v\n", err)
        return
    }

    // Save to file
    err = os.WriteFile(filename, data, 0644)
    if err != nil {
        fmt.Printf("Error writing file: %v\n", err)
    } else {
        fmt.Printf("Successfully exported %d points to %s\n", len(coordinates), filename)
    }
}
