package store

import (
	"fmt"
	"malta-dispatch/internal/engine"
	"sync"
)

type Registry struct {
	mu 		sync.RWMutex

	// cells: maps Hex to a set of Driver IDs
	// We use map[string]struct{} because it is memory efficient
	Cells	map[engine.Hex]map[string]struct{}

	Drivers map[string]engine.Hex
}

func NewRegistry() *Registry {
	return &Registry{
		Cells: make(map[engine.Hex]map[string]struct{}),
		Drivers: make(map[string]engine.Hex),
	}
}

// Moves a driver from their current hex to a nex hex
func (r *Registry) updateLocation(driverId string, lat, lon float64, mask *engine.LandMask) bool {
	newHex := engine.LatLngToHex(lat, lon, 1000.0)

	if !mask.IsLand(newHex) {
		fmt.Printf("Warning: Driver %s reported location at sea. Ignoring.\n", driverId)
		return false
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if oldHex, ok := r.Drivers[driverId]; ok {
		if oldHex == newHex {
			return true
		}

		delete(r.Cells[oldHex], driverId)

		if len(r.Cells[oldHex]) == 0 {
			delete(r.Cells, oldHex)
		}
	}

	// a map[string]struct{} occupies 0 bytes in memory, this is a Set for Go
	if r.Cells[newHex] == nil {
		r.Cells[newHex] = make(map[string]struct{})
	}

	r.Cells[newHex][driverId] = struct{}{}

	r.Drivers[driverId] = newHex
	return true
}

func (r *Registry) FindNearby(center engine.Hex, radius int) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var found []string

	searchArea := engine.GetRange(center, radius)

	// get all hexes in range
	for _, hex := range searchArea {
		if driversInHex, ok := r.Cells[hex]; ok {
			for id := range driversInHex {
				found = append(found, id)
			}
		}
	}
	return found
}

func (r *Registry) HandleDriverUpdate(driverId string, lat, lon float64, mask *engine.LandMask) bool {
	return r.updateLocation(driverId, lat, lon, mask)
}

