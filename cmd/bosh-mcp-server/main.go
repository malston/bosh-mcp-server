// ABOUTME: Entry point for the BOSH MCP server.
// ABOUTME: Initializes MCP server with stdio transport and registers tools.

package main

import (
	"fmt"
	"os"

	"github.com/malston/bosh-mcp-server/internal/auth"
	"github.com/malston/bosh-mcp-server/internal/tools"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Create auth provider
	authProvider := auth.NewProvider("")

	// Create tool registry
	registry := tools.NewRegistry(authProvider)

	// Create MCP server
	s := server.NewMCPServer(
		"bosh-mcp-server",
		"0.1.0",
		server.WithToolCapabilities(true),
	)

	// Register tools
	registry.RegisterTools(s)

	// Run server with stdio transport
	return server.ServeStdio(s)
}
