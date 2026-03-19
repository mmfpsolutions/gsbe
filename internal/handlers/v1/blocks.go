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
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mmfpsolutions/gsbe/internal/config"
	"github.com/mmfpsolutions/gsbe/internal/services"
	v1types "github.com/mmfpsolutions/gsbe/internal/types/v1"
)

// HandleGetRecentBlocks returns the most recent N blocks
func HandleGetRecentBlocks(cfgManager *config.Manager, nodeSvc *services.NodeService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		nodeID := chi.URLParam(r, "node")

		count := 10
		if c := r.URL.Query().Get("count"); c != "" {
			if parsed, err := strconv.Atoi(c); err == nil && parsed > 0 && parsed <= 50 {
				count = parsed
			}
		}

		blocks, err := nodeSvc.GetRecentBlocks(nodeID, count)
		if err != nil {
			v1types.RespondErrorMsg(w, http.StatusBadGateway, "NODE_ERROR", err.Error())
			return
		}

		v1types.RespondOK(w, blocks, v1types.NewMeta(start))
	}
}

// HandleGetBlock returns a block by hash or height
func HandleGetBlock(cfgManager *config.Manager, nodeSvc *services.NodeService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		nodeID := chi.URLParam(r, "node")
		hashOrHeight := chi.URLParam(r, "hashOrHeight")

		var hash string

		// Check if it's a numeric height
		if height, err := strconv.ParseInt(hashOrHeight, 10, 64); err == nil {
			resolvedHash, err := nodeSvc.GetBlockHashByHeight(nodeID, height)
			if err != nil {
				v1types.RespondErrorMsg(w, http.StatusBadGateway, "NODE_ERROR", err.Error())
				return
			}
			hash = resolvedHash
		} else {
			hash = hashOrHeight
		}

		block, err := nodeSvc.GetBlock(nodeID, hash)
		if err != nil {
			v1types.RespondErrorMsg(w, http.StatusBadGateway, "NODE_ERROR", err.Error())
			return
		}

		v1types.RespondOK(w, block, v1types.NewMeta(start))
	}
}

// HandleGetTransaction fetches a transaction from a block
func HandleGetTransaction(cfgManager *config.Manager, nodeSvc *services.NodeService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		nodeID := chi.URLParam(r, "node")
		txid := chi.URLParam(r, "txid")
		blockhash := r.URL.Query().Get("blockhash")

		if blockhash == "" {
			v1types.RespondErrorMsg(w, http.StatusBadRequest, "MISSING_PARAM", "blockhash query parameter is required")
			return
		}

		block, err := nodeSvc.GetBlock(nodeID, blockhash)
		if err != nil {
			v1types.RespondErrorMsg(w, http.StatusBadGateway, "NODE_ERROR", err.Error())
			return
		}

		for _, tx := range block.Tx {
			if tx.TxID == txid {
				v1types.RespondOK(w, tx, v1types.NewMeta(start))
				return
			}
		}

		v1types.RespondErrorMsg(w, http.StatusNotFound, "TX_NOT_FOUND",
			"transaction not found in block")
	}
}

// HandleSearch determines if a query is a block height, block hash, or txid
func HandleSearch(cfgManager *config.Manager, nodeSvc *services.NodeService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		nodeID := chi.URLParam(r, "node")
		query := r.URL.Query().Get("q")

		if query == "" {
			v1types.RespondErrorMsg(w, http.StatusBadRequest, "MISSING_PARAM", "q query parameter is required")
			return
		}

		// Try as block height first
		if height, err := strconv.ParseInt(query, 10, 64); err == nil {
			hash, err := nodeSvc.GetBlockHashByHeight(nodeID, height)
			if err == nil {
				v1types.RespondOK(w, map[string]interface{}{
					"type":   "block",
					"hash":   hash,
					"height": height,
				}, v1types.NewMeta(start))
				return
			}
		}

		// Try as block hash (64 hex chars)
		if len(query) == 64 {
			block, err := nodeSvc.GetBlock(nodeID, query)
			if err == nil {
				v1types.RespondOK(w, map[string]interface{}{
					"type":   "block",
					"hash":   block.Hash,
					"height": block.Height,
				}, v1types.NewMeta(start))
				return
			}
		}

		v1types.RespondErrorMsg(w, http.StatusNotFound, "NOT_FOUND",
			"could not resolve query to a block or transaction")
	}
}
