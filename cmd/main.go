package main

import (
	"goiam/internal" // lint:ignore ST1019 (This should be fixed, but is not a priority)
	// lint:ignore ST1019
	"log"
	"os"
)

func main() {

	// Printout current working dir
	wd, _ := os.Getwd()
	log.Printf("Starting GoIAM 0.0.1 with pwd: %s\n", wd)

	// Init Flows
	internal.Initialize()
}
