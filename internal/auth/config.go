// ABOUTME: Reads BOSH credentials from ~/.bosh/config file.
// ABOUTME: Supports named environments and defaults to first environment.

package auth

import (
	"os"

	"gopkg.in/yaml.v3"
)

// ConfigProvider reads credentials from BOSH config file.
type ConfigProvider struct {
	Path        string // Path to config file (default: ~/.bosh/config)
	Environment string // Named environment to use (optional)
}

type boshConfig struct {
	Environments map[string]boshEnvironment `yaml:"environments"`
}

type boshEnvironment struct {
	URL          string `yaml:"url"`
	Client       string `yaml:"client"`
	ClientSecret string `yaml:"client_secret"`
	CACert       string `yaml:"ca_cert"`
}

// GetCredentials reads credentials from BOSH config file.
// Returns nil if file doesn't exist or environment not found.
func (p *ConfigProvider) GetCredentials() (*Credentials, error) {
	path := p.Path
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, nil
		}
		path = home + "/.bosh/config"
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var config boshConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	if len(config.Environments) == 0 {
		return nil, nil
	}

	var env boshEnvironment
	var found bool

	if p.Environment != "" {
		env, found = config.Environments[p.Environment]
	} else {
		// Use first environment
		for _, e := range config.Environments {
			env = e
			found = true
			break
		}
	}

	if !found {
		return nil, nil
	}

	creds := &Credentials{
		Environment:  env.URL,
		Client:       env.Client,
		ClientSecret: env.ClientSecret,
		CACert:       env.CACert,
	}

	if !creds.Valid() {
		return nil, nil
	}

	return creds, nil
}
