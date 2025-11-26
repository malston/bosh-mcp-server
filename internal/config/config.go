// ABOUTME: Loads server configuration from file or defaults.
// ABOUTME: Controls which operations require confirmation tokens.

package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds server configuration.
type Config struct {
	TokenTTL          int      `yaml:"token_ttl"`
	ConfirmOperations []string `yaml:"confirm_operations"`
	BlockedOperations []string `yaml:"blocked_operations"`
}

// DefaultConfirmOperations lists operations requiring confirmation by default.
var DefaultConfirmOperations = []string{
	"delete_deployment",
	"recreate",
	"stop",
	"cck",
}

// Load reads configuration from file or returns defaults.
func Load(path string) *Config {
	cfg := &Config{
		TokenTTL:          300,
		ConfirmOperations: DefaultConfirmOperations,
		BlockedOperations: []string{},
	}

	if path == "" {
		return cfg
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg
	}

	var fileCfg Config
	if err := yaml.Unmarshal(data, &fileCfg); err != nil {
		return cfg
	}

	if fileCfg.TokenTTL > 0 {
		cfg.TokenTTL = fileCfg.TokenTTL
	}
	if len(fileCfg.ConfirmOperations) > 0 {
		cfg.ConfirmOperations = fileCfg.ConfirmOperations
	}
	if len(fileCfg.BlockedOperations) > 0 {
		cfg.BlockedOperations = fileCfg.BlockedOperations
	}

	return cfg
}

// RequiresConfirmation returns true if the operation needs a confirmation token.
func (c *Config) RequiresConfirmation(operation string) bool {
	for _, op := range c.ConfirmOperations {
		if op == operation {
			return true
		}
	}
	return false
}

// IsBlocked returns true if the operation is blocked.
func (c *Config) IsBlocked(operation string) bool {
	for _, op := range c.BlockedOperations {
		if op == operation {
			return true
		}
	}
	return false
}
