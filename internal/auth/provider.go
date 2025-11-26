// ABOUTME: Chains auth providers with defined precedence.
// ABOUTME: Resolves credentials: env vars > config file > om bosh-env.

package auth

import "fmt"

// Provider chains multiple auth providers with precedence.
type Provider struct {
	env    *EnvProvider
	config *ConfigProvider
	om     *OMProvider
}

// NewProvider creates a provider chain with default config path.
func NewProvider(configPath string) *Provider {
	return &Provider{
		env:    &EnvProvider{},
		config: &ConfigProvider{Path: configPath},
		om:     &OMProvider{},
	}
}

// GetCredentials resolves credentials using provider chain.
// If environment is specified, it targets that named environment in config.
func (p *Provider) GetCredentials(environment string) (*Credentials, error) {
	// 1. Try environment variables (highest priority)
	creds, err := p.env.GetCredentials()
	if err != nil {
		return nil, fmt.Errorf("env provider: %w", err)
	}
	if creds != nil {
		return creds, nil
	}

	// 2. Try config file
	if environment != "" {
		p.config.Environment = environment
	}
	creds, err = p.config.GetCredentials()
	if err != nil {
		return nil, fmt.Errorf("config provider: %w", err)
	}
	if creds != nil {
		return creds, nil
	}

	// 3. Try om bosh-env
	creds, err = p.om.GetCredentials()
	if err != nil {
		return nil, fmt.Errorf("om provider: %w", err)
	}
	if creds != nil {
		return creds, nil
	}

	return nil, fmt.Errorf("no BOSH credentials available")
}
