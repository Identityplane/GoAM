package main

import (
	"github.com/gianlucafrei/GoAM/internal" // lint:ignore ST1019 (This should be fixed, but is not a priority)
	"github.com/gianlucafrei/GoAM/internal/logger"
	"github.com/gianlucafrei/GoAM/internal/web"

	// lint:ignore ST1019

	"os"

	"github.com/valyala/fasthttp"
)

func main() {

	// Printout current working dir
	wd, _ := os.Getwd()
	logger.DebugNoContext("Starting GoIAM 0.0.1 with pwd: %s\n", wd)

	// Init Flows
	internal.Initialize()

	// Start web adapter
	startWebAdapter()
}

// startWebAdapter initializes and starts the web server
func startWebAdapter() {
	r := web.New()
	logger.DebugNoContext("Server running on http://localhost:8080")
	if err := fasthttp.ListenAndServe(":8080", web.TopLevelMiddleware(r.Handler)); err != nil {
		logger.PanicNoContext("Error: %s", err)
	}
}
