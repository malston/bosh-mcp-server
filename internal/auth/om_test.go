// ABOUTME: Tests for Ops Manager bosh-env auth provider.
// ABOUTME: Verifies parsing of om bosh-env output.

package auth

import (
	"testing"
	"time"
)

func TestOMProvider_ParseOutput(t *testing.T) {
	output := `export BOSH_CLIENT=ops_manager
export BOSH_CLIENT_SECRET=om-secret-123
export BOSH_CA_CERT=/var/tempest/workspaces/default/root_ca_certificate
export BOSH_ENVIRONMENT=10.0.0.5`

	provider := &OMProvider{}
	creds, err := provider.parseOutput(output)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds.Client != "ops_manager" {
		t.Errorf("expected client ops_manager, got %s", creds.Client)
	}
	if creds.ClientSecret != "om-secret-123" {
		t.Errorf("expected secret om-secret-123, got %s", creds.ClientSecret)
	}
	if creds.Environment != "10.0.0.5" {
		t.Errorf("expected environment 10.0.0.5, got %s", creds.Environment)
	}
}

func TestOMProvider_CacheExpiry(t *testing.T) {
	provider := &OMProvider{
		CacheTTL: 100 * time.Millisecond,
	}

	// Manually set cached credentials
	provider.cached = &Credentials{
		Environment:  "cached-env",
		Client:       "cached-client",
		ClientSecret: "cached-secret",
	}
	provider.cachedAt = time.Now()

	// Should return cached
	if !provider.isCacheValid() {
		t.Error("expected cache to be valid")
	}

	// Wait for expiry
	time.Sleep(150 * time.Millisecond)

	if provider.isCacheValid() {
		t.Error("expected cache to be expired")
	}
}
