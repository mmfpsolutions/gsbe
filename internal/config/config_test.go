/*
 * Copyright 2026 Scott Walter, MMFP Solutions LLC
 *
 * This program is free software; you can redistribute it and/or modify it
 * under the terms of the GNU General Public License as published by the Free
 * Software Foundation; either version 3 of the License, or (at your option)
 * any later version.  See LICENSE for more details.
 */

package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/mmfpsolutions/gsbe/internal/logger"
)

// The Manager logs on every load and save. Tests exercise those paths dozens
// of times, so quiet the logger rather than drowning `go test -v` output.
func TestMain(m *testing.M) {
	logger.SetGlobalLevel("FATAL")
	os.Exit(m.Run())
}

// writeConfig drops raw JSON at <dir>/config.json. Tests pass raw strings
// rather than marshalled structs so that malformed and partial documents —
// the cases that actually reach an operator — can be expressed directly.
func writeConfig(t *testing.T, dir, contents string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(contents), 0644); err != nil {
		t.Fatalf("writing fixture config: %v", err)
	}
}

// ---------------------------------------------------------------------------
// LoadConfig
// ---------------------------------------------------------------------------

// A missing config file is NOT an error: the explorer starts on defaults and
// the setup flow takes over. Pinning this because the obvious "fix" — treating
// a missing file as fatal — would break first-run.
func TestLoadConfigMissingFileFallsBackToDefaults(t *testing.T) {
	m := GetManager(t.TempDir())

	if err := m.LoadConfig(); err != nil {
		t.Fatalf("LoadConfig() on missing file = %v, want nil", err)
	}

	cfg := m.GetConfig()
	if cfg.Port != 3007 {
		t.Errorf("Port = %d, want 3007", cfg.Port)
	}
	if cfg.Title != "GSBE - GoSlimBlockExplorer" {
		t.Errorf("Title = %q, want the default title", cfg.Title)
	}
	if cfg.Nodes == nil {
		t.Error("Nodes = nil, want an empty non-nil slice")
	}
	if len(cfg.Nodes) != 0 {
		t.Errorf("len(Nodes) = %d, want 0", len(cfg.Nodes))
	}
	if cfg.Logging == nil {
		t.Fatal("Logging = nil, want defaults")
	}
	if cfg.Logging.Level != "INFO" {
		t.Errorf("Logging.Level = %q, want INFO", cfg.Logging.Level)
	}
}

// Malformed JSON IS an error — it means the operator edited the file and got
// it wrong, and silently reverting to defaults would hide that.
func TestLoadConfigMalformedJSONIsAnError(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, `{"port": 3007,`)

	m := GetManager(dir)
	err := m.LoadConfig()
	if err == nil {
		t.Fatal("LoadConfig() on malformed JSON = nil, want an error")
	}
	if !strings.Contains(err.Error(), "failed to parse config") {
		t.Errorf("error = %q, want it to name the parse failure", err)
	}
}

// Field-level defaults, applied after a successful parse. Each case supplies a
// partial document so a failure names exactly one missing default.
func TestLoadConfigAppliesFieldDefaults(t *testing.T) {
	tests := []struct {
		name  string
		json  string
		check func(*testing.T, *Config)
	}{
		{
			name: "absent port defaults to 3007",
			json: `{"title":"Custom"}`,
			check: func(t *testing.T, c *Config) {
				if c.Port != 3007 {
					t.Errorf("Port = %d, want 3007", c.Port)
				}
			},
		},
		{
			// KNOWN behaviour, not a bug worth fixing: because the check is
			// `== 0` rather than a presence check, an operator cannot ask for
			// port 0. Nobody wants port 0, but a future move to a pointer
			// field should be deliberate.
			name: "explicit port 0 is indistinguishable from absent",
			json: `{"port":0}`,
			check: func(t *testing.T, c *Config) {
				if c.Port != 3007 {
					t.Errorf("Port = %d, want 3007", c.Port)
				}
			},
		},
		{
			name: "explicit port is preserved",
			json: `{"port":8080}`,
			check: func(t *testing.T, c *Config) {
				if c.Port != 8080 {
					t.Errorf("Port = %d, want 8080", c.Port)
				}
			},
		},
		{
			name: "empty title defaults",
			json: `{"port":3007,"title":""}`,
			check: func(t *testing.T, c *Config) {
				if c.Title != "GSBE - GoSlimBlockExplorer" {
					t.Errorf("Title = %q, want the default title", c.Title)
				}
			},
		},
		{
			name: "absent logging block defaults wholesale",
			json: `{"port":3007}`,
			check: func(t *testing.T, c *Config) {
				if c.Logging == nil {
					t.Fatal("Logging = nil, want defaults")
				}
				if c.Logging.Level != "INFO" || c.Logging.LogToFile || c.Logging.LogFilePath != "logs/gsbe.log" {
					t.Errorf("Logging = %+v, want the default block", *c.Logging)
				}
			},
		},
		{
			// KNOWN GAP: defaults are applied to the logging block only when
			// it is absent ENTIRELY. A present-but-partial block keeps its
			// zero values, so `{"logging":{"log_to_file":true}}` yields an
			// empty Level and an empty LogFilePath. logger.SetGlobalLevel
			// treats the empty string as INFO, so the level is harmless — the
			// empty file path is the one that bites, and it is why
			// SetupFileLogging checks for it.
			name: "partial logging block does NOT get per-field defaults",
			json: `{"port":3007,"logging":{"log_to_file":true}}`,
			check: func(t *testing.T, c *Config) {
				if c.Logging == nil {
					t.Fatal("Logging = nil, want the supplied block")
				}
				if c.Logging.Level != "" {
					t.Errorf("Logging.Level = %q, want \"\" (documents the gap)", c.Logging.Level)
				}
				if c.Logging.LogFilePath != "" {
					t.Errorf("Logging.LogFilePath = %q, want \"\" (documents the gap)", c.Logging.LogFilePath)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			writeConfig(t, dir, tt.json)

			m := GetManager(dir)
			if err := m.LoadConfig(); err != nil {
				t.Fatalf("LoadConfig() = %v, want nil", err)
			}
			tt.check(t, m.GetConfig())
		})
	}
}

func TestLoadConfigPreservesNodes(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, `{
		"port": 3007,
		"nodes": [
			{"id":"abc12345","name":"DigiByte","symbol":"DGB","host":"192.168.1.10","port":14022,"network":"mainnet","rest_enabled":true},
			{"id":"def67890","name":"Bitcoin","symbol":"BTC","host":"192.168.1.11","port":8332,"network":"mainnet","rest_enabled":false}
		]
	}`)

	m := GetManager(dir)
	if err := m.LoadConfig(); err != nil {
		t.Fatalf("LoadConfig() = %v, want nil", err)
	}

	cfg := m.GetConfig()
	if len(cfg.Nodes) != 2 {
		t.Fatalf("len(Nodes) = %d, want 2", len(cfg.Nodes))
	}
	if cfg.Nodes[0].Symbol != "DGB" || cfg.Nodes[0].Port != 14022 || !cfg.Nodes[0].RESTEnabled {
		t.Errorf("Nodes[0] = %+v, want the DGB node decoded intact", cfg.Nodes[0])
	}
	if cfg.Nodes[1].RESTEnabled {
		t.Error("Nodes[1].RESTEnabled = true, want false — rest_enabled must decode per node")
	}
}

// ---------------------------------------------------------------------------
// Save / round trip
// ---------------------------------------------------------------------------

func TestSaveConfigRoundTrips(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "config")

	m := GetManager(dir)
	want := &Config{
		Port:  9000,
		Title: "Round Trip",
		Nodes: []NodeConnection{
			{ID: "node0001", Name: "DigiByte", Symbol: "DGB", Host: "10.0.0.5", Port: 14022, Network: "mainnet", RESTEnabled: true},
		},
		Logging: &LoggingConfig{Level: "DEBUG", LogToFile: true, LogFilePath: "logs/x.log"},
	}
	m.UpdateConfig(want)

	// SaveConfig must create the directory tree — the config dir is a mounted
	// volume that may be empty on first run.
	if err := m.SaveConfig(); err != nil {
		t.Fatalf("SaveConfig() = %v, want nil", err)
	}

	reloaded := GetManager(dir)
	if err := reloaded.LoadConfig(); err != nil {
		t.Fatalf("LoadConfig() after save = %v, want nil", err)
	}

	got := reloaded.GetConfig()
	if got.Port != want.Port || got.Title != want.Title {
		t.Errorf("round trip: Port/Title = %d/%q, want %d/%q", got.Port, got.Title, want.Port, want.Title)
	}
	if len(got.Nodes) != 1 || got.Nodes[0] != want.Nodes[0] {
		t.Errorf("round trip: Nodes = %+v, want %+v", got.Nodes, want.Nodes)
	}
	if got.Logging == nil || *got.Logging != *want.Logging {
		t.Errorf("round trip: Logging = %+v, want %+v", got.Logging, want.Logging)
	}
}

// The config file can hold node hosts and ports for an operator's private
// infrastructure, so it must not be group- or world-writable.
func TestSaveConfigFileIsNotWorldWritable(t *testing.T) {
	dir := t.TempDir()
	m := GetManager(dir)
	m.UpdateConfig(m.defaultConfig())

	if err := m.SaveConfig(); err != nil {
		t.Fatalf("SaveConfig() = %v, want nil", err)
	}

	info, err := os.Stat(filepath.Join(dir, "config.json"))
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	// Asserted as a mask rather than an exact 0644 so a restrictive CI umask
	// (which would yield 0600) does not fail the test for the wrong reason.
	if mode := info.Mode().Perm(); mode&0o022 != 0 {
		t.Errorf("config.json mode = %#o, want no group/world write bits", mode)
	}
}

func TestSaveConfigWithoutConfigIsAnError(t *testing.T) {
	m := GetManager(t.TempDir())

	// Deliberately no LoadConfig / UpdateConfig: m.config is nil.
	err := m.SaveConfig()
	if err == nil {
		t.Fatal("SaveConfig() with no config = nil, want an error")
	}
	if !strings.Contains(err.Error(), "no config to save") {
		t.Errorf("error = %q, want it to name the empty config", err)
	}
}

func TestWriteDefaultConfigProducesLoadableFile(t *testing.T) {
	dir := t.TempDir()

	m := GetManager(dir)
	if err := m.WriteDefaultConfig(); err != nil {
		t.Fatalf("WriteDefaultConfig() = %v, want nil", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "config.json"))
	if err != nil {
		t.Fatalf("reading written config: %v", err)
	}

	// Written with MarshalIndent — an operator edits this by hand, so it must
	// not be a single minified line.
	if !strings.Contains(string(data), "\n  ") {
		t.Error("written config is not indented; operators edit this file by hand")
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("written config does not parse: %v", err)
	}
	if cfg.Port != 3007 {
		t.Errorf("written Port = %d, want 3007", cfg.Port)
	}
	if cfg.Nodes == nil {
		t.Error("written Nodes = null, want []; the setup UI appends to this")
	}
}

// ---------------------------------------------------------------------------
// SetupRequired
// ---------------------------------------------------------------------------

func TestSetupRequired(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*Manager)
		want  bool
	}{
		{
			name:  "nil config (LoadConfig never called)",
			setup: func(*Manager) {},
			want:  true,
		},
		{
			name:  "loaded defaults, no nodes",
			setup: func(m *Manager) { m.UpdateConfig(m.defaultConfig()) },
			want:  true,
		},
		{
			// A nodes key present but empty is the same as no nodes: the
			// operator deleted their last node and must be sent back to setup.
			name: "explicitly empty node list",
			setup: func(m *Manager) {
				m.UpdateConfig(&Config{Port: 3007, Nodes: []NodeConnection{}})
			},
			want: true,
		},
		{
			name: "one node configured",
			setup: func(m *Manager) {
				m.UpdateConfig(&Config{Port: 3007, Nodes: []NodeConnection{{ID: "abc12345"}}})
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := GetManager(t.TempDir())
			tt.setup(m)
			if got := m.SetupRequired(); got != tt.want {
				t.Errorf("SetupRequired() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// GetNodeByID
// ---------------------------------------------------------------------------

func TestGetNodeByID(t *testing.T) {
	m := GetManager(t.TempDir())
	m.UpdateConfig(&Config{
		Port: 3007,
		Nodes: []NodeConnection{
			{ID: "abc12345", Symbol: "DGB"},
			{ID: "def67890", Symbol: "BTC"},
		},
	})

	tests := []struct {
		name       string
		id         string
		wantSymbol string // "" means want nil
	}{
		{"first node", "abc12345", "DGB"},
		{"second node", "def67890", "BTC"},
		{"unknown id", "nope0000", ""},
		{"empty id", "", ""},
		// IDs come straight off the URL path, so case and whitespace variants
		// must NOT resolve — the lookup is an exact match by design.
		{"wrong case", "ABC12345", ""},
		{"trailing space", "abc12345 ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.GetNodeByID(tt.id)
			if tt.wantSymbol == "" {
				if got != nil {
					t.Errorf("GetNodeByID(%q) = %+v, want nil", tt.id, *got)
				}
				return
			}
			if got == nil {
				t.Fatalf("GetNodeByID(%q) = nil, want the %s node", tt.id, tt.wantSymbol)
			}
			if got.Symbol != tt.wantSymbol {
				t.Errorf("GetNodeByID(%q).Symbol = %q, want %q", tt.id, got.Symbol, tt.wantSymbol)
			}
		})
	}
}

// GetNodeByID takes the address of the range variable. Under Go 1.22+ loop
// semantics that is a fresh copy per iteration, so callers cannot reach into
// the stored config through it. This test fails loudly if the loop is ever
// rewritten to hand out `&m.config.Nodes[i]`.
func TestGetNodeByIDReturnsACopy(t *testing.T) {
	m := GetManager(t.TempDir())
	m.UpdateConfig(&Config{
		Port:  3007,
		Nodes: []NodeConnection{{ID: "abc12345", Host: "10.0.0.5"}},
	})

	node := m.GetNodeByID("abc12345")
	if node == nil {
		t.Fatal("GetNodeByID = nil, want the node")
	}
	node.Host = "evil.example.com"

	if again := m.GetNodeByID("abc12345"); again.Host != "10.0.0.5" {
		t.Errorf("stored Host = %q after mutating the returned node, want 10.0.0.5", again.Host)
	}
}

// ---------------------------------------------------------------------------
// GetConfig
// ---------------------------------------------------------------------------

// GetConfig on an unloaded Manager synthesises defaults but does NOT store
// them, so each call returns a distinct object. Pinning this because a caller
// that mutates the result and expects it to stick would silently lose writes.
func TestGetConfigOnUnloadedManagerDoesNotCache(t *testing.T) {
	m := GetManager(t.TempDir())

	first := m.GetConfig()
	first.Port = 9999

	if second := m.GetConfig(); second.Port != 3007 {
		t.Errorf("second GetConfig().Port = %d, want 3007 — defaults must not be cached", second.Port)
	}
}

// ---------------------------------------------------------------------------
// GenerateID
// ---------------------------------------------------------------------------

func TestGenerateID(t *testing.T) {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

	// IDs land in URL paths and in config.json keys, so the alphabet matters
	// as much as the length.
	seen := make(map[string]bool, 500)
	for i := 0; i < 500; i++ {
		id := GenerateID()

		if len(id) != 8 {
			t.Fatalf("GenerateID() = %q, want 8 characters", id)
		}
		for _, r := range id {
			if !strings.ContainsRune(charset, r) {
				t.Fatalf("GenerateID() = %q, contains %q outside the lowercase alphanumeric charset", id, r)
			}
		}
		seen[id] = true
	}

	// 36^8 keyspace over 500 draws: a collision is possible but a heavy
	// clustering of them means the generator has stopped being random.
	if len(seen) < 495 {
		t.Errorf("500 generated IDs yielded only %d distinct values; generator looks degenerate", len(seen))
	}
}

// ---------------------------------------------------------------------------
// Concurrency
// ---------------------------------------------------------------------------

// The Manager is shared by every HTTP handler, so its RWMutex has to hold up.
// Meaningful under `go test -race`, which CI runs.
func TestManagerConcurrentAccess(t *testing.T) {
	m := GetManager(t.TempDir())
	m.UpdateConfig(m.defaultConfig())

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(3)
		go func() { defer wg.Done(); _ = m.GetConfig() }()
		go func() { defer wg.Done(); _ = m.SetupRequired() }()
		go func() {
			defer wg.Done()
			m.UpdateConfig(&Config{Port: 3007, Nodes: []NodeConnection{{ID: GenerateID()}}})
		}()
	}
	wg.Wait()
}
