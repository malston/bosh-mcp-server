// ABOUTME: Registers MCP tools and provides access to BOSH client.
// ABOUTME: Acts as dependency injection container for tool handlers.

package tools

import (
	"github.com/malston/bosh-mcp-server/internal/auth"
	"github.com/malston/bosh-mcp-server/internal/bosh"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Registry holds tool dependencies and registrations.
type Registry struct {
	authProvider *auth.Provider
}

// NewRegistry creates a tool registry with the given auth provider.
func NewRegistry(authProvider *auth.Provider) *Registry {
	return &Registry{
		authProvider: authProvider,
	}
}

// GetClient returns a BOSH client for the given environment.
func (r *Registry) GetClient(environment string) (*bosh.Client, error) {
	creds, err := r.authProvider.GetCredentials(environment)
	if err != nil {
		return nil, err
	}
	return bosh.NewClient(creds)
}

// RegisterTools registers all tools with the MCP server.
func (r *Registry) RegisterTools(s *server.MCPServer) {
	r.registerDiagnosticTools(s)
}

func (r *Registry) registerDiagnosticTools(s *server.MCPServer) {
	// bosh_vms
	s.AddTool(mcp.NewTool("bosh_vms",
		mcp.WithDescription("List VMs for a BOSH deployment"),
		mcp.WithString("deployment",
			mcp.Required(),
			mcp.Description("Name of the deployment")),
		mcp.WithString("environment",
			mcp.Description("Named BOSH environment (optional)")),
	), r.handleBoshVMs)

	// bosh_instances
	s.AddTool(mcp.NewTool("bosh_instances",
		mcp.WithDescription("List instances with process details for a BOSH deployment"),
		mcp.WithString("deployment",
			mcp.Required(),
			mcp.Description("Name of the deployment")),
		mcp.WithString("environment",
			mcp.Description("Named BOSH environment (optional)")),
	), r.handleBoshInstances)

	// bosh_tasks
	s.AddTool(mcp.NewTool("bosh_tasks",
		mcp.WithDescription("List recent BOSH tasks"),
		mcp.WithString("state",
			mcp.Description("Filter by state: queued, processing, done, error")),
		mcp.WithString("deployment",
			mcp.Description("Filter by deployment name")),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of tasks to return")),
		mcp.WithString("environment",
			mcp.Description("Named BOSH environment (optional)")),
	), r.handleBoshTasks)

	// bosh_task
	s.AddTool(mcp.NewTool("bosh_task",
		mcp.WithDescription("Get details of a specific BOSH task"),
		mcp.WithNumber("id",
			mcp.Required(),
			mcp.Description("Task ID")),
		mcp.WithBoolean("output",
			mcp.Description("Include task output")),
		mcp.WithString("environment",
			mcp.Description("Named BOSH environment (optional)")),
	), r.handleBoshTask)
}
