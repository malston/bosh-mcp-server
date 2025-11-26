// ABOUTME: Tests for the auth provider chain.
// ABOUTME: Verifies precedence: env vars > config file > om bosh-env.

package auth

import (
	"path/filepath"
	"testing"
)

func TestProvider_EnvTakesPrecedence(t *testing.T) {
	// Set env vars
	t.Setenv("BOSH_ENVIRONMENT", "https://env.example.com:25555")
	t.Setenv("BOSH_CLIENT", "env-client")
	t.Setenv("BOSH_CLIENT_SECRET", "env-secret")

	// Also have config file
	configPath := filepath.Join("testdata", "bosh-config.yml")

	provider := NewProvider(configPath)
	creds, err := provider.GetCredentials("")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should use env vars, not config file
	if creds.Environment != "https://env.example.com:25555" {
		t.Errorf("expected env URL, got %s", creds.Environment)
	}
	if creds.Client != "env-client" {
		t.Errorf("expected env-client, got %s", creds.Client)
	}
}

func TestProvider_FallsBackToConfig(t *testing.T) {
	// Clear env vars
	t.Setenv("BOSH_ENVIRONMENT", "")
	t.Setenv("BOSH_CLIENT", "")
	t.Setenv("BOSH_CLIENT_SECRET", "")
	t.Setenv("OM_TARGET", "")

	configPath := filepath.Join("testdata", "bosh-config.yml")

	provider := NewProvider(configPath)
	creds, err := provider.GetCredentials("")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should use config file
	if creds.Client != "admin" {
		t.Errorf("expected admin from config, got %s", creds.Client)
	}
}

func TestProvider_NamedEnvironment(t *testing.T) {
	// Clear env vars
	t.Setenv("BOSH_ENVIRONMENT", "")
	t.Setenv("BOSH_CLIENT", "")
	t.Setenv("BOSH_CLIENT_SECRET", "")

	configPath := filepath.Join("testdata", "bosh-config.yml")

	provider := NewProvider(configPath)
	creds, err := provider.GetCredentials("sandbox")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds.Client != "sandbox-admin" {
		t.Errorf("expected sandbox-admin, got %s", creds.Client)
	}
}
