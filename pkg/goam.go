package pkg

import (
	"github.com/Identityplane/GoAM/internal"
	"github.com/Identityplane/GoAM/internal/config"
	"github.com/Identityplane/GoAM/internal/logger"
	"github.com/Identityplane/GoAM/internal/web"
	"github.com/valyala/fasthttp"
)

func Run(settings *GoamServerSettings) {

	// Init Flows
	internal.Initialize()

	// Start web adapter
	startWebAdapter(settings)
}

func SetInfrastructureAsCodeMode(mode bool) {
	config.InfrastrcutureAsCodeMode = mode
}

func SetUnsafeDisableAdminAuthzCheck(mode bool) {
	config.UnsafeDisableAdminAuthzCheck = mode
}

// startWebAdapter initializes and starts the web server
func startWebAdapter(settings *GoamServerSettings) {

	r := web.New()
	log := logger.GetLogger()

	if settings.ListenerHTTPS != "" {
		if err := fasthttp.ListenAndServeTLS(settings.ListenerHTTPS, settings.TlsCertFile, settings.TlsKeyFile, web.TopLevelMiddleware(r.Handler)); err != nil {
			log.Panic().Err(err).Msg("server error")
		}
		log.Info().Msgf("https server running on %s", settings.ListenerHTTPS)
	}

	if settings.ListenerHttp != "" {
		if err := fasthttp.ListenAndServe(settings.ListenerHttp, web.TopLevelMiddleware(r.Handler)); err != nil {
			log.Panic().Err(err).Msg("server error")
		}

		log.Info().Msgf("http server running on %s", settings.ListenerHttp)
	}
}
