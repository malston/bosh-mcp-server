// ABOUTME: Tests for environment variable auth provider.
// ABOUTME: Verifies credential extraction from BOSH_* environment variables.

package auth

import (
	"testing"
)

func TestEnvProvider_Success(t *testing.T) {
	// Set up environment
	t.Setenv("BOSH_ENVIRONMENT", "https://10.0.0.5:25555")
	t.Setenv("BOSH_CLIENT", "admin")
	t.Setenv("BOSH_CLIENT_SECRET", "secret123")
	t.Setenv("BOSH_CA_CERT", "/path/to/ca.crt")

	provider := &EnvProvider{}
	creds, err := provider.GetCredentials()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds.Environment != "https://10.0.0.5:25555" {
		t.Errorf("expected environment https://10.0.0.5:25555, got %s", creds.Environment)
	}
	if creds.Client != "admin" {
		t.Errorf("expected client admin, got %s", creds.Client)
	}
	if creds.ClientSecret != "secret123" {
		t.Errorf("expected secret secret123, got %s", creds.ClientSecret)
	}
	if creds.CACert != "/path/to/ca.crt" {
		t.Errorf("expected ca cert /path/to/ca.crt, got %s", creds.CACert)
	}
}

func TestEnvProvider_MissingRequired(t *testing.T) {
	// Don't set any environment variables
	provider := &EnvProvider{}
	creds, err := provider.GetCredentials()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds != nil {
		t.Errorf("expected nil credentials when env vars missing, got %+v", creds)
	}
}
