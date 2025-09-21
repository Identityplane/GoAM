package pkg

import (
	"sync"

	"github.com/Identityplane/GoAM/internal"
	"github.com/Identityplane/GoAM/internal/logger"
	"github.com/Identityplane/GoAM/internal/web"
	"github.com/Identityplane/GoAM/pkg/server_settings"
	"github.com/valyala/fasthttp"
)

var log = logger.GetGoamLogger()

func Run(settings *server_settings.GoamServerSettings) {

	// Init Flows
	internal.Initialize(settings)

	// Start web adapter
	startWebAdapter(settings)
}

// startWebAdapter initializes and starts the web server
func startWebAdapter(settings *server_settings.GoamServerSettings) {

	// Call all server start callbacks
	for _, callback := range serverStartCallbacks {
		if err := callback(settings); err != nil {
			log.Panic().Err(err).Msg("server start callback error")
		}
	}

	// Start the server
	r := web.New()
	var wg sync.WaitGroup

	if settings.ListenerHTTPS != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			log.Info().Msgf("https server starting on %s", settings.ListenerHTTPS)
			if err := fasthttp.ListenAndServeTLS(settings.ListenerHTTPS, settings.TlsCertFile, settings.TlsKeyFile, web.TopLevelMiddleware(r.Handler)); err != nil {
				log.Panic().Err(err).Msg("https server error")
			}
		}()
	}

	if settings.ListenerHttp != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			log.Info().Msgf("http  server starting on %s", settings.ListenerHttp)
			if err := fasthttp.ListenAndServe(settings.ListenerHttp, web.TopLevelMiddleware(r.Handler)); err != nil {
				log.Panic().Err(err).Msg("http server error")
			}
		}()
	}

	// Wait for both servers to start (or fail)
	wg.Wait()
}
