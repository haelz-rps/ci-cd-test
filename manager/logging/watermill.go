package logging

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

// GoKitWatermillLogger adapts Go-Kit logger to Watermill's Logger interface
type goKitWatermillLogger struct {
	logger log.Logger
}

// NewGoKitWatermillLogger initializes the adapter with a Go-Kit logger
func NewGoKitWatermillLogger(logger log.Logger) watermill.LoggerAdapter {
	return &goKitWatermillLogger{
		logger: logger,
	}
}

// Trace logs a trace message (mapped to Debug level)
func (gk *goKitWatermillLogger) Trace(msg string, fields watermill.LogFields) {
	gk.Debug(msg, fields)
}

// Debug logs a debug message
func (gk *goKitWatermillLogger) Debug(msg string, fields watermill.LogFields) {
	l := level.Debug(gk.logger)
	l.Log(append([]interface{}{"msg", msg}, convertFields(fields)...)...)
}

// Info logs an info message
func (gk *goKitWatermillLogger) Info(msg string, fields watermill.LogFields) {
	l := level.Info(gk.logger)
	l.Log(append([]interface{}{"msg", msg}, convertFields(fields)...)...)
}

// Warn logs a warning message
func (gk *goKitWatermillLogger) Warn(msg string, fields watermill.LogFields) {
	l := level.Warn(gk.logger)
	l.Log(append([]interface{}{"msg", msg}, convertFields(fields)...)...)
}

// Error logs an error message with an error
func (gk *goKitWatermillLogger) Error(msg string, err error, fields watermill.LogFields) {
	l := level.Error(gk.logger)
	// Include the error in the fields
	if fields == nil {
		fields = make(watermill.LogFields)
	}
	fields["error"] = err.Error()
	l.Log(append([]interface{}{"msg", msg}, convertFields(fields)...)...)
}

// With adds contextual fields to the logger and returns a new Logger instance
func (gk *goKitWatermillLogger) With(fields watermill.LogFields) watermill.LoggerAdapter {
	// Convert map to key-value pairs
	kv := convertFields(fields)
	// Create a new Go-Kit logger with the additional fields
	newLogger := log.With(gk.logger, kv...)
	return &goKitWatermillLogger{
		logger: newLogger,
	}
}

// convertFields transforms Watermill's LogFields to key-value pairs for Go-Kit
func convertFields(fields map[string]interface{}) []interface{} {
	kv := make([]interface{}, 0, len(fields)*2)
	for key, value := range fields {
		kv = append(kv, key, value)
	}
	return kv
}
