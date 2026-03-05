package geojsonparsers

import (
	"encoding/json"
	"fmt"
	"malta-dispatch/internal/engine"
	"os"

	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/geojson"
)

type GeoJsonParser struct {

}

func NewGeoJsonParser() *GeoJsonParser {
	return &GeoJsonParser{}
}

func (g *GeoJsonParser) ReadGeoJson(filename string) (*geom.MultiPolygon, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var fc geojson.FeatureCollection
	err = json.Unmarshal(data, &fc)
	if err != nil {
		return nil, err
	}

	if mp, ok := fc.Features[0].Geometry.(*geom.MultiPolygon); ok {
		return mp, nil
	}

	if p, ok := fc.Features[0].Geometry.(*geom.Polygon); ok {
		mp := geom.NewMultiPolygon(p.Layout())
		err = mp.Push(p)
		if err != nil {
			return nil, err
		}
		return mp, nil
	}

	return nil, fmt.Errorf("unsupported geometry type")
}

func (g *GeoJsonParser) DetermineLimits(mp *geom.MultiPolygon) (float64, float64, float64, float64) {
	b := mp.Bounds()
	minLon, minLat, maxLon, maxLat := b.Min(0), b.Min(1), b.Max(0), b.Max(1)

	padding := 0.001
	return minLon - padding, minLat - padding, maxLon + padding, maxLat + padding
}

func (g GeoJsonParser) DetermineLandHexes(mp *geom.MultiPolygon, minLon, minLat, maxLon, maxLat float64) *engine.LandMask {
	landMask := make(map[engine.Hex]struct{})
	size := 1000.0
	step := 0.0005

	coords := mp.Coords()

	for lon := minLon; lon <= maxLon; lon += step {
		for lat := minLat; lat <= maxLat; lat += step {
			if g.isInside(coords, lon, lat) {
				hex := engine.LatLngToHex(lat, lon, size)
				// If we haven't processed this hex yet
				if _, processed := landMask[hex]; !processed {
				// Check if the CENTER of the hex is on land
				// (This is usually enough for a clean look)
					if g.isInside(coords, lon, lat) {
						landMask[hex] = struct{}{}
					}
				}
			}
		}
	}
	g.saveMask("results/v1/land_mask.json", landMask)
	return engine.NewLandMask(landMask)
}

func (g *GeoJsonParser) isInside(polygons [][][]geom.Coord, lon, lat float64) bool {
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

func (g *GeoJsonParser) saveMask(filename string, mask map[engine.Hex]struct{}) {
	var list []engine.Hex
	for hex := range mask {
		list = append(list, hex)
	}

	data, _ := json.Marshal(list)
	os.WriteFile(filename, data, 0644)
	fmt.Printf("Saved %d hexes to %s\n", len(list), filename)
}
