// Package log provides structured logging for membrane.
package log

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Level is a log level.
type Level int

const (
	// LevelError only logs errors.
	LevelError Level = iota
	// LevelWarn logs warnings and errors.
	LevelWarn
	// LevelInfo logs info, warnings, and errors.
	LevelInfo
	// LevelDebug logs everything including debug messages.
	LevelDebug
)

var levelNames = map[Level]string{
	LevelError: "ERROR",
	LevelWarn:  "WARN",
	LevelInfo:  "INFO",
	LevelDebug: "DEBUG",
}

// Logger is a structured logger.
type Logger struct {
	mu     sync.Mutex
	level  Level
	output io.Writer
	fields map[string]interface{}
}

// defaultLogger is the global default logger.
var defaultLogger = &Logger{
	level:  LevelInfo,
	output: os.Stderr,
	fields: make(map[string]interface{}),
}

// SetLevel sets the global log level.
func SetLevel(level Level) {
	defaultLogger.mu.Lock()
	defer defaultLogger.mu.Unlock()
	defaultLogger.level = level
}

// SetOutput sets the global log output.
func SetOutput(w io.Writer) {
	defaultLogger.mu.Lock()
	defer defaultLogger.mu.Unlock()
	defaultLogger.output = w
}

// ParseLevel parses a level string.
func ParseLevel(s string) Level {
	switch s {
	case "error":
		return LevelError
	case "warn", "warning":
		return LevelWarn
	case "info":
		return LevelInfo
	case "debug":
		return LevelDebug
	default:
		return LevelInfo
	}
}

// WithField returns a new logger with the given field.
func WithField(key string, value interface{}) *Logger {
	return defaultLogger.WithField(key, value)
}

// WithFields returns a new logger with the given fields.
func WithFields(fields map[string]interface{}) *Logger {
	return defaultLogger.WithFields(fields)
}

// Error logs an error message.
func Error(msg string) {
	defaultLogger.log(LevelError, msg)
}

// Errorf logs a formatted error message.
func Errorf(format string, args ...interface{}) {
	defaultLogger.log(LevelError, fmt.Sprintf(format, args...))
}

// Warn logs a warning message.
func Warn(msg string) {
	defaultLogger.log(LevelWarn, msg)
}

// Warnf logs a formatted warning message.
func Warnf(format string, args ...interface{}) {
	defaultLogger.log(LevelWarn, fmt.Sprintf(format, args...))
}

// Info logs an info message.
func Info(msg string) {
	defaultLogger.log(LevelInfo, msg)
}

// Infof logs a formatted info message.
func Infof(format string, args ...interface{}) {
	defaultLogger.log(LevelInfo, fmt.Sprintf(format, args...))
}

// Debug logs a debug message.
func Debug(msg string) {
	defaultLogger.log(LevelDebug, msg)
}

// Debugf logs a formatted debug message.
func Debugf(format string, args ...interface{}) {
	defaultLogger.log(LevelDebug, fmt.Sprintf(format, args...))
}

// WithField returns a new logger with the given field.
func (l *Logger) WithField(key string, value interface{}) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	newFields := make(map[string]interface{}, len(l.fields)+1)
	for k, v := range l.fields {
		newFields[k] = v
	}
	newFields[key] = value

	return &Logger{
		level:  l.level,
		output: l.output,
		fields: newFields,
	}
}

// WithFields returns a new logger with the given fields.
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	newFields := make(map[string]interface{}, len(l.fields)+len(fields))
	for k, v := range l.fields {
		newFields[k] = v
	}
	for k, v := range fields {
		newFields[k] = v
	}

	return &Logger{
		level:  l.level,
		output: l.output,
		fields: newFields,
	}
}

// Error logs an error message.
func (l *Logger) Error(msg string) {
	l.log(LevelError, msg)
}

// Errorf logs a formatted error message.
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.log(LevelError, fmt.Sprintf(format, args...))
}

// Warn logs a warning message.
func (l *Logger) Warn(msg string) {
	l.log(LevelWarn, msg)
}

// Warnf logs a formatted warning message.
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.log(LevelWarn, fmt.Sprintf(format, args...))
}

// Info logs an info message.
func (l *Logger) Info(msg string) {
	l.log(LevelInfo, msg)
}

// Infof logs a formatted info message.
func (l *Logger) Infof(format string, args ...interface{}) {
	l.log(LevelInfo, fmt.Sprintf(format, args...))
}

// Debug logs a debug message.
func (l *Logger) Debug(msg string) {
	l.log(LevelDebug, msg)
}

// Debugf logs a formatted debug message.
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.log(LevelDebug, fmt.Sprintf(format, args...))
}

func (l *Logger) log(level Level, msg string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if level > l.level {
		return
	}

	timestamp := time.Now().Format(time.RFC3339)
	levelStr := levelNames[level]

	// Build field string
	fieldStr := ""
	for k, v := range l.fields {
		fieldStr += fmt.Sprintf(" %s=%v", k, v)
	}

	fmt.Fprintf(l.output, "%s [%s] %s%s\n", timestamp, levelStr, msg, fieldStr)
}
