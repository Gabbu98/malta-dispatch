package main

import (
	"fmt"
	"malta-dispatch/internal/engine"
	"malta-dispatch/internal/store"
)

func main() {
	reg := store.NewRegistry()

	// simulate driver udates
	reg.UpdateLocation("Driver_1", engine.Hex{Q: 1, R: 2})
	reg.UpdateLocation("Driver_2", engine.Hex{Q: 5, R: 3})

	// passenger's current position
	passengerHex := engine.Hex{Q: 1, R: 2}

	nearby := reg.FindNearby(passengerHex, 1)

	fmt.Printf("Found %d drivers near the passengers: %v\n", len(nearby), nearby)
}
