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
	"github.com/mmfpsolutions/gsbe/internal/config"
	v1 "github.com/mmfpsolutions/gsbe/internal/handlers/v1"
	"github.com/mmfpsolutions/gsbe/internal/services"
)

func registerV1Routes(r chi.Router, cfgManager *config.Manager, nodeSvc *services.NodeService) {
	r.Route("/api/v1", func(r chi.Router) {
		// Health and status
		r.Get("/health", v1.HandleHealth())
		r.Get("/status", v1.HandleStatus(cfgManager))

		// Nodes list
		r.Get("/nodes", v1.HandleGetNodes(cfgManager, nodeSvc))

		// Config management
		r.Route("/config", func(r chi.Router) {
			r.Get("/", v1.HandleGetConfig(cfgManager))
			r.Patch("/", v1.HandleUpdateConfig(cfgManager))

			// Node management
			r.Post("/nodes", v1.HandleCreateNode(cfgManager, nodeSvc))
			r.Post("/nodes/test", v1.HandleTestNode(nodeSvc))
			r.Put("/nodes/{id}", v1.HandleUpdateNode(cfgManager))
			r.Delete("/nodes/{id}", v1.HandleDeleteNode(cfgManager))
		})

		// Node-specific blockchain data
		r.Route("/{node}", func(r chi.Router) {
			r.Get("/chain", v1.HandleGetChainInfo(cfgManager, nodeSvc))
			r.Get("/blocks/recent", v1.HandleGetRecentBlocks(cfgManager, nodeSvc))
			r.Get("/block/{hashOrHeight}", v1.HandleGetBlock(cfgManager, nodeSvc))
			r.Get("/tx/{txid}", v1.HandleGetTransaction(cfgManager, nodeSvc))
			r.Get("/mempool", v1.HandleGetMempoolInfo(cfgManager, nodeSvc))
			r.Get("/search", v1.HandleSearch(cfgManager, nodeSvc))
		})
	})

	// Convenience health endpoint at root
	r.Get("/health", v1.HandleHealth())
}
