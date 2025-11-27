package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

// Logger wraps zerolog.Logger with custom methods
type Logger struct {
	logger zerolog.Logger
}

// Config holds logger configuration
type Config struct {
	Level       string // debug, info, warn, error, fatal, panic
	Environment string // development, staging, production
}

// contextKey is used for context values
type contextKey string

const (
	// RequestIDKey is the key for request ID in context
	RequestIDKey contextKey = "request_id"
	// UserIDKey is the key for user ID in context
	UserIDKey contextKey = "user_id"
	// TraceIDKey is the key for trace ID in context
	TraceIDKey contextKey = "trace_id"
)

var (
	// Global logger instance (singleton)
	instance *Logger
	once     sync.Once
)

func init() {
	// Setup error stack marshaler
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.TimeFieldFormat = time.RFC3339Nano
}

// Init initializes the global logger with config (optional, will use defaults if not called)
func Init(cfg Config) {
	once.Do(func() {
		instance = createLogger(cfg)
	})
}

// createLogger creates a new logger instance
func createLogger(cfg Config) *Logger {
	// Set log level
	level := parseLevel(cfg.Level)
	zerolog.SetGlobalLevel(level)

	var output io.Writer = os.Stdout

	// Pretty logging for development
	if cfg.Environment == "development" {
		output = &customConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "2006-01-02 15:04:05",
		}
	}

	// Create base logger
	zlog := zerolog.New(output).
		With().
		Timestamp().
		Logger()

	logger := &Logger{
		logger: zlog,
	}

	log.Logger = zlog

	return logger
}

// customConsoleWriter implements a custom console writer for better formatting
type customConsoleWriter struct {
	Out        io.Writer
	TimeFormat string
}

func (w *customConsoleWriter) Write(p []byte) (n int, err error) {
	var evt map[string]interface{}
	if err := json.Unmarshal(p, &evt); err != nil {
		return w.Out.Write(p)
	}

	buf := &bytes.Buffer{}

	// Timestamp
	if ts, ok := evt["time"].(string); ok {
		t, _ := time.Parse(time.RFC3339Nano, ts)
		buf.WriteString(t.Format(w.TimeFormat))
		buf.WriteString("  ")
	}

	// Level with color
	if level, ok := evt["level"].(string); ok {
		buf.WriteString(colorizeLevel(level))
		buf.WriteString("  ")
	}

	// Trace ID, Request ID, User ID (jika ada) - dengan separator
	if traceID, ok := evt["trace_id"].(string); ok {
		buf.WriteString(fmt.Sprintf("[%s] ", traceID))
	}
	if requestID, ok := evt["request_id"].(string); ok {
		buf.WriteString(fmt.Sprintf("[req:%s] ", requestID))
	}
	if userID, ok := evt["user_id"].(string); ok {
		buf.WriteString(fmt.Sprintf("[user:%s] ", userID))
	}

	// Message
	if msg, ok := evt["message"].(string); ok {
		buf.WriteString(msg)
	}

	// Other fields
	skipFields := map[string]bool{
		"time": true, "level": true, "message": true,
		"trace_id": true, "request_id": true, "user_id": true,
	}

	for k, v := range evt {
		if skipFields[k] {
			continue
		}
		buf.WriteString(fmt.Sprintf(" %s:", k))
		if vStr, ok := v.(string); ok {
			buf.WriteString(vStr)
		} else {
			vJSON, _ := json.Marshal(v)
			buf.WriteString(string(vJSON))
		}
	}

	buf.WriteString("\n")
	return w.Out.Write(buf.Bytes())
}

func colorizeLevel(level string) string {
	switch level {
	case "debug":
		return "\033[36m| DEBUG |\033[0m" // Cyan
	case "info":
		return "\033[32m| INFO  |\033[0m" // Green
	case "warn":
		return "\033[33m| WARN  |\033[0m" // Yellow
	case "error":
		return "\033[31m| ERROR |\033[0m" // Red
	case "fatal":
		return "\033[35m| FATAL |\033[0m" // Magenta
	case "panic":
		return "\033[35m| PANIC |\033[0m" // Magenta
	default:
		return fmt.Sprintf("| %-6s|", level)
	}
}

// parseLevel converts string level to zerolog.Level
func parseLevel(level string) zerolog.Level {
	switch level {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	default:
		return zerolog.InfoLevel
	}
}

// Get returns the global logger instance (lazy singleton)
func Get() *Logger {
	once.Do(func() {
		// Create default logger if not initialized
		instance = createLogger(Config{
			Level:       "info",
			Environment: "development",
		})
	})
	return instance
}

// WithContext returns a logger with context values
func (l *Logger) WithContext(ctx context.Context) *Logger {
	logger := l.logger

	// Add request ID if available
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		logger = logger.With().Str("request_id", requestID).Logger()
	}

	// Add user ID if available
	if userID, ok := ctx.Value(UserIDKey).(string); ok {
		logger = logger.With().Str("user_id", userID).Logger()
	}

	// Add trace ID if available
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		logger = logger.With().Str("trace_id", traceID).Logger()
	}

	return &Logger{logger: logger}
}

// WithField returns a logger with an additional field
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{
		logger: l.logger.With().Interface(key, value).Logger(),
	}
}

// WithFields returns a logger with multiple additional fields
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	logger := l.logger.With()
	for k, v := range fields {
		logger = logger.Interface(k, v)
	}
	return &Logger{logger: logger.Logger()}
}

// WithError returns a logger with an error field
func (l *Logger) WithError(err error) *Logger {
	return &Logger{
		logger: l.logger.With().Err(err).Logger(),
	}
}

// WithStruct returns a logger with a struct serialized as fields
func (l *Logger) WithStruct(key string, value interface{}) *Logger {
	return &Logger{
		logger: l.logger.With().Interface(key, value).Logger(),
	}
}

// LogStruct logs a struct with all its fields
func (l *Logger) LogStruct(msg string, value interface{}) {
	l.logger.Info().Interface("data", value).Msg(msg)
}

// LogStructDebug logs a struct at debug level
func (l *Logger) LogStructDebug(msg string, value interface{}) {
	l.logger.Debug().Interface("data", value).Msg(msg)
}

// MarshalStructToFields converts a struct to fields that can be logged
// This is useful when you want to flatten a struct into log fields
func (l *Logger) MarshalStructToFields(value interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Marshal to JSON then unmarshal to map for easy field extraction
	data, err := json.Marshal(value)
	if err != nil {
		l.logger.Error().Err(err).Msg("Failed to marshal struct")
		return result
	}

	if err := json.Unmarshal(data, &result); err != nil {
		l.logger.Error().Err(err).Msg("Failed to unmarshal to map")
		return result
	}

	return result
}

// Debug logs a debug message
func (l *Logger) Debug(msg string) {
	l.logger.Debug().Msg(msg)
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.logger.Debug().Msgf(format, args...)
}

// Info logs an info message
func (l *Logger) Info(msg string) {
	l.logger.Info().Msg(msg)
}

// Infof logs a formatted info message
func (l *Logger) Infof(format string, args ...interface{}) {
	l.logger.Info().Msgf(format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string) {
	l.logger.Warn().Msg(msg)
}

// Warnf logs a formatted warning message
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.logger.Warn().Msgf(format, args...)
}

// Error logs an error message
func (l *Logger) Error(msg string) {
	l.logger.Error().Msg(msg)
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.logger.Error().Msgf(format, args...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string) {
	l.logger.Fatal().Msg(msg)
}

// Fatalf logs a formatted fatal message and exits
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.logger.Fatal().Msgf(format, args...)
}

// Panic logs a panic message and panics
func (l *Logger) Panic(msg string) {
	l.logger.Panic().Msg(msg)
}

// Panicf logs a formatted panic message and panics
func (l *Logger) Panicf(format string, args ...interface{}) {
	l.logger.Panic().Msgf(format, args...)
}

// HTTP logs HTTP request information
func (l *Logger) HTTP(method, path string, statusCode int, duration time.Duration, clientIP string) {
	event := l.logger.Info().
		Str("method", method).
		Str("path", path).
		Int("status", statusCode).
		Dur("duration_ms", duration).
		Str("client_ip", clientIP)

	if statusCode >= 500 {
		event = l.logger.Error().
			Str("method", method).
			Str("path", path).
			Int("status", statusCode).
			Dur("duration_ms", duration).
			Str("client_ip", clientIP)
	} else if statusCode >= 400 {
		event = l.logger.Warn().
			Str("method", method).
			Str("path", path).
			Int("status", statusCode).
			Dur("duration_ms", duration).
			Str("client_ip", clientIP)
	}

	event.Msg("HTTP request")
}

// Database logs database operations
func (l *Logger) Database(operation, query string, duration time.Duration, err error) {
	event := l.logger.Debug().
		Str("operation", operation).
		Str("query", query).
		Dur("duration_ms", duration)

	if err != nil {
		event = l.logger.Error().
			Str("operation", operation).
			Str("query", query).
			Dur("duration_ms", duration).
			Err(err)
	}

	event.Msg("Database operation")
}

// Auth logs authentication events
func (l *Logger) Auth(userID, action string, success bool, reason string) {
	event := l.logger.Info().
		Str("user_id", userID).
		Str("action", action).
		Bool("success", success)

	if !success {
		event = event.Str("reason", reason)
		event = l.logger.Warn().
			Str("user_id", userID).
			Str("action", action).
			Bool("success", success).
			Str("reason", reason)
	}

	event.Msg("Authentication event")
}

// Performance logs performance metrics
func (l *Logger) Performance(operation string, duration time.Duration, metadata map[string]interface{}) {
	event := l.logger.Info().
		Str("operation", operation).
		Dur("duration_ms", duration)

	for k, v := range metadata {
		event = event.Interface(k, v)
	}

	event.Msg("Performance metric")
}

// Global convenience functions

// Debug logs a debug message using global logger
func Debug(msg string) {
	Get().Debug(msg)
}

// Debugf logs a formatted debug message using global logger
func Debugf(format string, args ...interface{}) {
	Get().Debugf(format, args...)
}

// Info logs an info message using global logger
func Info(msg string) {
	Get().Info(msg)
}

// Infof logs a formatted info message using global logger
func Infof(format string, args ...interface{}) {
	Get().Infof(format, args...)
}

// Warn logs a warning message using global logger
func Warn(msg string) {
	Get().Warn(msg)
}

// Warnf logs a formatted warning message using global logger
func Warnf(format string, args ...interface{}) {
	Get().Warnf(format, args...)
}

// Error logs an error message using global logger
func Error(msg string) {
	Get().Error(msg)
}

// Errorf logs a formatted error message using global logger
func Errorf(format string, args ...interface{}) {
	Get().Errorf(format, args...)
}

// Fatal logs a fatal message and exits using global logger
func Fatal(msg string) {
	Get().Fatal(msg)
}

// Fatalf logs a formatted fatal message and exits using global logger
func Fatalf(format string, args ...interface{}) {
	Get().Fatalf(format, args...)
}

// WithContext returns a logger with context values using global logger
func WithContext(ctx context.Context) *Logger {
	return Get().WithContext(ctx)
}

// WithField returns a logger with an additional field using global logger
func WithField(key string, value interface{}) *Logger {
	return Get().WithField(key, value)
}

// WithFields returns a logger with multiple fields using global logger
func WithFields(fields map[string]interface{}) *Logger {
	return Get().WithFields(fields)
}

// WithError returns a logger with an error field using global logger
func WithError(err error) *Logger {
	return Get().WithError(err)
}

// WithStruct returns a logger with a struct using global logger
func WithStruct(key string, value interface{}) *Logger {
	return Get().WithStruct(key, value)
}

// LogStruct logs a struct using global logger
func LogStruct(msg string, value interface{}) {
	Get().LogStruct(msg, value)
}
