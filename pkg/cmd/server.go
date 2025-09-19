package cmd

import (
	"github.com/Identityplane/GoAM/internal/logger"
	"github.com/Identityplane/GoAM/pkg"
	"github.com/Identityplane/GoAM/pkg/server_settings"
	"github.com/spf13/cobra"
)

var cfgFile string
var log = logger.GetLogger()

var rootCmd = &cobra.Command{
	Use:   "goam",
	Short: "GOAM Server",
	Long: `A server application with configurable HTTP/HTTPS listeners, 
TLS settings, and custom configuration options.

Configuration can be provided via:
- Command line flags
- Environment variables (prefixed with GOAM_)
- Configuration file (config.yaml, config.json, etc.)
- Programmatic API`,

	Run: func(cmd *cobra.Command, args []string) {

		initConfigSource()

		// Printout current working dir
		log := logger.GetLogger()

		// Load settings from all sources
		settings, err := server_settings.InitWithViper()
		if err != nil {
			log.Panic().Err(err).Msg("failed to initialize settings")
		}

		log.Info().Msg(settings.Banner)

		// Start the server
		pkg.Run(settings)
	},
}

func init() {

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		`Configuration file path (supports .yaml, .json, .toml formats)
If not specified, looks for 'config.yaml' in current directory, 
$HOME/.goam/, or /etc/goam/`)

	err := server_settings.BindCobraFlags(rootCmd)
	if err != nil {
		log.Panic().Err(err).Msg("failed to bind flags")
	}
}

func initConfigSource() {

	if cfgFile != "" {
		// If there is a config file via command line argument we only use this one
		server_settings.SetConfigFile(cfgFile)

	} else {
		// If there is no config file set via the command line we use the following path to look for a config file
		server_settings.SetDefaultSources()
	}
}

func Execute() {

	if err := rootCmd.Execute(); err != nil {
		log.Panic().Err(err).Msg("error executing command")
	}
}
