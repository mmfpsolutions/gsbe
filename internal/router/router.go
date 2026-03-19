/*
 * Copyright 2026 Scott Walter, MMFP Solutions LLC
 *
 * This program is free software; you can redistribute it and/or modify it
 * under the terms of the GNU General Public License as published by the Free
 * Software Foundation; either version 3 of the License, or (at your option)
 * any later version.  See LICENSE for more details.
 */

package router

import (
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/mmfpsolutions/gsbe/internal/config"
	"github.com/mmfpsolutions/gsbe/internal/middleware"
	"github.com/mmfpsolutions/gsbe/internal/services"
	"github.com/mmfpsolutions/gsbe/internal/web"
)

// SetupRouter creates and configures the application router
func SetupRouter(cfgManager *config.Manager, configDir string) chi.Router {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimw.RealIP)
	r.Use(chimw.Recoverer)
	r.Use(middleware.LoggingMiddleware)

	// Create services
	nodeSvc := services.NewNodeService(cfgManager)

	// Register API routes
	registerV1Routes(r, cfgManager, nodeSvc)

	// Register web routes (pages + static)
	web.RegisterRoutes(r, cfgManager)

	return r
}
