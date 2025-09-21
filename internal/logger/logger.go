package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// Initialize the global logger
var globalLogger zerolog.Logger

// customCallerFormat returns a formatted caller string with only package/file:line
func customCallerFormat(pc uintptr, file string, line int) string {
	// Extract just the filename from the full path
	filename := filepath.Base(file)

	// Get the function name to extract package
	fn := runtime.FuncForPC(pc)
	if fn != nil {
		funcName := fn.Name()
		// Look for the last package name in the function path
		// Function names are typically: github.com/user/repo/path/to/package.FunctionName
		parts := strings.Split(funcName, ".")
		if len(parts) >= 2 {
			// Get everything before the last dot (package path)
			pkgPath := strings.Join(parts[:len(parts)-1], ".")
			// Split by "/" and get the last part (package name)
			pkgParts := strings.Split(pkgPath, "/")
			if len(pkgParts) > 0 {
				pkgName := pkgParts[len(pkgParts)-1]
				return fmt.Sprintf("%s/%s:%d", pkgName, filename, line)
			}
		}
	}

	// Fallback to just filename if we can't parse the package
	return fmt.Sprintf("%s:%d", filename, line)
}

func init() {
	// Configure zerolog for pretty console output with timestamp
	zerolog.TimeFieldFormat = time.RFC3339

	// Create console writer with custom caller format
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
		FormatCaller: func(i interface{}) string {
			if c, ok := i.(string); ok {
				// Parse the caller string and format it
				// The caller string is typically in format: /full/path/to/file.go:line
				parts := strings.Split(c, ":")
				if len(parts) == 2 {
					filePath := parts[0]
					line := parts[1]

					// Extract just the filename
					filename := filepath.Base(filePath)

					// Try to extract package name from the path
					// Look for internal/ or pkg/ in the path
					if strings.Contains(filePath, "/internal/") {
						internalParts := strings.Split(filePath, "/internal/")
						if len(internalParts) == 2 {
							pkgPath := internalParts[1]
							pkgParts := strings.Split(pkgPath, "/")
							if len(pkgParts) >= 2 {
								pkgName := pkgParts[0]
								return fmt.Sprintf("%s/%s:%s", pkgName, filename, line)
							}
						}
					} else if strings.Contains(filePath, "/pkg/") {
						pkgParts := strings.Split(filePath, "/pkg/")
						if len(pkgParts) == 2 {
							pkgPath := pkgParts[1]
							pkgParts := strings.Split(pkgPath, "/")
							if len(pkgParts) >= 2 {
								pkgName := pkgParts[0]
								return fmt.Sprintf("%s/%s:%s", pkgName, filename, line)
							}
						}
					}

					// Fallback to just filename
					return fmt.Sprintf("%s:%s", filename, line)
				}
				return c
			}
			return ""
		},
	}

	globalLogger = zerolog.New(consoleWriter).
		Level(zerolog.TraceLevel).
		With().
		Timestamp().
		CallerWithSkipFrameCount(3).
		Int("pid", os.Getpid()).
		Logger()

	globalLogger.Debug().Msg("logger initialized")
}

// GetGoamLogger returns the global logger instance for direct zerolog usage
func GetGoamLogger() zerolog.Logger {
	return globalLogger
}

// SetLogLevel sets the global log level
func SetLogLevel(level string) {
	var logLevel zerolog.Level
	switch level {
	case "trace":
		logLevel = zerolog.TraceLevel
	case "debug":
		logLevel = zerolog.DebugLevel
	case "info":
		logLevel = zerolog.InfoLevel
	case "warn":
		logLevel = zerolog.WarnLevel
	case "error":
		logLevel = zerolog.ErrorLevel
	case "fatal":
		logLevel = zerolog.FatalLevel
	case "panic":
		logLevel = zerolog.PanicLevel
	default:
		logLevel = zerolog.InfoLevel
	}

	// Create console writer with custom caller format
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
		FormatCaller: func(i interface{}) string {
			if c, ok := i.(string); ok {
				// Parse the caller string and format it
				// The caller string is typically in format: /full/path/to/file.go:line
				parts := strings.Split(c, ":")
				if len(parts) == 2 {
					filePath := parts[0]
					line := parts[1]

					// Extract just the filename
					filename := filepath.Base(filePath)

					// Try to extract package name from the path
					// Look for internal/ or pkg/ in the path
					if strings.Contains(filePath, "/internal/") {
						internalParts := strings.Split(filePath, "/internal/")
						if len(internalParts) == 2 {
							pkgPath := internalParts[1]
							pkgParts := strings.Split(pkgPath, "/")
							if len(pkgParts) >= 2 {
								pkgName := pkgParts[0]
								return fmt.Sprintf("%s/%s:%s", pkgName, filename, line)
							}
						}
					} else if strings.Contains(filePath, "/pkg/") {
						pkgParts := strings.Split(filePath, "/pkg/")
						if len(pkgParts) == 2 {
							pkgPath := pkgParts[1]
							pkgParts := strings.Split(pkgPath, "/")
							if len(pkgParts) >= 2 {
								pkgName := pkgParts[0]
								return fmt.Sprintf("%s/%s:%s", pkgName, filename, line)
							}
						}
					}

					// Fallback to just filename
					return fmt.Sprintf("%s:%s", filename, line)
				}
				return c
			}
			return ""
		},
	}

	// Create a new logger with the updated level
	globalLogger = zerolog.New(consoleWriter).
		Level(logLevel).
		With().
		Timestamp().
		CallerWithSkipFrameCount(3).
		Int("pid", os.Getpid()).
		Logger()
}
