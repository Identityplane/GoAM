package pkg

import (
	"github.com/Identityplane/GoAM/internal"
	"github.com/Identityplane/GoAM/internal/config"
	"github.com/Identityplane/GoAM/internal/logger"
	"github.com/Identityplane/GoAM/internal/web"
	"github.com/valyala/fasthttp"
)

func Run(addr string) {

	// Init Flows
	internal.Initialize()

	// Start web adapter
	startWebAdapter(addr)
}

func SetInfrastructureAsCodeMode(mode bool) {
	config.InfrastrcutureAsCodeMode = mode
}

func SetUnsafeDisableAdminAuthzCheck(mode bool) {
	config.UnsafeDisableAdminAuthzCheck = mode
}

// startWebAdapter initializes and starts the web server
func startWebAdapter(addr string) {

	r := web.New()
	log := logger.GetLogger()

	log.Info().Msgf("server running on %s", addr)

	if err := fasthttp.ListenAndServe(addr, web.TopLevelMiddleware(r.Handler)); err != nil {
		log.Panic().Err(err).Msg("server error")
	}
}
