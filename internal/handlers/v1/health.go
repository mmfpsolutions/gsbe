package v1

import (
	"net/http"
	"runtime"
	"time"

	"github.com/mmfpsolutions/gsbe/internal/config"
	v1types "github.com/mmfpsolutions/gsbe/internal/types/v1"
	"github.com/mmfpsolutions/gsbe/internal/version"
)

// HandleHealth returns a simple health check response
func HandleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		v1types.RespondOK(w, map[string]interface{}{
			"status":  "healthy",
			"service": "gsbe",
			"version": version.Version,
		}, v1types.NewMeta(start))
	}
}

// HandleStatus returns version, uptime, and runtime stats
func HandleStatus(cfgManager *config.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		var mem runtime.MemStats
		runtime.ReadMemStats(&mem)

		v1types.RespondOK(w, map[string]interface{}{
			"version":    version.Version,
			"build_date": version.BuildDate,
			"commit":     version.Commit,
			"uptime":     version.Uptime().String(),
			"goroutines": runtime.NumGoroutine(),
			"memory": map[string]interface{}{
				"alloc_mb":       mem.Alloc / 1024 / 1024,
				"total_alloc_mb": mem.TotalAlloc / 1024 / 1024,
				"sys_mb":         mem.Sys / 1024 / 1024,
				"num_gc":         mem.NumGC,
			},
		}, v1types.NewMeta(start))
	}
}
