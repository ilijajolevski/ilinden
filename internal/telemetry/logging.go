// Logging setup and configuration
//
// Structured logging framework:
// - Log level management
// - Output formatting
// - Field standardization
// - Contextual logging

package telemetry

import (
	"context"
	"io"
	"os"
	"strings"
)

// LogLevel represents the logging level
type LogLevel int

const (
	// Log levels
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
)

// Logger defines the interface for logging
type Logger interface {
	// Log methods
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	
	// With methods
	With(args ...interface{}) Logger
	WithField(key string, value interface{}) Logger
	
	// Context methods
	WithContext(ctx context.Context) Logger
}

// SimpleLogger is a simple implementation of the Logger interface
type SimpleLogger struct {
	level  LogLevel
	writer io.Writer
	fields map[string]interface{}
}

// NewLogger creates a new logger
func NewLogger(level string, format string, output string) Logger {
	// Determine log level
	var logLevel LogLevel
	switch strings.ToLower(level) {
	case "debug":
		logLevel = LevelDebug
	case "info":
		logLevel = LevelInfo
	case "warn":
		logLevel = LevelWarn
	case "error":
		logLevel = LevelError
	default:
		logLevel = LevelInfo
	}
	
	// Determine output writer
	var writer io.Writer
	switch strings.ToLower(output) {
	case "stdout":
		writer = os.Stdout
	case "stderr":
		writer = os.Stderr
	default:
		// Could add file output here
		writer = os.Stdout
	}
	
	return &SimpleLogger{
		level:  logLevel,
		writer: writer,
		fields: make(map[string]interface{}),
	}
}

// Debug logs a debug message
func (l *SimpleLogger) Debug(msg string, args ...interface{}) {
	if l.level <= LevelDebug {
		l.log("DEBUG", msg, args...)
	}
}

// Info logs an info message
func (l *SimpleLogger) Info(msg string, args ...interface{}) {
	if l.level <= LevelInfo {
		l.log("INFO", msg, args...)
	}
}

// Warn logs a warning message
func (l *SimpleLogger) Warn(msg string, args ...interface{}) {
	if l.level <= LevelWarn {
		l.log("WARN", msg, args...)
	}
}

// Error logs an error message
func (l *SimpleLogger) Error(msg string, args ...interface{}) {
	if l.level <= LevelError {
		l.log("ERROR", msg, args...)
	}
}

// With adds fields to the logger
func (l *SimpleLogger) With(args ...interface{}) Logger {
	// Create a new logger with the same level and writer
	newLogger := &SimpleLogger{
		level:  l.level,
		writer: l.writer,
		fields: make(map[string]interface{}),
	}
	
	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}
	
	// Add new fields
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			key, ok := args[i].(string)
			if ok {
				newLogger.fields[key] = args[i+1]
			}
		}
	}
	
	return newLogger
}

// WithField adds a field to the logger
func (l *SimpleLogger) WithField(key string, value interface{}) Logger {
	return l.With(key, value)
}

// WithContext adds context to the logger
func (l *SimpleLogger) WithContext(ctx context.Context) Logger {
	// In a real implementation, this would extract values from the context
	return l
}

// log logs a message with the given level
func (l *SimpleLogger) log(level, msg string, args ...interface{}) {
	// In a real implementation, this would format the message and fields
	// For this simple example, we just print to the writer
	// The format would depend on the format option (JSON, console, etc.)
	
	// Process args as key-value pairs
	fields := make(map[string]interface{})
	for k, v := range l.fields {
		fields[k] = v
	}
	
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			key, ok := args[i].(string)
			if ok {
				fields[key] = args[i+1]
			}
		}
	}
	
	// For this simple implementation, we just print a basic message
	// In a real implementation, this would be formatted as JSON or other format
	output := level + ": " + msg
	if len(fields) > 0 {
		output += " " + fieldsToString(fields)
	}
	output += "\n"
	
	l.writer.Write([]byte(output))
}

// fieldsToString converts fields to a string
func fieldsToString(fields map[string]interface{}) string {
	var parts []string
	for k, v := range fields {
		parts = append(parts, k+"="+toString(v))
	}
	return strings.Join(parts, " ")
}

// toString converts a value to a string
func toString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case error:
		return v.Error()
	default:
		return "<?>"
	}
}