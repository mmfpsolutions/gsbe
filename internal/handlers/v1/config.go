package v1

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mmfpsolutions/gsbe/internal/config"
	"github.com/mmfpsolutions/gsbe/internal/logger"
	"github.com/mmfpsolutions/gsbe/internal/services"
	v1types "github.com/mmfpsolutions/gsbe/internal/types/v1"
)

// HandleGetConfig returns the current configuration
func HandleGetConfig(cfgManager *config.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		cfg := cfgManager.GetConfig()
		v1types.RespondOK(w, cfg, v1types.NewMeta(start))
	}
}

// HandleUpdateConfig updates general config settings (port, title, logging)
func HandleUpdateConfig(cfgManager *config.Manager) http.HandlerFunc {
	log := logger.New(logger.ModuleHandler)

	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		var updates struct {
			Port    *int                `json:"port,omitempty"`
			Title   *string             `json:"title,omitempty"`
			Logging *config.LoggingConfig `json:"logging,omitempty"`
		}

		if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
			v1types.RespondErrorMsg(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
			return
		}

		cfg := cfgManager.GetConfig()

		if updates.Port != nil {
			cfg.Port = *updates.Port
		}
		if updates.Title != nil {
			cfg.Title = *updates.Title
		}
		if updates.Logging != nil {
			cfg.Logging = updates.Logging
			logger.SetGlobalLevel(updates.Logging.Level)
		}

		cfgManager.UpdateConfig(cfg)
		if err := cfgManager.SaveConfig(); err != nil {
			log.Error("Failed to save config: %v", err)
			v1types.RespondErrorMsg(w, http.StatusInternalServerError, "SAVE_ERROR", err.Error())
			return
		}

		v1types.RespondOK(w, cfg, v1types.NewMeta(start))
	}
}

// HandleCreateNode adds a new node to the configuration
func HandleCreateNode(cfgManager *config.Manager, nodeSvc *services.NodeService) http.HandlerFunc {
	log := logger.New(logger.ModuleHandler)

	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		var node config.NodeConnection
		if err := json.NewDecoder(r.Body).Decode(&node); err != nil {
			v1types.RespondErrorMsg(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
			return
		}

		// Generate ID
		node.ID = config.GenerateID()

		// Validate required fields
		if node.Name == "" || node.Host == "" || node.Port == 0 {
			v1types.RespondErrorMsg(w, http.StatusBadRequest, "VALIDATION_ERROR", "name, host, and port are required")
			return
		}

		cfg := cfgManager.GetConfig()
		cfg.Nodes = append(cfg.Nodes, node)
		cfgManager.UpdateConfig(cfg)

		if err := cfgManager.SaveConfig(); err != nil {
			log.Error("Failed to save config: %v", err)
			v1types.RespondErrorMsg(w, http.StatusInternalServerError, "SAVE_ERROR", err.Error())
			return
		}

		log.Info("Node created: %s (%s)", node.Name, node.ID)
		v1types.RespondOK(w, node, v1types.NewMeta(start))
	}
}

// HandleUpdateNode updates an existing node configuration
func HandleUpdateNode(cfgManager *config.Manager) http.HandlerFunc {
	log := logger.New(logger.ModuleHandler)

	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		nodeID := chi.URLParam(r, "id")

		var updates config.NodeConnection
		if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
			v1types.RespondErrorMsg(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
			return
		}

		cfg := cfgManager.GetConfig()
		found := false
		for i, node := range cfg.Nodes {
			if node.ID == nodeID {
				updates.ID = nodeID
				cfg.Nodes[i] = updates
				found = true
				break
			}
		}

		if !found {
			v1types.RespondErrorMsg(w, http.StatusNotFound, "NOT_FOUND", "node not found")
			return
		}

		cfgManager.UpdateConfig(cfg)
		if err := cfgManager.SaveConfig(); err != nil {
			log.Error("Failed to save config: %v", err)
			v1types.RespondErrorMsg(w, http.StatusInternalServerError, "SAVE_ERROR", err.Error())
			return
		}

		v1types.RespondOK(w, updates, v1types.NewMeta(start))
	}
}

// HandleDeleteNode removes a node from the configuration
func HandleDeleteNode(cfgManager *config.Manager) http.HandlerFunc {
	log := logger.New(logger.ModuleHandler)

	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		nodeID := chi.URLParam(r, "id")

		cfg := cfgManager.GetConfig()
		found := false
		newNodes := make([]config.NodeConnection, 0, len(cfg.Nodes))
		for _, node := range cfg.Nodes {
			if node.ID == nodeID {
				found = true
				continue
			}
			newNodes = append(newNodes, node)
		}

		if !found {
			v1types.RespondErrorMsg(w, http.StatusNotFound, "NOT_FOUND", "node not found")
			return
		}

		cfg.Nodes = newNodes
		cfgManager.UpdateConfig(cfg)

		if err := cfgManager.SaveConfig(); err != nil {
			log.Error("Failed to save config: %v", err)
			v1types.RespondErrorMsg(w, http.StatusInternalServerError, "SAVE_ERROR", err.Error())
			return
		}

		log.Info("Node deleted: %s", nodeID)
		v1types.RespondOK(w, map[string]string{"deleted": nodeID}, v1types.NewMeta(start))
	}
}

// HandleTestNode tests connectivity to a node
func HandleTestNode(nodeSvc *services.NodeService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		var node config.NodeConnection
		if err := json.NewDecoder(r.Body).Decode(&node); err != nil {
			v1types.RespondErrorMsg(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
			return
		}

		if node.Host == "" || node.Port == 0 {
			v1types.RespondErrorMsg(w, http.StatusBadRequest, "VALIDATION_ERROR", "host and port are required")
			return
		}

		err := nodeSvc.TestConnection(&node)
		if err != nil {
			v1types.RespondOK(w, map[string]interface{}{
				"success": false,
				"message": err.Error(),
			}, v1types.NewMeta(start))
			return
		}

		v1types.RespondOK(w, map[string]interface{}{
			"success": true,
			"message": "Connection successful",
		}, v1types.NewMeta(start))
	}
}
