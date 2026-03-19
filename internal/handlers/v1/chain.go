package v1

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mmfpsolutions/gsbe/internal/config"
	"github.com/mmfpsolutions/gsbe/internal/services"
	v1types "github.com/mmfpsolutions/gsbe/internal/types/v1"
)

// HandleGetNodes returns all configured nodes with their online status
func HandleGetNodes(cfgManager *config.Manager, nodeSvc *services.NodeService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		statuses := nodeSvc.GetNodeStatuses()
		v1types.RespondOK(w, statuses, v1types.NewMeta(start))
	}
}

// HandleGetChainInfo returns blockchain info for a specific node
func HandleGetChainInfo(cfgManager *config.Manager, nodeSvc *services.NodeService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		nodeID := chi.URLParam(r, "node")

		info, err := nodeSvc.GetChainInfo(nodeID)
		if err != nil {
			v1types.RespondErrorMsg(w, http.StatusBadGateway, "NODE_ERROR", err.Error())
			return
		}

		v1types.RespondOK(w, info, v1types.NewMeta(start))
	}
}
