package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
)

type Logger struct{}

func (logger *Logger) Tracef(format string, v ...interface{}) {
	if os.Getenv("LOG_LEVEL") == "TRACE" {
		logger.Printf("TRACE", format, v...)
	}
}
func (logger *Logger) Trace(message string) {
	if os.Getenv("LOG_LEVEL") == "TRACE" {
		logger.Printf("TRACE", message)
	}
}

func (logger *Logger) Infof(format string, v ...interface{}) {
	logger.Printf("INFO", format, v...)
}
func (logger *Logger) Info(message string) {
	logger.Printf("INFO", message)
}

func (logger *Logger) Warnf(format string, v ...interface{}) {
	logger.Printf("WARN", format, v...)
}
func (logger *Logger) Warn(message string) {
	logger.Printf("WARN", message)
}

func (logger *Logger) Errorf(format string, v ...interface{}) {
	logger.Printf("ERROR", format, v...)
}
func (logger *Logger) Error(message string) {
	logger.Printf("ERROR", message)
}

func (logger *Logger) Fatalf(format string, v ...interface{}) {
	logger.Printf("FATAL", format, v...)
}
func (logger *Logger) Fatal(message string) {
	logger.Printf("FATAL", message)
}

type LogEvent struct {
	Timestamp string `json:"@timestamp"`
	Level     string `json:"level"`
	Caller    string `json:"caller"`
	Message   string `json:"message"`
}

func (logger *Logger) Printf(level string, format string, v ...interface{}) {
	event := &LogEvent{
		Caller:    getCaller(2),
		Level:     level,
		Timestamp: time.Now().Format(time.RFC3339),
		Message:   fmt.Sprintf(format, v...),
	}

	jsonText, err := json.Marshal(event)
	if err != nil {
		panic(err)
	}

	dest := os.Stdout

	if level == "ERROR" || level == "FATAL" {
		dest = os.Stderr
	}

	_, err = dest.WriteString(fmt.Sprintf("%s\n", jsonText))
	if err != nil {
		panic(err)
	}

	if level == "FATAL" {
		os.Exit(1)
	}
}

func getCaller(callDepth int) string {
	_, file, line, ok := runtime.Caller(callDepth + 1)
	if !ok {
		file = "???"
		line = 0
	}

	split := strings.Split(file, "/")
	file = split[len(split)-1]
	return fmt.Sprintf("%s:%d", file, line)
}
