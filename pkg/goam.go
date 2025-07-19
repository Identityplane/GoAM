package pkg

import (
	"github.com/Identityplane/GoAM/internal"
	"github.com/Identityplane/GoAM/internal/logger"
	"github.com/Identityplane/GoAM/internal/web"
	"github.com/valyala/fasthttp"
)

func Run() {

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
