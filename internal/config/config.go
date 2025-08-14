package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/Identityplane/GoAM/internal/db/postgres_adapter"
	"github.com/Identityplane/GoAM/internal/db/sqlite_adapter"
	"github.com/Identityplane/GoAM/internal/logger"
)

// Manages global configuration for the whole server
var ConfigPath = getConfigPath()
var DBConnString = getDBConnString()
var PostgresUserDB *postgres_adapter.PostgresUserDB
var SqliteUserDB *sqlite_adapter.SQLiteUserDB

var (
	UnsafeDisableAdminAuthzCheck = false // Can be overwritten for development purposes
	NotFoundRedirectUrl          = ""
	EnableRequestTiming          = false
	InfrastrcutureAsCodeMode     = false
	ForwardingProxies            = 0
)

// Other global configurations

func GetDbDriverName() string {

	if strings.HasPrefix(DBConnString, "postgres://") {
		return "postgres"
	}
	return "sqlite"
}

// Reads database connection string from command line args, env var, or returns default
func getDBConnString() string {
	log := logger.GetLogger()

	// Check command line args first
	args := os.Args[1:]
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "--db" || args[i] == "-d" {
			connString := args[i+1]
			log.Debug().Str("conn_string", connString).Msg("using db connection string from command line")
			return connString
		}
	}

	// Check environment variable second
	connString := os.Getenv("GOIAM_DB_CONN_STRING")
	if connString != "" {
		log.Debug().Str("conn_string", connString).Msg("using db connection string from environment")
		return connString
	}

	// Use default as last resort
	connString = "goiam.db?_foreign_keys=on"
	log.Debug().Str("conn_string", connString).Msg("using default db connection string")
	return connString
}

func getConfigPath() string {
	log := logger.GetLogger()

	// Check command line args first
	args := os.Args[1:]
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "--config" || args[i] == "-c" {
			path := args[i+1]
			log.Debug().Str("config_path", path).Msg("using config path from command line")
			return path
		}
	}

	// Check environment variable second
	path := os.Getenv("GOIAM_CONFIG_PATH")
	if path != "" {
		log.Debug().Str("config_path", path).Msg("using config path from environment")
		return path
	}

	// Use default as last resort
	path = "../config" // fallback for local dev
	pwd, err := os.Getwd()
	if err != nil {
		log.Error().Err(err).Msg("failed to get current working directory")
	}

	log.Debug().Str("config_path", path).Str("pwd", pwd).Msg("using default config path")
	return path
}

func GetNotFoundRedirectUrl() string {
	return NotFoundRedirectUrl
}

func GetNumberOfProxies() int {
	return ForwardingProxies
}

func InitConfiguration() {
	log := logger.GetLogger()

	disableAdminAuthzCheck := os.Getenv("GOIAM_UNSAFE_DISABLE_ADMIN_AUTHZ_CHECK")
	if disableAdminAuthzCheck == "true" {
		UnsafeDisableAdminAuthzCheck = true
	}

	if UnsafeDisableAdminAuthzCheck {
		log.Info().Msg("disabling admin authz check")
	}

	if NotFoundRedirectUrl == "" {
		NotFoundRedirectUrl = os.Getenv("GOIAM_NOT_FOUND_REDIRECT_URL")
	}

	if os.Getenv("GOIAM_USE_X_FORWARDED_FOR") == "true" {
		log.Fatal().Msg("GOIAM_USE_X_FORWARDED_FOR is deprecated but still set to true")
	}

	if proxies := os.Getenv("GOIAM_PROXIES"); proxies != "" && ForwardingProxies == 0 {

		var err error
		ForwardingProxies, err = strconv.Atoi(proxies)
		if err != nil {
			log.Fatal().Msgf("The number of proxies specified by GOIAM_PROXIES (%v) could not be parsed: %v", proxies, err)
		}
	}

	if !EnableRequestTiming {
		EnableRequestTiming = os.Getenv("GOIAM_ENABLE_REQUEST_TIMING") == "true"
	}

	if !InfrastrcutureAsCodeMode {
		InfrastrcutureAsCodeMode = os.Getenv("GOIAM_INFRASTRUCTURE_AS_CODE_MODE") == "true"
	}

	log.Debug().
		Bool("infrastructure_as_code_mode", InfrastrcutureAsCodeMode).
		Bool("unsafe_disable_admin_authz_check", UnsafeDisableAdminAuthzCheck).
		Int("number_of_proxies", ForwardingProxies).
		Bool("enable_request_timing", EnableRequestTiming).
		Str("not_found_redirect_url", NotFoundRedirectUrl).
		Msg("loaded server configuration")
}
