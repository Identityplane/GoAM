package main

import (
	// lint:ignore ST1019 (This should be fixed, but is not a priority)
	"github.com/Identityplane/GoAM/internal/logger"
	"github.com/Identityplane/GoAM/pkg"

	// lint:ignore ST1019

	"os"
)

func main() {

	// Printout current working dir
	wd, _ := os.Getwd()
	logger.DebugNoContext("Starting GoIAM 0.0.1 with pwd: %s\n", wd)

	pkg.Run()
}
