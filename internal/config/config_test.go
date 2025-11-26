// ABOUTME: Tests for server configuration loading.
// ABOUTME: Verifies default values and config file parsing.

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfig_Defaults(t *testing.T) {
	cfg := Load("")

	if cfg.TokenTTL != 300 {
		t.Errorf("expected default TokenTTL 300, got %d", cfg.TokenTTL)
	}

	if !cfg.RequiresConfirmation("delete_deployment") {
		t.Error("expected delete_deployment to require confirmation by default")
	}

	if !cfg.RequiresConfirmation("recreate") {
		t.Error("expected recreate to require confirmation by default")
	}

	if cfg.RequiresConfirmation("restart") {
		t.Error("expected restart to NOT require confirmation by default")
	}
}

func TestConfig_FromFile(t *testing.T) {
	// Create temp config file
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	content := []byte(`
token_ttl: 600
confirm_operations:
  - delete_deployment
  - stop
blocked_operations:
  - cck
`)
	if err := os.WriteFile(configPath, content, 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg := Load(configPath)

	if cfg.TokenTTL != 600 {
		t.Errorf("expected TokenTTL 600, got %d", cfg.TokenTTL)
	}

	if !cfg.RequiresConfirmation("delete_deployment") {
		t.Error("expected delete_deployment to require confirmation")
	}

	if !cfg.RequiresConfirmation("stop") {
		t.Error("expected stop to require confirmation")
	}

	if cfg.RequiresConfirmation("recreate") {
		t.Error("expected recreate to NOT require confirmation (not in list)")
	}

	if !cfg.IsBlocked("cck") {
		t.Error("expected cck to be blocked")
	}
}
