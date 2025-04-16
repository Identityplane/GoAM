package config

import (
	"goiam/internal/db/postgres_adapter"
	"goiam/internal/db/sqlite_adapter"
	"log"
	"os"
)

// Manages global configuration for the whole server
var ConfigPath = getConfigPath()
var DBConnString = getDBConnString()
var PostgresUserDB *postgres_adapter.PostgresUserDB
var SqliteUserDB *sqlite_adapter.SQLiteUserDB

// Reads database connection string from command line args, env var, or returns default
func getDBConnString() string {
	// Check command line args first
	args := os.Args[1:]
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "--db" || args[i] == "-d" {
			connString := args[i+1]
			log.Printf("Using DB connection string from command line: %s", connString)
			return connString
		}
	}

	// Check environment variable second
	connString := os.Getenv("GOIAM_DB_CONN_STRING")
	if connString != "" {
		log.Printf("Using DB connection string from environment: %s", connString)
		return connString
	}

	// Use default as last resort
	connString = "goiam.db?_foreign_keys=on"
	log.Printf("Using default DB connection string: %s", connString)
	return connString
}

func getConfigPath() string {
	// Check command line args first
	args := os.Args[1:]
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "--config" || args[i] == "-c" {
			path := args[i+1]
			log.Printf("Using config path from command line: %s", path)
			return path
		}
	}

	// Check environment variable second
	path := os.Getenv("GOIAM_CONFIG_PATH")
	if path != "" {
		log.Printf("Using config path from environment: %s", path)
		return path
	}

	// Use default as last resort
	path = "../config" // fallback for local dev
	log.Printf("Using default config path: %s", path)
	return path
}
