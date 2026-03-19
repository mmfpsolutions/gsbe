package config

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"sync"

	"github.com/mmfpsolutions/gsbe/internal/logger"
)

// NodeConnection represents a blockchain node configuration
type NodeConnection struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Symbol      string `json:"symbol"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Network     string `json:"network"`
	RESTEnabled bool   `json:"rest_enabled"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level       string `json:"level"`
	LogToFile   bool   `json:"log_to_file"`
	LogFilePath string `json:"log_file_path"`
}

// Config is the top-level application configuration
type Config struct {
	Port    int              `json:"port"`
	Title   string           `json:"title"`
	Nodes   []NodeConnection `json:"nodes"`
	Logging *LoggingConfig   `json:"logging"`
}

// Manager handles reading and writing configuration
type Manager struct {
	mu        sync.RWMutex
	configDir string
	config    *Config
	log       *logger.Logger
}

// GetManager creates a new config Manager for the given directory
func GetManager(configDir string) *Manager {
	return &Manager{
		configDir: configDir,
		log:       logger.New(logger.ModuleConfig),
	}
}

func (m *Manager) configPath() string {
	return filepath.Join(m.configDir, "config.json")
}

// LoadConfig reads config.json from the config directory
func (m *Manager) LoadConfig() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.configPath())
	if err != nil {
		if os.IsNotExist(err) {
			m.log.Warn("Config file not found at %s", m.configPath())
			m.config = m.defaultConfig()
			return nil
		}
		return fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	// Apply defaults for missing fields
	if cfg.Port == 0 {
		cfg.Port = 3007
	}
	if cfg.Title == "" {
		cfg.Title = "GSBE - GoSlimBlockExplorer"
	}
	if cfg.Logging == nil {
		cfg.Logging = &LoggingConfig{
			Level:       "INFO",
			LogToFile:   false,
			LogFilePath: "logs/gsbe.log",
		}
	}

	m.config = &cfg
	m.log.Info("Config loaded from %s", m.configPath())
	return nil
}

// GetConfig returns the current config (read-locked)
func (m *Manager) GetConfig() *Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.config == nil {
		return m.defaultConfig()
	}
	return m.config
}

// SaveConfig writes the current config to disk
func (m *Manager) SaveConfig() error {
	m.mu.RLock()
	cfg := m.config
	m.mu.RUnlock()

	if cfg == nil {
		return fmt.Errorf("no config to save")
	}

	if err := os.MkdirAll(m.configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(m.configPath(), data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	m.log.Info("Config saved to %s", m.configPath())
	return nil
}

// WriteDefaultConfig creates a default config.json
func (m *Manager) WriteDefaultConfig() error {
	m.mu.Lock()
	m.config = m.defaultConfig()
	m.mu.Unlock()
	return m.SaveConfig()
}

// SetupRequired returns true when no nodes are configured
func (m *Manager) SetupRequired() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.config == nil {
		return true
	}
	return len(m.config.Nodes) == 0
}

// GetNodeByID finds a node by its ID
func (m *Manager) GetNodeByID(id string) *NodeConnection {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.config == nil {
		return nil
	}
	for _, node := range m.config.Nodes {
		if node.ID == id {
			return &node
		}
	}
	return nil
}

// UpdateConfig updates the config (thread-safe)
func (m *Manager) UpdateConfig(cfg *Config) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config = cfg
}

func (m *Manager) defaultConfig() *Config {
	return &Config{
		Port:  3007,
		Title: "GSBE - GoSlimBlockExplorer",
		Nodes: []NodeConnection{},
		Logging: &LoggingConfig{
			Level:       "INFO",
			LogToFile:   false,
			LogFilePath: "logs/gsbe.log",
		},
	}
}

// GenerateID creates an 8-character random alphanumeric ID
func GenerateID() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, 8)
	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			// Fallback should never happen with crypto/rand
			result[i] = charset[0]
			continue
		}
		result[i] = charset[n.Int64()]
	}
	return string(result)
}
