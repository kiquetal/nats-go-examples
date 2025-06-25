// Package logger provides internal logging functionality for the application
package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

// Level represents the logging level
type Level int

const (
	// DEBUG level for detailed information
	DEBUG Level = iota
	// INFO level for general information
	INFO
	// WARN level for warning messages
	WARN
	// ERROR level for error messages
	ERROR
	// FATAL level for fatal errors that lead to program termination
	FATAL
)

var levelNames = map[Level]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
	FATAL: "FATAL",
}

// Logger represents a custom logger instance
type Logger struct {
	level     Level
	logger    *log.Logger
	component string
}

// NewLogger creates a new logger instance
func NewLogger(component string, level Level, output io.Writer) *Logger {
	if output == nil {
		output = os.Stdout
	}

	return &Logger{
		level:     level,
		logger:    log.New(output, "", 0),
		component: component,
	}
}

// DefaultLogger creates a new logger with default settings
func DefaultLogger(component string) *Logger {
	return NewLogger(component, INFO, os.Stdout)
}

func (l *Logger) log(level Level, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	msg := fmt.Sprintf(format, args...)
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	levelName := levelNames[level]

	l.logger.Printf("[%s] [%s] [%s] %s", timestamp, levelName, l.component, msg)

	if level == FATAL {
		os.Exit(1)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// Fatal logs a fatal message and exits the program
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(FATAL, format, args...)
}
