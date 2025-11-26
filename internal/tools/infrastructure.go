// ABOUTME: Implements infrastructure tool handlers (stemcells, releases, configs).
// ABOUTME: Each handler calls BOSH API and returns structured JSON.

package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

func (r *Registry) handleBoshStemcells(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	environment := request.GetString("environment", "")

	client, err := r.GetClient(environment)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("auth failed: %v", err)), nil
	}

	stemcells, err := client.ListStemcells()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list stemcells: %v", err)), nil
	}

	result := map[string]interface{}{
		"stemcells": stemcells,
	}

	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
}

func (r *Registry) handleBoshReleases(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	environment := request.GetString("environment", "")

	client, err := r.GetClient(environment)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("auth failed: %v", err)), nil
	}

	releases, err := client.ListReleases()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list releases: %v", err)), nil
	}

	result := map[string]interface{}{
		"releases": releases,
	}

	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
}

func (r *Registry) handleBoshDeployments(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	environment := request.GetString("environment", "")

	client, err := r.GetClient(environment)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("auth failed: %v", err)), nil
	}

	deployments, err := client.ListDeployments()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list deployments: %v", err)), nil
	}

	result := map[string]interface{}{
		"deployments": deployments,
	}

	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
}

func (r *Registry) handleBoshCloudConfig(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	environment := request.GetString("environment", "")

	client, err := r.GetClient(environment)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("auth failed: %v", err)), nil
	}

	config, err := client.GetCloudConfig()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get cloud config: %v", err)), nil
	}

	result := map[string]interface{}{
		"cloud_config": config,
	}

	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
}

func (r *Registry) handleBoshRuntimeConfig(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	environment := request.GetString("environment", "")

	client, err := r.GetClient(environment)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("auth failed: %v", err)), nil
	}

	configs, err := client.GetRuntimeConfigs()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get runtime configs: %v", err)), nil
	}

	result := map[string]interface{}{
		"runtime_configs": configs,
	}

	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
}

func (r *Registry) handleBoshCPIConfig(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	environment := request.GetString("environment", "")

	client, err := r.GetClient(environment)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("auth failed: %v", err)), nil
	}

	config, err := client.GetCPIConfig()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get CPI config: %v", err)), nil
	}

	result := map[string]interface{}{
		"cpi_config": config,
	}

	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
}

func (r *Registry) handleBoshVariables(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	deployment := request.GetString("deployment", "")
	environment := request.GetString("environment", "")

	if deployment == "" {
		return mcp.NewToolResultError("deployment is required"), nil
	}

	client, err := r.GetClient(environment)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("auth failed: %v", err)), nil
	}

	variables, err := client.ListVariables(deployment)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list variables: %v", err)), nil
	}

	result := map[string]interface{}{
		"deployment": deployment,
		"variables":  variables,
	}

	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
}

func (r *Registry) handleBoshLocks(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	environment := request.GetString("environment", "")

	client, err := r.GetClient(environment)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("auth failed: %v", err)), nil
	}

	locks, err := client.ListLocks()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list locks: %v", err)), nil
	}

	result := map[string]interface{}{
		"locks": locks,
	}

	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
}
