// ABOUTME: Implements diagnostic tool handlers (vms, instances, tasks).
// ABOUTME: Each handler validates input, calls BOSH API, returns structured JSON.

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/malston/bosh-mcp-server/internal/bosh"
	"github.com/mark3labs/mcp-go/mcp"
)

func (r *Registry) handleBoshVMs(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	deployment := request.GetString("deployment", "")
	environment := request.GetString("environment", "")

	if deployment == "" {
		return mcp.NewToolResultError("deployment is required"), nil
	}

	client, err := r.GetClient(environment)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("auth failed: %v", err)), nil
	}

	vms, err := client.ListVMs(deployment)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list VMs: %v", err)), nil
	}

	result := map[string]interface{}{
		"deployment": deployment,
		"vms":        vms,
	}

	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
}

func (r *Registry) handleBoshInstances(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	deployment := request.GetString("deployment", "")
	environment := request.GetString("environment", "")

	if deployment == "" {
		return mcp.NewToolResultError("deployment is required"), nil
	}

	client, err := r.GetClient(environment)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("auth failed: %v", err)), nil
	}

	instances, err := client.ListInstances(deployment)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list instances: %v", err)), nil
	}

	result := map[string]interface{}{
		"deployment": deployment,
		"instances":  instances,
	}

	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
}

func (r *Registry) handleBoshTasks(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	environment := request.GetString("environment", "")

	filter := bosh.TaskFilter{
		State:      request.GetString("state", ""),
		Deployment: request.GetString("deployment", ""),
		Limit:      request.GetInt("limit", 0),
	}

	client, err := r.GetClient(environment)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("auth failed: %v", err)), nil
	}

	tasks, err := client.ListTasks(filter)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list tasks: %v", err)), nil
	}

	result := map[string]interface{}{
		"tasks": tasks,
	}

	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
}

func (r *Registry) handleBoshTask(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	environment := request.GetString("environment", "")
	id := request.GetInt("id", 0)

	if id == 0 {
		return mcp.NewToolResultError("id is required"), nil
	}

	includeOutput := request.GetBool("output", false)

	client, err := r.GetClient(environment)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("auth failed: %v", err)), nil
	}

	task, err := client.GetTask(id)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get task: %v", err)), nil
	}

	result := map[string]interface{}{
		"task": task,
	}

	if includeOutput {
		output, err := client.GetTaskOutput(id, "result")
		if err == nil {
			result["output"] = output
		}
	}

	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
}

func (r *Registry) handleBoshTaskWait(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	environment := request.GetString("environment", "")

	idFloat := request.GetInt("id", 0)
	if idFloat == 0 {
		return mcp.NewToolResultError("id is required"), nil
	}
	id := int(idFloat)

	timeoutSecs := request.GetInt("timeout", 600) // default 10 minutes
	timeout := time.Duration(timeoutSecs) * time.Second

	client, err := r.GetClient(environment)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("auth failed: %v", err)), nil
	}

	task, err := client.WaitForTask(id, timeout, 2*time.Second)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed waiting for task: %v", err)), nil
	}

	result := map[string]interface{}{
		"task": task,
	}

	// Include output for completed tasks
	if task.State == "done" || task.State == "error" {
		output, err := client.GetTaskOutput(id, "result")
		if err == nil {
			result["output"] = output
		}
	}

	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
}
