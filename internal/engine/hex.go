 package engine

 import "math"

 type Hex struct {
	Q, R int
 }

 func (h Hex) S() int {
	return -h.Q - h.R
 }

 // for coordinate conversion and rounding 
 type FractionalHex struct {
	Q, R, S float64
 }

 func HexRound(fh FractionalHex) Hex {
	q := int(math.Round(fh.Q))
	r := int(math.Round(fh.R))
	s := int(math.Round(fh.S))

	qDiff := math.Abs(float64(q) - fh.Q)
	rDiff := math.Abs(float64(r) - fh.R)
	sDiff := math.Abs(float64(s) - fh.S)

	if qDiff > rDiff && qDiff > sDiff {
		q = -r - s
	} else {
		r = -q - s
	}

	// we only store Axial (Q, R)
	return Hex{q, r}
 }

 var directions = []Hex{
	{1, 0}, {1, -1}, {0, -1},
	{-1, 0}, {-1, 1}, {0, 1},
 }

 func (h Hex) Neighbour(direction int) Hex {
	dir := directions[direction%6]
	return Hex{h.Q + dir.Q, h.R + dir.R}
 }

 func Distance(a, b Hex) int {
	return (abs(a.Q-b.Q) + abs(a.Q+a.R) + abs(a.R-b.R)) / 2
 }

 func abs(x int) int {
	if x < 0 {return -x}
	return x
 }

 func GetRange(center Hex, dist int) []Hex {
	var results []Hex
	for q := -dist; q <= dist; q++ {
		// constraints
		rMin := max(-dist, -q-dist)
		rMax := min(dist, -q+dist)
		for r := rMin; r <= rMax; r++ {
			results = append(results, Hex{center.Q + q, center.R + r})
		}
	}
	return results
 }

 func max(a, b int) int { if a > b { return a }; return b }
 func min(a, b int) int { if a < b { return a }; return b }
