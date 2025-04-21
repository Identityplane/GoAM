package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

// LogFuncType defines the type for the logging function
type LogFuncType func(v ...interface{})

// PanicFuncType defines the type for panic handling
type PanicFuncType func(v ...interface{})

// LogFunc is the function used to log output
var LogFunc LogFuncType = func(v ...interface{}) {
	fmt.Fprintln(os.Stdout, v...)
}

// PanicFunc is the function used to handle panic
var PanicFunc PanicFuncType = log.Panicln

// LogEntry defines the structure of a log
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	TraceId   string                 `json:"trace_id,omitempty"`
	Message   string                 `json:"message,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// customLog formats and outputs a structured JSON log
func customLog(level, traceId, message string, fields map[string]interface{}) {
	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     level,
		TraceId:   traceId,
		Message:   message,
		Fields:    fields,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		LogFunc(fmt.Sprintf("LOGGING ERROR [%s] traceId=%s %s", level, traceId, message))
		return
	}

	LogFunc(string(data))
}

// Info logs a basic info message with formatting
func Info(traceId, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	customLog("INFO", traceId, message, nil)
}

// InfoNoContext logs an info message without trace id
func InfoNoContext(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	customLog("INFO", "", message, nil)
}

// InfoWithFields logs an info message with structured fields
func InfoWithFields(traceId, message string, fields map[string]interface{}) {
	customLog("INFO", traceId, message, fields)
}

func InfoWithFieldsNoContext(message string, fields map[string]interface{}) {
	customLog("INFO", "", message, fields)
}

// Debug logs a debug message
func Debug(traceId, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	customLog("DEBUG", traceId, message, nil)
}

func DebugNoContext(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	customLog("DEBUG", "", message, nil)
}

// Error logs an error message
func Error(traceId, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	customLog("ERROR", traceId, message, nil)
}

func ErrorNoContext(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	customLog("ERROR", "", message, nil)
}

// Panic logs a panic message and calls panic
func Panic(traceId, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)

	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "PANIC",
		TraceId:   traceId,
		Message:   message,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		PanicFunc(fmt.Sprintf("PANIC LOGGING ERROR traceId=%s %s", traceId, message))
		return
	}

	PanicFunc(string(data))
}

func PanicNoContext(format string, args ...interface{}) {
	Panic("", format, args...)
}

// ObjectToFields converts any struct to a map[string]interface{}
func ObjectToFields(obj interface{}) map[string]interface{} {
	var result map[string]interface{}
	data, err := json.Marshal(obj)
	if err != nil {
		return nil
	}
	json.Unmarshal(data, &result)
	return result
}
