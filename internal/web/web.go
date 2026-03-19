/*
 * Copyright 2026 Scott Walter, MMFP Solutions LLC
 *
 * This program is free software; you can redistribute it and/or modify it
 * under the terms of the GNU General Public License as published by the Free
 * Software Foundation; either version 3 of the License, or (at your option)
 * any later version.  See LICENSE for more details.
 */

package web

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mmfpsolutions/gsbe/internal/config"
	"github.com/mmfpsolutions/gsbe/internal/logger"
	"github.com/mmfpsolutions/gsbe/internal/version"
)

//go:embed templates/* static/*
var embeddedFiles embed.FS

// PageData holds data passed to HTML templates
type PageData struct {
	Title      string
	ActivePage string
	NodeID     string
	BlockHash  string
	TxID       string
	Query      string
	Version    string
	Year       int
}

var templates *template.Template

func init() {
	funcMap := template.FuncMap{
		"formatTime": func(unix int64) string {
			return time.Unix(unix, 0).Format("2006-01-02 15:04:05 MST")
		},
		"truncateHash": func(hash string, length int) string {
			if len(hash) <= length*2 {
				return hash
			}
			return hash[:length] + "..." + hash[len(hash)-length:]
		},
	}

	templates = template.Must(
		template.New("").Funcs(funcMap).ParseFS(embeddedFiles,
			"templates/layout/*.html",
			"templates/pages/*.html",
		),
	)
}

// RegisterRoutes registers web page routes and static file server
func RegisterRoutes(r chi.Router, cfgManager *config.Manager) {
	log := logger.New(logger.ModuleWeb)

	// Static file server
	staticFS, err := fs.Sub(embeddedFiles, "static")
	if err != nil {
		log.Fatal("Failed to create static file sub-FS: %v", err)
	}
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	// Setup redirect middleware
	r.Group(func(r chi.Router) {
		r.Use(setupRedirect(cfgManager))

		r.Get("/", servePage("dashboard"))
		r.Get("/dashboard", servePage("dashboard"))
		r.Get("/blocks", servePage("blocks"))
		r.Get("/block/{hash}", serveBlockPage())
		r.Get("/tx/{txid}", serveTxPage())
		r.Get("/mempool", servePage("mempool"))
	})

	// Config page is always accessible (no redirect)
	r.Get("/config", servePage("config"))
}

func servePage(page string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := PageData{
			Title:      "GSBE - GoSlimBlockExplorer",
			ActivePage: page,
			Version:    version.Version,
			Year:       time.Now().Year(),
		}
		renderPage(w, data)
	}
}

func serveBlockPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := PageData{
			Title:      "GSBE - Block Detail",
			ActivePage: "block-detail",
			BlockHash:  chi.URLParam(r, "hash"),
			Version:    version.Version,
			Year:       time.Now().Year(),
		}
		renderPage(w, data)
	}
}

func serveTxPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := PageData{
			Title:      "GSBE - Transaction Detail",
			ActivePage: "tx-detail",
			TxID:       chi.URLParam(r, "txid"),
			Query:      r.URL.Query().Get("blockhash"),
			Version:    version.Version,
			Year:       time.Now().Year(),
		}
		renderPage(w, data)
	}
}

func renderPage(w http.ResponseWriter, data PageData) {
	log := logger.New(logger.ModuleWeb)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := templates.ExecuteTemplate(w, "base", data); err != nil {
		log.Error("Template render error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// setupRedirect redirects to /config when no nodes are configured
func setupRedirect(cfgManager *config.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfgManager.SetupRequired() {
				http.Redirect(w, r, "/config", http.StatusTemporaryRedirect)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
