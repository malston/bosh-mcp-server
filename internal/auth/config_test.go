// ABOUTME: Tests for BOSH config file auth provider.
// ABOUTME: Verifies parsing of ~/.bosh/config format.

package auth

import (
	"path/filepath"
	"testing"
)

func TestConfigProvider_Success(t *testing.T) {
	configPath := filepath.Join("testdata", "bosh-config.yml")
	provider := &ConfigProvider{Path: configPath}

	creds, err := provider.GetCredentials()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds == nil {
		t.Fatal("expected credentials, got nil")
	}
	if creds.Environment != "https://10.0.0.5:25555" {
		t.Errorf("expected environment https://10.0.0.5:25555, got %s", creds.Environment)
	}
	if creds.Client != "admin" {
		t.Errorf("expected client admin, got %s", creds.Client)
	}
}

func TestConfigProvider_NamedEnvironment(t *testing.T) {
	configPath := filepath.Join("testdata", "bosh-config.yml")
	provider := &ConfigProvider{Path: configPath, Environment: "sandbox"}

	creds, err := provider.GetCredentials()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds == nil {
		t.Fatal("expected credentials, got nil")
	}
	if creds.Environment != "https://sandbox.example.com:25555" {
		t.Errorf("expected sandbox URL, got %s", creds.Environment)
	}
	if creds.Client != "sandbox-admin" {
		t.Errorf("expected sandbox-admin, got %s", creds.Client)
	}
}

func TestConfigProvider_FileNotFound(t *testing.T) {
	provider := &ConfigProvider{Path: "/nonexistent/path"}

	creds, err := provider.GetCredentials()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds != nil {
		t.Errorf("expected nil credentials for missing file, got %+v", creds)
	}
}
