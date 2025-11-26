// ABOUTME: Implements deployment tool handlers (delete, recreate, stop, start, restart).
// ABOUTME: Uses confirmation tokens for destructive operations.

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/malston/bosh-mcp-server/internal/confirm"
	"github.com/malston/bosh-mcp-server/internal/config"
	"github.com/mark3labs/mcp-go/mcp"
)

// DeploymentRegistry extends Registry with confirmation support.
type DeploymentRegistry struct {
	*Registry
	tokenStore *confirm.TokenStore
	config     *config.Config
}

// NewDeploymentRegistry creates a registry with confirmation token support.
func NewDeploymentRegistry(registry *Registry, cfg *config.Config) *DeploymentRegistry {
	ttl := time.Duration(cfg.TokenTTL) * time.Second
	return &DeploymentRegistry{
		Registry:   registry,
		tokenStore: confirm.NewTokenStore(ttl),
		config:     cfg,
	}
}

func (r *DeploymentRegistry) handleBoshDeleteDeployment(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	deployment := request.GetString("deployment", "")
	environment := request.GetString("environment", "")
	confirmToken := request.GetString("confirm", "")
	force := request.GetBool("force", false)

	if deployment == "" {
		return mcp.NewToolResultError("deployment is required"), nil
	}

	if r.config.IsBlocked("delete_deployment") {
		return mcp.NewToolResultError("delete_deployment is blocked by configuration"), nil
	}

	// Check if confirmation required
	if r.config.RequiresConfirmation("delete_deployment") {
		if confirmToken == "" {
			// Generate confirmation token
			token := r.tokenStore.Generate("delete_deployment", deployment)
			result := map[string]interface{}{
				"requires_confirmation": true,
				"confirmation_token":    token,
				"operation":             "delete_deployment",
				"deployment":            deployment,
				"expires_in_seconds":    r.config.TokenTTL,
			}
			jsonBytes, _ := json.MarshalIndent(result, "", "  ")
			return mcp.NewToolResultText(string(jsonBytes)), nil
		}

		// Validate confirmation token
		if !r.tokenStore.Validate(confirmToken, "delete_deployment", deployment) {
			return mcp.NewToolResultError("invalid or expired confirmation token"), nil
		}
	}

	client, err := r.GetClient(environment)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("auth failed: %v", err)), nil
	}

	taskID, err := client.DeleteDeployment(deployment, force)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to delete deployment: %v", err)), nil
	}

	result := map[string]interface{}{
		"task_id":    taskID,
		"state":      "queued",
		"deployment": deployment,
	}
	jsonBytes, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(jsonBytes)), nil
}

func (r *DeploymentRegistry) handleBoshRecreate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	deployment := request.GetString("deployment", "")
	job := request.GetString("job", "")
	index := request.GetString("index", "")
	environment := request.GetString("environment", "")
	confirmToken := request.GetString("confirm", "")

	if deployment == "" {
		return mcp.NewToolResultError("deployment is required"), nil
	}

	resource := deployment
	if job != "" {
		resource = deployment + "/" + job
	}

	if r.config.RequiresConfirmation("recreate") {
		if confirmToken == "" {
			token := r.tokenStore.Generate("recreate", resource)
			result := map[string]interface{}{
				"requires_confirmation": true,
				"confirmation_token":    token,
				"operation":             "recreate",
				"deployment":            deployment,
				"job":                   job,
				"expires_in_seconds":    r.config.TokenTTL,
			}
			jsonBytes, _ := json.MarshalIndent(result, "", "  ")
			return mcp.NewToolResultText(string(jsonBytes)), nil
		}

		if !r.tokenStore.Validate(confirmToken, "recreate", resource) {
			return mcp.NewToolResultError("invalid or expired confirmation token"), nil
		}
	}

	client, err := r.GetClient(environment)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("auth failed: %v", err)), nil
	}

	taskID, err := client.Recreate(deployment, job, index)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to recreate: %v", err)), nil
	}

	result := map[string]interface{}{
		"task_id":    taskID,
		"state":      "queued",
		"deployment": deployment,
	}
	jsonBytes, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(jsonBytes)), nil
}

func (r *DeploymentRegistry) handleBoshStop(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	deployment := request.GetString("deployment", "")
	job := request.GetString("job", "")
	environment := request.GetString("environment", "")
	confirmToken := request.GetString("confirm", "")

	if deployment == "" {
		return mcp.NewToolResultError("deployment is required"), nil
	}

	resource := deployment
	if job != "" {
		resource = deployment + "/" + job
	}

	if r.config.RequiresConfirmation("stop") {
		if confirmToken == "" {
			token := r.tokenStore.Generate("stop", resource)
			result := map[string]interface{}{
				"requires_confirmation": true,
				"confirmation_token":    token,
				"operation":             "stop",
				"deployment":            deployment,
				"job":                   job,
				"expires_in_seconds":    r.config.TokenTTL,
			}
			jsonBytes, _ := json.MarshalIndent(result, "", "  ")
			return mcp.NewToolResultText(string(jsonBytes)), nil
		}

		if !r.tokenStore.Validate(confirmToken, "stop", resource) {
			return mcp.NewToolResultError("invalid or expired confirmation token"), nil
		}
	}

	client, err := r.GetClient(environment)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("auth failed: %v", err)), nil
	}

	taskID, err := client.ChangeJobState(deployment, job, "stopped")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to stop: %v", err)), nil
	}

	result := map[string]interface{}{
		"task_id":    taskID,
		"state":      "queued",
		"deployment": deployment,
	}
	jsonBytes, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(jsonBytes)), nil
}

func (r *DeploymentRegistry) handleBoshStart(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	deployment := request.GetString("deployment", "")
	job := request.GetString("job", "")
	environment := request.GetString("environment", "")

	if deployment == "" {
		return mcp.NewToolResultError("deployment is required"), nil
	}

	// start doesn't require confirmation by default

	client, err := r.GetClient(environment)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("auth failed: %v", err)), nil
	}

	taskID, err := client.ChangeJobState(deployment, job, "started")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to start: %v", err)), nil
	}

	result := map[string]interface{}{
		"task_id":    taskID,
		"state":      "queued",
		"deployment": deployment,
	}
	jsonBytes, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(jsonBytes)), nil
}

func (r *DeploymentRegistry) handleBoshRestart(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	deployment := request.GetString("deployment", "")
	job := request.GetString("job", "")
	environment := request.GetString("environment", "")

	if deployment == "" {
		return mcp.NewToolResultError("deployment is required"), nil
	}

	// restart doesn't require confirmation by default

	client, err := r.GetClient(environment)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("auth failed: %v", err)), nil
	}

	taskID, err := client.ChangeJobState(deployment, job, "restart")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to restart: %v", err)), nil
	}

	result := map[string]interface{}{
		"task_id":    taskID,
		"state":      "queued",
		"deployment": deployment,
	}
	jsonBytes, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(jsonBytes)), nil
}
