package cmd

import (
	"os"

	"github.com/Identityplane/GoAM/internal/logger"
	"github.com/Identityplane/GoAM/pkg"
	"github.com/Identityplane/GoAM/pkg/server_settings"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

		initConfig()

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

	for _, doc := range server_settings.GetConfigDocumentation() {
		rootCmd.PersistentFlags().String(doc.Field, doc.Default, doc.Description)
		viper.BindPFlag(doc.Field, rootCmd.PersistentFlags().Lookup(doc.Field))
	}

}

func initConfig() {

	wd, err := os.Getwd()
	if err != nil {
		log.Panic().Err(err).Msg("failed to get working directory")
	}

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".cobra" (without extension).
		viper.SetConfigType("env")
		viper.AddConfigPath(home)
		viper.AddConfigPath(wd)
		viper.SetConfigType("yaml")
		viper.SetConfigName("goam.yaml")
	}

	viper.AutomaticEnv()

	err = viper.ReadInConfig()

	if err == nil {
		log.Debug().Str("config_file", viper.ConfigFileUsed()).Msg("using config file")
	}

	if err != nil {

		// Check if the file exists
		_, readFileErr := os.Stat(cfgFile)
		if os.IsNotExist(readFileErr) {
			log.Debug().Str("config_file", cfgFile).Msg("config file does not exist")
		} else {
			log.Panic().Str("config_file", cfgFile).Err(err).Msg("error reading config file")
		}

	}
}

func Execute() {

	if err := rootCmd.Execute(); err != nil {
		log.Panic().Err(err).Msg("error executing command")
	}
}
