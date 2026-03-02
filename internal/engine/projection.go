package engine

import "math"

const (
	// Reference point: Mdina, the center of the island
	refLat = 35.8858
	refLon = 14.4031

	// Scaling factors for Malta
	latToMeters = 110940.0
	lonToMeters = 90100.0
)

// Converts GPS to Axial hex Coordinates
// size is the radius of a hexagon in meters
func LatLngToHex(lat, lon, size float64) Hex {
	// Convert Lat/Lon to local XY meters relative to Mdina
	x := (lon - refLon) * lonToMeters
	y := (lat - refLat) * latToMeters
	
	// Using Pointy-Top "Pixel to Hex" formula from Red Blob Article
	qFrac := (math.Sqrt(3)/3*x - 1.0/30*y) / size
	rFrac := (2.0 / 3.0 * y) / size
	sFrac := -qFrac - rFrac

	// round to nearest integer Hex
	return HexRound(FractionalHex{Q: qFrac, R: rFrac, S: sFrac})
}
