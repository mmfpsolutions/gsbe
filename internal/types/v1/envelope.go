/*
 * Copyright 2026 Scott Walter, MMFP Solutions LLC
 *
 * This program is free software; you can redistribute it and/or modify it
 * under the terms of the GNU General Public License as published by the Free
 * Software Foundation; either version 3 of the License, or (at your option)
 * any later version.  See LICENSE for more details.
 */

package v1types

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// APIResponse is the standard JSON envelope for all API responses
type APIResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data,omitempty"`
	Errors []APIError  `json:"errors,omitempty"`
	Meta   *Meta       `json:"meta,omitempty"`
}

// APIError represents a single error in the response
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Target  string `json:"target,omitempty"`
}

// Meta contains request metadata
type Meta struct {
	RequestDuration string `json:"request_duration"`
	Timestamp       int64  `json:"timestamp"`
}

// NewMeta creates a Meta with duration calculated from start time
func NewMeta(start time.Time) *Meta {
	return &Meta{
		RequestDuration: fmt.Sprintf("%dms", time.Since(start).Milliseconds()),
		Timestamp:       time.Now().Unix(),
	}
}

// RespondOK writes a 200 JSON response with status "ok"
func RespondOK(w http.ResponseWriter, data interface{}, meta *Meta) {
	writeJSON(w, http.StatusOK, APIResponse{
		Status: "ok",
		Data:   data,
		Meta:   meta,
	})
}

// RespondError writes an error JSON response
func RespondError(w http.ResponseWriter, statusCode int, errors []APIError) {
	writeJSON(w, statusCode, APIResponse{
		Status: "error",
		Errors: errors,
	})
}

// RespondErrorMsg is a shorthand for a single error response
func RespondErrorMsg(w http.ResponseWriter, statusCode int, code, message string) {
	RespondError(w, statusCode, []APIError{
		{Code: code, Message: message},
	})
}

func writeJSON(w http.ResponseWriter, statusCode int, response APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
