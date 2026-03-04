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

func LatLngToHex(lat, lon, size float64) Hex {
	x := (lon - refLon) * lonToMeters
	y := (lat - refLat) * latToMeters
	
	// Flat-Top "Pixel to Hex" formula
	qFrac := (2.0 / 3.0 * x) / size
	rFrac := (-1.0/3.0*x + math.Sqrt(3.0)/3.0*y) / size
	sFrac := -qFrac - rFrac

	return HexRound(FractionalHex{Q: qFrac, R: rFrac, S: sFrac})
}
