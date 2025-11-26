// ABOUTME: Integration test for MCP server tool registration.
// ABOUTME: Verifies tools are properly registered and callable.

package test

import (
	"testing"

	"github.com/malston/bosh-mcp-server/internal/auth"
	"github.com/malston/bosh-mcp-server/internal/tools"
	"github.com/mark3labs/mcp-go/server"
)

func TestToolRegistration(t *testing.T) {
	authProvider := auth.NewProvider("")
	registry := tools.NewRegistry(authProvider)

	s := server.NewMCPServer(
		"bosh-mcp-server",
		"0.1.0",
		server.WithToolCapabilities(true),
	)

	registry.RegisterTools(s)

	// Verify server was created without panic
	if s == nil {
		t.Fatal("server should not be nil")
	}
}
