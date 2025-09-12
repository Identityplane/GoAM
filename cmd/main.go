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
	log := logger.GetLogger()
	log.Debug().Str("pwd", wd).Msg("starting goiam 0.0.1")

	settings := pkg.NewGoamServerSettings().
		WithListener(":8080").
		WithTls(":443", "../certificates/goam.network.crt", "../certificates/goam.network.key")

	pkg.Run(settings)
}
