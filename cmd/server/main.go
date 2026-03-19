/*
 * Copyright 2026 Scott Walter, MMFP Solutions LLC
 *
 * This program is free software; you can redistribute it and/or modify it
 * under the terms of the GNU General Public License as published by the Free
 * Software Foundation; either version 3 of the License, or (at your option)
 * any later version.  See LICENSE for more details.
 */

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/mmfpsolutions/gsbe/internal/config"
	"github.com/mmfpsolutions/gsbe/internal/logger"
	"github.com/mmfpsolutions/gsbe/internal/router"
	"github.com/mmfpsolutions/gsbe/internal/version"
)

func main() {
	log := logger.New(logger.ModuleMain)
	log.Info("Starting GSBE v%s (build: %s, commit: %s)", version.Version, version.BuildDate, version.Commit)

	// Resolve config directory
	configDir := resolveConfigDir()
	log.Info("Config directory: %s", configDir)

	// Load configuration
	cfgManager := config.GetManager(configDir)

	// Check if config file exists; if not, write defaults
	configPath := filepath.Join(configDir, "config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Info("No config file found, writing defaults")
		if err := cfgManager.WriteDefaultConfig(); err != nil {
			log.Fatal("Failed to write default config: %v", err)
		}
	}

	if err := cfgManager.LoadConfig(); err != nil {
		log.Fatal("Failed to load config: %v", err)
	}

	// Setup logging
	cfg := cfgManager.GetConfig()
	if cfg.Logging != nil {
		logger.SetGlobalLevel(cfg.Logging.Level)
		if err := logger.SetupFileLogging(cfg.Logging.LogToFile, cfg.Logging.LogFilePath); err != nil {
			log.Warn("Failed to setup file logging: %v", err)
		}
	}

	// Determine port
	port := cfg.Port
	if envPort := os.Getenv("PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			port = p
		}
	}

	// Setup router
	r := router.SetupRouter(cfgManager, configDir)

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Info("Server listening on port %d", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Info("Received signal %s, shutting down...", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Server forced shutdown: %v", err)
	}

	log.Info("Server stopped")
	logger.CloseLogFile()
}

func resolveConfigDir() string {
	// Check if ./config exists relative to working directory
	if info, err := os.Stat("./config"); err == nil && info.IsDir() {
		abs, err := filepath.Abs("./config")
		if err == nil {
			return abs
		}
	}

	// Fall back to executable directory + /config
	exe, err := os.Executable()
	if err != nil {
		return "./config"
	}
	return filepath.Join(filepath.Dir(exe), "config")
}
