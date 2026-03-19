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
