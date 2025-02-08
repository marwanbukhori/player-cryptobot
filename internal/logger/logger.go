package logger

import (
	"log"
	"os"
)

type Logger struct {
	info  *log.Logger
	error *log.Logger
	debug *log.Logger
}

func NewLogger() *Logger {
	return &Logger{
		info:  log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		error: log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
		debug: log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

func (l *Logger) Info(format string, v ...interface{}) {
	l.info.Printf(format, v...)
}

func (l *Logger) Error(format string, v ...interface{}) {
	l.error.Printf(format, v...)
}

func (l *Logger) Debug(format string, v ...interface{}) {
	l.debug.Printf(format, v...)
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.error.Fatalf(format, v...)
}
