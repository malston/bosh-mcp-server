// ABOUTME: Defines credential types for BOSH Director authentication.
// ABOUTME: Used by all auth providers (env, config, om).

package auth

// Credentials holds BOSH Director authentication details.
type Credentials struct {
	Environment  string // BOSH Director URL
	Client       string // UAA client name
	ClientSecret string // UAA client secret
	CACert       string // CA certificate (path or PEM content)
}

// Valid returns true if minimum required fields are set.
func (c *Credentials) Valid() bool {
	return c.Environment != "" && c.Client != "" && c.ClientSecret != ""
}
