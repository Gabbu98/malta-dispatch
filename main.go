package main

import (
	"fmt"
	"malta-dispatch/internal/engine"
	"malta-dispatch/internal/store"
)

func main() {
	reg := store.NewRegistry()

	// simulate driver udates
	reg.HandleDriverUpdate("Driver_1", 35.921125, 14.389969)
	reg.HandleDriverUpdate("Driver_2", 35.864116, 14.534132)

	// passenger's current position
	passengerHex := engine.Hex{Q: 1, R: 2}

	nearby := reg.FindNearby(passengerHex, 10)

	fmt.Printf("Found %d drivers near the passengers: %v\n", len(nearby), nearby)
}
