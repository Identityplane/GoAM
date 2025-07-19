package config

import (
	"os"
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
	UseXForwardedFor             = false
	NotFoundRedirectUrl          = ""
	EnableRequestTiming          = false
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

func IsXForwardedForEnabled() bool {
	return UseXForwardedFor
}

func InitConfiguration() {
	log := logger.GetLogger()

	disableAdminAuthzCheck := os.Getenv("GOIAM_UNSAFE_DISABLE_ADMIN_AUTHZ_CHECK")
	if disableAdminAuthzCheck == "true" {
		log.Debug().Msg("disabling admin authz check")
		UnsafeDisableAdminAuthzCheck = true
	}

	NotFoundRedirectUrl = os.Getenv("GOIAM_NOT_FOUND_REDIRECT_URL")
	UseXForwardedFor = os.Getenv("GOIAM_USE_X_FORWARDED_FOR") == "true"
	EnableRequestTiming = os.Getenv("GOIAM_ENABLE_REQUEST_TIMING") == "true"
}
