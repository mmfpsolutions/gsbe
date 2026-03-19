/*
 * Copyright 2026 Scott Walter, MMFP Solutions LLC
 *
 * This program is free software; you can redistribute it and/or modify it
 * under the terms of the GNU General Public License as published by the Free
 * Software Foundation; either version 3 of the License, or (at your option)
 * any later version.  See LICENSE for more details.
 */

package v1

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mmfpsolutions/gsbe/internal/config"
	"github.com/mmfpsolutions/gsbe/internal/services"
	v1types "github.com/mmfpsolutions/gsbe/internal/types/v1"
)

// HandleGetMempoolInfo returns mempool info for a node
func HandleGetMempoolInfo(cfgManager *config.Manager, nodeSvc *services.NodeService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		nodeID := chi.URLParam(r, "node")

		info, err := nodeSvc.GetMempoolInfo(nodeID)
		if err != nil {
			v1types.RespondErrorMsg(w, http.StatusBadGateway, "NODE_ERROR", err.Error())
			return
		}

		v1types.RespondOK(w, info, v1types.NewMeta(start))
	}
}
