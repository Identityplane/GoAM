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
	// Check command line args first
	args := os.Args[1:]
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "--db" || args[i] == "-d" {
			connString := args[i+1]
			logger.DebugNoContext("Using DB connection string from command line: %s", connString)
			return connString
		}
	}

	// Check environment variable second
	connString := os.Getenv("GOIAM_DB_CONN_STRING")
	if connString != "" {
		logger.DebugNoContext("Using DB connection string from environment: %s", connString)
		return connString
	}

	// Use default as last resort
	connString = "goiam.db?_foreign_keys=on"
	logger.DebugNoContext("Using default DB connection string: %s", connString)
	return connString
}

func getConfigPath() string {
	// Check command line args first
	args := os.Args[1:]
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "--config" || args[i] == "-c" {
			path := args[i+1]
			logger.DebugNoContext("Using config path from command line: %s", path)
			return path
		}
	}

	// Check environment variable second
	path := os.Getenv("GOIAM_CONFIG_PATH")
	if path != "" {
		logger.DebugNoContext("Using config path from environment: %s", path)
		return path
	}

	// Use default as last resort
	path = "../config" // fallback for local dev
	pwd, err := os.Getwd()
	if err != nil {
		logger.ErrorNoContext("Failed to get current working directory: %s", err)
	}

	logger.DebugNoContext("Using default config path: %s, current working directory: %s", path, pwd)
	return path
}

func GetNotFoundRedirectUrl() string {
	return NotFoundRedirectUrl
}

func IsXForwardedForEnabled() bool {
	return UseXForwardedFor
}

func InitConfiguration() {

	disableAdminAuthzCheck := os.Getenv("GOIAM_UNSAFE_DISABLE_ADMIN_AUTHZ_CHECK")
	if disableAdminAuthzCheck == "true" {
		logger.DebugNoContext("Disabling admin authz check")
		UnsafeDisableAdminAuthzCheck = true
	}

	NotFoundRedirectUrl = os.Getenv("GOIAM_NOT_FOUND_REDIRECT_URL")
	UseXForwardedFor = os.Getenv("GOIAM_USE_X_FORWARDED_FOR") == "true"
	EnableRequestTiming = os.Getenv("GOIAM_ENABLE_REQUEST_TIMING") == "true"
}
