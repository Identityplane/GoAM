package logger_test

import (
	"testing"

	"github.com/Identityplane/GoAM/internal/logger"
)

func TestLogger_GetLogger(t *testing.T) {
	// Test that GetLogger returns a valid logger
	log := logger.GetGoamLogger()
	// Just test that we can call methods on it
	log.Info().Msg("test message")
}

func TestLogger_Logging(t *testing.T) {
	// Test that logging works without crashing
	log := logger.GetGoamLogger()

	// Test different log levels
	log.Debug().Msg("debug message")
	log.Info().Msg("info message")
	log.Warn().Msg("warn message")
	log.Error().Msg("error message")

	// Test with fields
	log.Info().
		Str("key", "value").
		Int("number", 42).
		Msg("message with fields")

	// If we get here without panic, the test passes
}

func TestLogger_SetLogLevel(t *testing.T) {
	// Test that SetLogLevel works
	logger.SetLogLevel("debug")
	log := logger.GetGoamLogger()

	// Test that debug logging works
	log.Debug().Msg("debug test")

	// If we get here without panic, the test passes
}
