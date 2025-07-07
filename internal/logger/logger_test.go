package logger_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/gianlucafrei/GoAM/internal/logger"
)

func TestLogger_InfoFormatting(t *testing.T) {
	var captured []string
	logger.LogFunc = func(v ...interface{}) {
		captured = append(captured, v[0].(string))
	}
	defer func() { logger.LogFunc = nil }()

	logger.Info("trace-123", "User %s logged in", "bro123")

	if len(captured) != 1 {
		t.Fatalf("Expected 1 log output, got %d", len(captured))
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(captured[0]), &parsed); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	if parsed["level"] != "INFO" {
		t.Errorf("Expected level INFO, got %v", parsed["level"])
	}

	if parsed["trace_id"] != "trace-123" {
		t.Errorf("Expected trace_id trace-123, got %v", parsed["trace_id"])
	}

	if !strings.Contains(parsed["message"].(string), "User bro123 logged in") {
		t.Errorf("Expected formatted message, got: %v", parsed["message"])
	}
}

func TestLogger_InfoWithFields(t *testing.T) {
	var captured []string
	logger.LogFunc = func(v ...interface{}) {
		captured = append(captured, v[0].(string))
	}
	defer func() { logger.LogFunc = nil }()

	type Order struct {
		OrderID string  `json:"order_id"`
		Amount  float64 `json:"amount"`
		Status  string  `json:"status"`
	}

	order := Order{
		OrderID: "ORD789",
		Amount:  49.99,
		Status:  "shipped",
	}

	logger.InfoWithFields("trace-456", "Order processed", logger.ObjectToFields(order))

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(captured[0]), &parsed); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	fields := parsed["fields"].(map[string]interface{})
	if fields["order_id"] != "ORD789" {
		t.Errorf("Expected order_id ORD789, got %v", fields["order_id"])
	}
	if fields["status"] != "shipped" {
		t.Errorf("Expected status shipped, got %v", fields["status"])
	}
}

func TestLogger_Debug(t *testing.T) {
	var captured []string
	logger.LogFunc = func(v ...interface{}) {
		captured = append(captured, v[0].(string))
	}
	defer func() { logger.LogFunc = nil }()

	logger.Debug("trace-debug", "Debugging value %d", 42)

	if len(captured) != 1 {
		t.Fatalf("Expected 1 debug log, got %d", len(captured))
	}

	var parsed map[string]interface{}
	json.Unmarshal([]byte(captured[0]), &parsed)

	if parsed["level"] != "DEBUG" {
		t.Errorf("Expected level DEBUG, got %v", parsed["level"])
	}
}

func TestLogger_PanicFormat(t *testing.T) {
	var captured string
	var panicked bool

	logger.PanicFunc = func(v ...interface{}) {
		captured = v[0].(string)
		panicked = true
	}
	defer func() { logger.PanicFunc = nil }()

	err := errors.New("panic error bro")
	logger.PanicNoContext("Critical: %v", err)

	if !panicked {
		t.Fatal("Expected panic to trigger but it didn't")
	}

	if !strings.Contains(captured, "panic error bro") {
		t.Errorf("Expected captured panic log to contain error, got: %s", captured)
	}
}

func TestLogger_TraceIdIncluded(t *testing.T) {
	var captured []string
	logger.LogFunc = func(v ...interface{}) {
		captured = append(captured, v[0].(string))
	}
	defer func() { logger.LogFunc = nil }()

	traceId := "trace-bro-999"
	logger.Info(traceId, "Testing trace id")

	if len(captured) != 1 {
		t.Fatalf("Expected 1 log output, got %d", len(captured))
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(captured[0]), &parsed); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	gotTraceId, ok := parsed["trace_id"].(string)
	if !ok {
		t.Fatalf("Expected trace_id to be string, but missing or wrong type")
	}

	if gotTraceId != traceId {
		t.Errorf("Expected trace_id %s, got %s", traceId, gotTraceId)
	}

	if parsed["level"] != "INFO" {
		t.Errorf("Expected level INFO, got %v", parsed["level"])
	}
}
