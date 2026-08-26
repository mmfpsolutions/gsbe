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
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"
)

// decodeBody parses a recorded response as a generic map so the tests assert
// on the WIRE shape — which keys are present — rather than on a re-decode into
// APIResponse, which would hide every omitempty mistake.
func decodeBody(t *testing.T, rec *httptest.ResponseRecorder) map[string]json.RawMessage {
	t.Helper()
	var body map[string]json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response body is not valid JSON: %v\nbody: %s", err, rec.Body.String())
	}
	return body
}

// ---------------------------------------------------------------------------
// NewMeta
// ---------------------------------------------------------------------------

func TestNewMetaFormat(t *testing.T) {
	meta := NewMeta(time.Now().Add(-1500 * time.Millisecond))

	// The web UI parses this string, so the "<int>ms" shape is contractual.
	if !regexp.MustCompile(`^\d+ms$`).MatchString(meta.RequestDuration) {
		t.Errorf("RequestDuration = %q, want the form \"<int>ms\"", meta.RequestDuration)
	}

	// Truncation, not rounding: 1500ms is 1500, and a sub-millisecond request
	// reports "0ms" rather than being omitted.
	if meta.RequestDuration != "1500ms" && meta.RequestDuration != "1501ms" {
		t.Errorf("RequestDuration = %q, want ~1500ms", meta.RequestDuration)
	}

	now := time.Now().Unix()
	if meta.Timestamp < now-5 || meta.Timestamp > now+5 {
		t.Errorf("Timestamp = %d, want within 5s of %d", meta.Timestamp, now)
	}
}

func TestNewMetaSubMillisecondReportsZero(t *testing.T) {
	meta := NewMeta(time.Now())
	if meta.RequestDuration != "0ms" {
		t.Errorf("RequestDuration = %q, want 0ms for an instant request", meta.RequestDuration)
	}
}

// ---------------------------------------------------------------------------
// RespondOK
// ---------------------------------------------------------------------------

func TestRespondOK(t *testing.T) {
	rec := httptest.NewRecorder()
	RespondOK(rec, map[string]string{"chain": "main"}, NewMeta(time.Now()))

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	body := decodeBody(t, rec)
	if string(body["status"]) != `"ok"` {
		t.Errorf("status field = %s, want \"ok\"", body["status"])
	}
	if _, present := body["errors"]; present {
		t.Error("errors key present on a success response, want it omitted")
	}
	if _, present := body["meta"]; !present {
		t.Error("meta key missing, want it present when a Meta is supplied")
	}
}

func TestRespondOKOmitsMetaWhenNil(t *testing.T) {
	rec := httptest.NewRecorder()
	RespondOK(rec, map[string]string{"ok": "yes"}, nil)

	if _, present := decodeBody(t, rec)["meta"]; present {
		t.Error("meta key present for a nil Meta, want it omitted")
	}
}

// The `data,omitempty` tag is the sharp edge in this envelope, because
// omitempty on an interface{} field tests the INTERFACE for nil, not the value
// inside it. These three cases are the ones handlers actually hit.
func TestRespondOKDataOmitemptySemantics(t *testing.T) {
	tests := []struct {
		name string
		data interface{}
		// wantRaw is the exact serialisation of the data key; "" means the key
		// must be absent entirely.
		wantRaw string
	}{
		{
			name:    "untyped nil is omitted",
			data:    nil,
			wantRaw: "",
		},
		{
			// The trap. A handler that declares `var blocks []Block`, never
			// appends, and passes it here does NOT get `"data":[]` — it gets
			// `"data":null`, because the interface holding a typed nil slice
			// is itself non-nil. Front-end code doing `data.map(...)` breaks
			// on null and survives []. Handlers must initialise empty slices.
			name:    "typed nil slice serialises as null, NOT omitted",
			data:    []Block(nil),
			wantRaw: "null",
		},
		{
			name:    "empty non-nil slice serialises as []",
			data:    []Block{},
			wantRaw: "[]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			RespondOK(rec, tt.data, nil)

			raw, present := decodeBody(t, rec)["data"]
			if tt.wantRaw == "" {
				if present {
					t.Errorf("data key = %s, want it absent", raw)
				}
				return
			}
			if !present {
				t.Fatalf("data key absent, want %s", tt.wantRaw)
			}
			if got := strings.TrimSpace(string(raw)); got != tt.wantRaw {
				t.Errorf("data = %s, want %s", got, tt.wantRaw)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// RespondError / RespondErrorMsg
// ---------------------------------------------------------------------------

func TestRespondErrorHonoursStatusCode(t *testing.T) {
	for _, code := range []int{
		http.StatusBadRequest,
		http.StatusNotFound,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
	} {
		rec := httptest.NewRecorder()
		RespondError(rec, code, []APIError{{Code: "TEST", Message: "test"}})

		if rec.Code != code {
			t.Errorf("status = %d, want %d", rec.Code, code)
		}
		body := decodeBody(t, rec)
		if string(body["status"]) != `"error"` {
			t.Errorf("status field = %s, want \"error\"", body["status"])
		}
		// An error response carries no data and no meta — the UI branches on
		// the presence of `errors`.
		if _, present := body["data"]; present {
			t.Error("data key present on an error response, want it omitted")
		}
		if _, present := body["meta"]; present {
			t.Error("meta key present on an error response, want it omitted")
		}
	}
}

func TestRespondErrorCarriesEveryError(t *testing.T) {
	rec := httptest.NewRecorder()
	RespondError(rec, http.StatusBadRequest, []APIError{
		{Code: "BAD_HEIGHT", Message: "height must be numeric", Target: "height"},
		{Code: "BAD_NODE", Message: "unknown node id", Target: "node"},
	})

	var resp APIResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if len(resp.Errors) != 2 {
		t.Fatalf("len(Errors) = %d, want 2 — validation reports every failure at once", len(resp.Errors))
	}
	if resp.Errors[1].Target != "node" {
		t.Errorf("Errors[1].Target = %q, want node", resp.Errors[1].Target)
	}
}

func TestRespondErrorMsgOmitsEmptyTarget(t *testing.T) {
	rec := httptest.NewRecorder()
	RespondErrorMsg(rec, http.StatusNotFound, "NOT_FOUND", "block not found")

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}

	var resp struct {
		Status string              `json:"status"`
		Errors []map[string]string `json:"errors"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if len(resp.Errors) != 1 {
		t.Fatalf("len(Errors) = %d, want 1", len(resp.Errors))
	}
	if _, present := resp.Errors[0]["target"]; present {
		t.Error("target key present when unset, want it omitted")
	}
	if resp.Errors[0]["code"] != "NOT_FOUND" || resp.Errors[0]["message"] != "block not found" {
		t.Errorf("error = %+v, want the supplied code and message", resp.Errors[0])
	}
}

// Every response ends in a newline because writeJSON uses json.Encoder. Pinned
// because switching to json.Marshal would silently drop it, and anything
// diffing raw response bodies would light up.
func TestResponsesEndWithNewline(t *testing.T) {
	rec := httptest.NewRecorder()
	RespondOK(rec, map[string]string{"a": "b"}, nil)

	if !strings.HasSuffix(rec.Body.String(), "\n") {
		t.Errorf("body = %q, want a trailing newline from json.Encoder", rec.Body.String())
	}
}

// KNOWN GAP, documented rather than fixed: writeJSON calls WriteHeader BEFORE
// encoding, so if encoding fails its http.Error fallback cannot change the
// status — the header is already on the wire and the client receives a 200
// with a truncated body. This test pins that reality so the fallback is not
// mistaken for working error handling. Fixing it means encoding to a buffer
// first, which is a real change in allocation behaviour on large block lists.
func TestWriteJSONCannotRecoverFromEncodeFailure(t *testing.T) {
	rec := httptest.NewRecorder()

	// A channel cannot be marshalled, so the encoder fails mid-response.
	RespondOK(rec, make(chan int), nil)

	// 200, not the 500 the http.Error fallback asks for.
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200 — the header is already sent when encoding fails", rec.Code)
	}
	// And the body is http.Error's plain text, not JSON, so a client that
	// trusts the 200 and parses the body fails at the parse rather than on a
	// status check.
	if got := rec.Body.String(); got != "failed to encode response\n" {
		t.Errorf("body = %q, want the plain-text fallback", got)
	}
	if json.Valid(rec.Body.Bytes()) {
		t.Error("body parsed as JSON; this gap is that it does NOT")
	}
}
