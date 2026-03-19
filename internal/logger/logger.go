/*
 * Copyright 2026 Scott Walter, MMFP Solutions LLC
 *
 * This program is free software; you can redistribute it and/or modify it
 * under the terms of the GNU General Public License as published by the Free
 * Software Foundation; either version 3 of the License, or (at your option)
 * any later version.  See LICENSE for more details.
 */

package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

// Module constants
const (
	ModuleMain       = "MAIN"
	ModuleConfig     = "CONFIG"
	ModuleHandler    = "HANDLER"
	ModuleService    = "SERVICE"
	ModuleMiddleware = "MIDDLEWARE"
	ModuleWeb        = "WEB"
)

// Log levels
const (
	LevelDebug = 0
	LevelInfo  = 1
	LevelWarn  = 2
	LevelError = 3
	LevelFatal = 4
)

var (
	globalLevel int = LevelInfo
	levelMu     sync.RWMutex
	logFile     *os.File
	logWriter   io.Writer = os.Stdout
	writerMu    sync.RWMutex
)

// SetGlobalLevel sets the minimum log level from a string
func SetGlobalLevel(level string) {
	levelMu.Lock()
	defer levelMu.Unlock()
	switch strings.ToUpper(level) {
	case "DEBUG":
		globalLevel = LevelDebug
	case "INFO":
		globalLevel = LevelInfo
	case "WARN", "WARNING":
		globalLevel = LevelWarn
	case "ERROR":
		globalLevel = LevelError
	case "FATAL":
		globalLevel = LevelFatal
	default:
		globalLevel = LevelInfo
	}
}

func getGlobalLevel() int {
	levelMu.RLock()
	defer levelMu.RUnlock()
	return globalLevel
}

// SetupFileLogging enables logging to a file in addition to stdout
func SetupFileLogging(enabled bool, filePath string) error {
	writerMu.Lock()
	defer writerMu.Unlock()

	if !enabled || filePath == "" {
		logWriter = os.Stdout
		return nil
	}

	// Ensure directory exists
	dir := filePath[:strings.LastIndex(filePath, "/")]
	if dir != "" {
		os.MkdirAll(dir, 0755)
	}

	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file %s: %w", filePath, err)
	}

	logFile = f
	logWriter = io.MultiWriter(os.Stdout, f)
	return nil
}

// CloseLogFile closes the log file if open
func CloseLogFile() {
	writerMu.Lock()
	defer writerMu.Unlock()
	if logFile != nil {
		logFile.Close()
		logFile = nil
		logWriter = os.Stdout
	}
}

// Logger provides structured logging for a specific module
type Logger struct {
	module string
}

// New creates a new Logger for the given module
func New(module string) *Logger {
	return &Logger{module: module}
}

func (l *Logger) output(level int, levelStr, format string, args ...interface{}) {
	if level < getGlobalLevel() {
		return
	}
	msg := fmt.Sprintf(format, args...)
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	line := fmt.Sprintf("[%s] [%s] [%s] %s\n", timestamp, l.module, levelStr, msg)

	writerMu.RLock()
	fmt.Fprint(logWriter, line)
	writerMu.RUnlock()
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.output(LevelDebug, "DEBUG", format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.output(LevelInfo, "INFO", format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.output(LevelWarn, "WARN", format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.output(LevelError, "ERROR", format, args...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.output(LevelFatal, "FATAL", format, args...)
	log.Fatal("Fatal error — exiting")
}
