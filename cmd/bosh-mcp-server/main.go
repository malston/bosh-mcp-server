// ABOUTME: Entry point for the BOSH MCP server.
// ABOUTME: Initializes MCP server with stdio transport and registers tools.

package main

import (
	"fmt"
	"os"

	"github.com/malston/bosh-mcp-server/internal/auth"
	"github.com/malston/bosh-mcp-server/internal/config"
	"github.com/malston/bosh-mcp-server/internal/tools"
	"github.com/mark3labs/mcp-go/server"
)

// version is set at build time via -ldflags
var version = "dev"

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("bosh-mcp-server %s\n", version)
		os.Exit(0)
	}

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Load configuration
	configPath := os.Getenv("BOSH_MCP_CONFIG")
	cfg := config.Load(configPath)

	// Create auth provider
	authProvider := auth.NewProvider("")

	// Create tool registry
	registry := tools.NewRegistry(authProvider)

	// Create deployment registry with confirmation support
	deploymentRegistry := tools.NewDeploymentRegistry(registry, cfg)

	// Create MCP server
	s := server.NewMCPServer(
		"bosh-mcp-server",
		version,
		server.WithToolCapabilities(true),
	)

	// Register tools
	registry.RegisterTools(s)
	deploymentRegistry.RegisterDeploymentTools(s)

	// Run server with stdio transport
	return server.ServeStdio(s)
}
