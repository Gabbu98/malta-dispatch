package engine

type LandMask struct {
	ValidHexes map[Hex]struct{} // land specific hexes
}

func NewLandMask(mask map[Hex]struct{}) *LandMask {
	if mask!=nil{
		return &LandMask{ValidHexes: mask}
	}

	return &LandMask{
		ValidHexes: make(map[Hex]struct{}),
	}
}

func (m *LandMask) IsLand(h Hex) bool {
	_, ok := m.ValidHexes[h]
	return ok
}

// in reality to be loaded from a file/db
func (m *LandMask) AddLandZone(q, r int) {
	m.ValidHexes[Hex{Q: q, R: r}] = struct{}{}
}
