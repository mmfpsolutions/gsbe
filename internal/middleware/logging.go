/*
 * Copyright 2026 Scott Walter, MMFP Solutions LLC
 *
 * This program is free software; you can redistribute it and/or modify it
 * under the terms of the GNU General Public License as published by the Free
 * Software Foundation; either version 3 of the License, or (at your option)
 * any later version.  See LICENSE for more details.
 */

package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/mmfpsolutions/gsbe/internal/logger"
)

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// LoggingMiddleware logs HTTP request method, path, status, and duration
func LoggingMiddleware(next http.Handler) http.Handler {
	log := logger.New(logger.ModuleMiddleware)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip logging for health and static paths
		if r.URL.Path == "/health" || strings.HasPrefix(r.URL.Path, "/static/") {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()
		rw := newResponseWriter(w)
		next.ServeHTTP(rw, r)
		duration := time.Since(start)

		log.Info("%s %s %d %s", r.Method, r.URL.Path, rw.statusCode, duration)
	})
}
