package logger

import (
	"log"
	"os"
)

// Logger represents application logger
type Logger struct {
	info  *log.Logger
	error *log.Logger
}

// New creates a new logger instance
func New() *Logger {
	return &Logger{
		info:  log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		error: log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

// Info logs info message
func (l *Logger) Info(v ...interface{}) {
	l.info.Println(v...)
}

// Infof logs formatted info message
func (l *Logger) Infof(format string, v ...interface{}) {
	l.info.Printf(format, v...)
}

// Error logs error message
func (l *Logger) Error(v ...interface{}) {
	l.error.Println(v...)
}

// Errorf logs formatted error message
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.error.Printf(format, v...)
}
