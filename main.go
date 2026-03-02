package main

import (
	"fmt"
	"malta-dispatch/internal/engine"
	"malta-dispatch/internal/store"
)

func main() {
	reg := store.NewRegistry()


	// passenger's current position
	passengerHex := engine.Hex{Q: 1, R: 2}

	nearby := reg.FindNearby(passengerHex, 10)

	fmt.Printf("Found %d drivers near the passengers: %v\n", len(nearby), nearby)
}
