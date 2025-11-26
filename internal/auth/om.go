// ABOUTME: Fetches BOSH credentials from Ops Manager using om bosh-env.
// ABOUTME: Caches credentials with configurable TTL to avoid repeated calls.

package auth

import (
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// OMProvider fetches credentials from Ops Manager via om CLI.
type OMProvider struct {
	CacheTTL time.Duration // Cache TTL (default: 5 minutes)

	mu       sync.Mutex
	cached   *Credentials
	cachedAt time.Time
}

// GetCredentials fetches BOSH credentials from Ops Manager.
// Returns nil if OM environment variables are not set.
func (p *OMProvider) GetCredentials() (*Credentials, error) {
	// Check if OM credentials are available
	if os.Getenv("OM_TARGET") == "" {
		return nil, nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.isCacheValid() {
		return p.cached, nil
	}

	cmd := exec.Command("om", "bosh-env")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	creds, err := p.parseOutput(string(output))
	if err != nil {
		return nil, err
	}

	p.cached = creds
	p.cachedAt = time.Now()

	return creds, nil
}

func (p *OMProvider) isCacheValid() bool {
	if p.cached == nil {
		return false
	}
	ttl := p.CacheTTL
	if ttl == 0 {
		ttl = 5 * time.Minute
	}
	return time.Since(p.cachedAt) < ttl
}

func (p *OMProvider) parseOutput(output string) (*Credentials, error) {
	creds := &Credentials{}

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimPrefix(line, "export ")
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		value := parts[1]

		switch key {
		case "BOSH_ENVIRONMENT":
			creds.Environment = value
		case "BOSH_CLIENT":
			creds.Client = value
		case "BOSH_CLIENT_SECRET":
			creds.ClientSecret = value
		case "BOSH_CA_CERT":
			creds.CACert = value
		}
	}

	if !creds.Valid() {
		return nil, nil
	}

	return creds, nil
}
