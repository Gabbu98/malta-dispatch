package geojsonparsers

import (
	"encoding/json"
	"fmt"
	"malta-dispatch/internal/engine"
	"os"

	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/geojson"
)

type GeoJsonParser2 struct {

}

func NewGeoJsonParser2() *GeoJsonParser2 {
	return &GeoJsonParser2{}
}

func (g *GeoJsonParser2) ReadGeoJson(filename string) (*geom.MultiPolygon, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }

    var fc geojson.FeatureCollection
    if err := json.Unmarshal(data, &fc); err != nil {
        return nil, err
    }

    // Create a container for all valid land polygons
    combinedMP := geom.NewMultiPolygon(geom.XY)

    for _, feature := range fc.Features {
        // Only process features that are actual landmass polygons
        // malta2.geojson contains some LineStrings which cannot be 'filled'
        switch g := feature.Geometry.(type) {
        case *geom.Polygon:
            combinedMP.Push(g)
        case *geom.MultiPolygon:
            for i := 0; i < g.NumPolygons(); i++ {
                combinedMP.Push(g.Polygon(i))
            }
        }
    }

    if combinedMP.NumPolygons() == 0 {
        return nil, fmt.Errorf("no processable polygons found")
    }

    return combinedMP, nil
}

func (g *GeoJsonParser2) DetermineLimits(mp *geom.MultiPolygon) (float64, float64, float64, float64) {
	b := mp.Bounds()
	minLon, minLat, maxLon, maxLat := b.Min(0), b.Min(1), b.Max(0), b.Max(1)

	padding := 0.001
	return minLon - padding, minLat - padding, maxLon + padding, maxLat + padding
}

func (g *GeoJsonParser2) DetermineLandHexes(mp *geom.MultiPolygon, minLon, minLat, maxLon, maxLat float64) *engine.LandMask {
    landMask := make(map[engine.Hex]struct{})
    size := 1000.0
    // Density of scanning - should be smaller than your hex size to avoid missing islands
    step := 0.0005 

    coords := mp.Coords()

    for lon := minLon; lon <= maxLon; lon += step {
        for lat := minLat; lat <= maxLat; lat += step {
            // 1. Check if the point is inside the land polygons
            if g.isInside(coords, lon, lat) {
                // 2. Convert to Hex using your engine's Lat, Lon order
                hex := engine.LatLngToHex(lat, lon, size)
                
                // 3. Map handles uniqueness automatically
                landMask[hex] = struct{}{}
            }
        }
    }
    g.saveMask("results/v2/land_mask.json", landMask)
	return engine.NewLandMask(landMask)
}

func (g *GeoJsonParser2) isInside(polygons [][][]geom.Coord, lon, lat float64) bool {
	inside := false
	for _, rings := range polygons {
		for _, ring := range rings {
			j := len(ring) - 1
			for i := 0; i < len(ring); i++ {
				if (ring[i][1] > lat) != (ring[j][1] > lat) {
					intersectX := (ring[j][0]-ring[i][0])*(lat-ring[i][1])/
						(ring[j][1]-ring[i][1]) + ring[i][0]
					if lon < intersectX {
						inside = !inside
					}
				}
				j = i
			}
		}
	}
	return inside
}

func (g *GeoJsonParser2) saveMask(filename string, mask map[engine.Hex]struct{}) {
	var list []engine.Hex
	for hex := range mask {
		list = append(list, hex)
	}

	data, _ := json.Marshal(list)
	os.WriteFile(filename, data, 0644)
	fmt.Printf("Saved %d hexes to %s\n", len(list), filename)
}

// LoadPointsFromGeoJson parses a FeatureCollection and extracts Point geometries
func (g *GeoJsonParser2) LoadPointsFromGeoJson(filepath string) ([]engine.Point, error) {
	// 1. Read the file into memory
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// 2. Define the anonymous struct matching the GeoJSON schema
	var collection struct {
		Type     string `json:"type"`
		Features []struct {
			Geometry struct {
				Type        string    `json:"type"`
				Coordinates []float64 `json:"coordinates"`
			} `json:"geometry"`
		} `json:"features"`
	}

	// 3. Unmarshal the data
	if err := json.Unmarshal(data, &collection); err != nil {
		return nil, fmt.Errorf("failed to unmarshal geojson: %w", err)
	}

	// 4. Extract and validate points
	var points []engine.Point
	for i, feature := range collection.Features {
		if feature.Geometry.Type != "Point" {
			continue // Skip Lines or Polygons if any
		}

		coords := feature.Geometry.Coordinates
		// GeoJSON standard is [Longitude, Latitude]
		if len(coords) < 2 {
			return nil, fmt.Errorf("feature at index %d has invalid coordinates", i)
		}

		points = append(points, engine.Point{
			Lon: coords[0],
			Lat: coords[1],
		})
	}

	return points, nil
}

// LoadLandMaskFromGeoJSON reads a GeoJSON file and populates a LandMask
func (g *GeoJsonParser2) LoadLandMaskFromGeoJSON(filepath string) (*engine.LandMask, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read geojson: %w", err)
	}

	// Schema to match your visual_mask.geojson structure
	var collection struct {
		Features []struct {
			Properties struct {
				Q int `json:"q"`
				R int `json:"r"`
			} `json:"properties"`
		} `json:"features"`
	}

	if err := json.Unmarshal(data, &collection); err != nil {
		return nil, fmt.Errorf("failed to parse geojson: %w", err)
	}

	mask := engine.NewLandMask(nil)
	for _, f := range collection.Features {
		// Add each hex from the geojson properties to the valid hexes map
		mask.AddLandZone(f.Properties.Q, f.Properties.R)
	}

	return mask, nil
}
//[out:json][timeout:60];
//(
  // This looks for the relations that define the islands
//  relation["place"~"island|islet"](35.78, 14.16, 36.10, 14.60);
//);
// Use 'out geom' to get the full polygon geometry
//out geom;
