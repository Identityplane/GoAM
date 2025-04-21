package main

import (
	"goiam/internal" // lint:ignore ST1019 (This should be fixed, but is not a priority)
	"goiam/internal/logger"
	// lint:ignore ST1019

	"os"
)

func main() {

	// Printout current working dir
	wd, _ := os.Getwd()
	logger.DebugNoContext("Starting GoIAM 0.0.1 with pwd: %s\n", wd)

	// Init Flows
	internal.Initialize()
}
