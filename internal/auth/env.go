// ABOUTME: Reads BOSH credentials from environment variables.
// ABOUTME: Returns nil credentials (not error) if env vars not set.

package auth

import "os"

// EnvProvider reads credentials from BOSH_* environment variables.
type EnvProvider struct{}

// GetCredentials reads BOSH credentials from environment variables.
// Returns nil if required variables are not set.
func (p *EnvProvider) GetCredentials() (*Credentials, error) {
	creds := &Credentials{
		Environment:  os.Getenv("BOSH_ENVIRONMENT"),
		Client:       os.Getenv("BOSH_CLIENT"),
		ClientSecret: os.Getenv("BOSH_CLIENT_SECRET"),
		CACert:       os.Getenv("BOSH_CA_CERT"),
	}

	if !creds.Valid() {
		return nil, nil
	}

	return creds, nil
}
