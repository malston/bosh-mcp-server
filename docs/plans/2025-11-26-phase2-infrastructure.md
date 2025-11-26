# Phase 2: Infrastructure Tools Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add infrastructure inspection tools for stemcells, releases, deployments, and configurations.

**Architecture:** Extend existing BOSH client with new API methods. Add infrastructure tool handlers following the same pattern as diagnostic tools. All tools are read-only API calls.

**Tech Stack:** Go 1.21+, github.com/mark3labs/mcp-go, existing auth and client infrastructure

---

## Task 1: Add Stemcell and Release Types

**Files:**
- Modify: `internal/bosh/types.go`

**Step 1: Add new types to types.go**

Add the following types after the existing `Release` struct:

```go
// CloudConfig represents a cloud config.
type CloudConfig struct {
	Properties string `json:"properties"`
	CreatedAt  string `json:"created_at"`
}

// RuntimeConfig represents a runtime config.
type RuntimeConfig struct {
	Name       string `json:"name"`
	Properties string `json:"properties"`
	CreatedAt  string `json:"created_at"`
}

// CPIConfig represents a CPI config.
type CPIConfig struct {
	Properties string `json:"properties"`
	CreatedAt  string `json:"created_at"`
}

// Variable represents a deployment variable.
type Variable struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Lock represents a deployment lock.
type Lock struct {
	Type     string `json:"type"`
	Resource string `json:"resource"`
	Timeout  string `json:"timeout"`
	TaskID   string `json:"task_id"`
}
```

**Step 2: Verify it compiles**

Run:
```bash
go build ./internal/bosh/...
```
Expected: Success

**Step 3: Commit**

```bash
git add internal/bosh/types.go
git commit -m "Add infrastructure types (configs, variables, locks)"
```

---

## Task 2: Add Infrastructure API Methods to Client

**Files:**
- Modify: `internal/bosh/client.go`
- Modify: `internal/bosh/client_test.go`

**Step 1: Write failing tests for new methods**

Add to `internal/bosh/client_test.go`:

```go
func TestClient_ListStemcells(t *testing.T) {
	stemcells := []Stemcell{
		{Name: "bosh-vsphere-esxi-ubuntu-jammy-go_agent", Version: "1.200", OperatingSystem: "ubuntu-jammy"},
	}

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/stemcells" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stemcells)
	}))
	defer server.Close()

	creds := &auth.Credentials{
		Environment:  server.URL,
		Client:       "admin",
		ClientSecret: "secret",
	}

	client, err := NewClient(creds)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	result, err := client.ListStemcells()
	if err != nil {
		t.Fatalf("ListStemcells failed: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("expected 1 stemcell, got %d", len(result))
	}
	if result[0].Name != "bosh-vsphere-esxi-ubuntu-jammy-go_agent" {
		t.Errorf("expected stemcell name, got %s", result[0].Name)
	}
}

func TestClient_ListReleases(t *testing.T) {
	releases := []Release{
		{Name: "cf", Version: "1.0.0", CommitHash: "abc123"},
	}

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/releases" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(releases)
	}))
	defer server.Close()

	creds := &auth.Credentials{
		Environment:  server.URL,
		Client:       "admin",
		ClientSecret: "secret",
	}

	client, err := NewClient(creds)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	result, err := client.ListReleases()
	if err != nil {
		t.Fatalf("ListReleases failed: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("expected 1 release, got %d", len(result))
	}
}

func TestClient_GetCloudConfig(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/configs" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("type") != "cloud" {
			t.Errorf("expected type=cloud query param")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]CloudConfig{{Properties: "azs: []", CreatedAt: "2024-01-01"}})
	}))
	defer server.Close()

	creds := &auth.Credentials{
		Environment:  server.URL,
		Client:       "admin",
		ClientSecret: "secret",
	}

	client, err := NewClient(creds)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	result, err := client.GetCloudConfig()
	if err != nil {
		t.Fatalf("GetCloudConfig failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected cloud config, got nil")
	}
}

func TestClient_ListLocks(t *testing.T) {
	locks := []Lock{
		{Type: "deployment", Resource: "cf", TaskID: "123"},
	}

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/locks" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(locks)
	}))
	defer server.Close()

	creds := &auth.Credentials{
		Environment:  server.URL,
		Client:       "admin",
		ClientSecret: "secret",
	}

	client, err := NewClient(creds)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	result, err := client.ListLocks()
	if err != nil {
		t.Fatalf("ListLocks failed: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("expected 1 lock, got %d", len(result))
	}
}
```

**Step 2: Run tests to verify they fail**

Run:
```bash
go test ./internal/bosh/... -v -run "TestClient_List(Stemcells|Releases)|TestClient_Get|TestClient_ListLocks"
```
Expected: FAIL - methods not defined

**Step 3: Implement the new client methods**

Add to `internal/bosh/client.go` after `ListDeployments`:

```go
// ListStemcells returns all uploaded stemcells.
func (c *Client) ListStemcells() ([]Stemcell, error) {
	body, err := c.doRequest("GET", "/stemcells", nil)
	if err != nil {
		return nil, err
	}

	var stemcells []Stemcell
	if err := json.Unmarshal(body, &stemcells); err != nil {
		return nil, err
	}

	return stemcells, nil
}

// ListReleases returns all uploaded releases.
func (c *Client) ListReleases() ([]Release, error) {
	body, err := c.doRequest("GET", "/releases", nil)
	if err != nil {
		return nil, err
	}

	var releases []Release
	if err := json.Unmarshal(body, &releases); err != nil {
		return nil, err
	}

	return releases, nil
}

// GetCloudConfig returns the current cloud config.
func (c *Client) GetCloudConfig() (*CloudConfig, error) {
	query := url.Values{"type": {"cloud"}, "latest": {"true"}}
	body, err := c.doRequest("GET", "/configs", query)
	if err != nil {
		return nil, err
	}

	var configs []CloudConfig
	if err := json.Unmarshal(body, &configs); err != nil {
		return nil, err
	}

	if len(configs) == 0 {
		return nil, nil
	}

	return &configs[0], nil
}

// GetRuntimeConfigs returns all runtime configs.
func (c *Client) GetRuntimeConfigs() ([]RuntimeConfig, error) {
	query := url.Values{"type": {"runtime"}, "latest": {"true"}}
	body, err := c.doRequest("GET", "/configs", query)
	if err != nil {
		return nil, err
	}

	var configs []RuntimeConfig
	if err := json.Unmarshal(body, &configs); err != nil {
		return nil, err
	}

	return configs, nil
}

// GetCPIConfig returns the current CPI config.
func (c *Client) GetCPIConfig() (*CPIConfig, error) {
	query := url.Values{"type": {"cpi"}, "latest": {"true"}}
	body, err := c.doRequest("GET", "/configs", query)
	if err != nil {
		return nil, err
	}

	var configs []CPIConfig
	if err := json.Unmarshal(body, &configs); err != nil {
		return nil, err
	}

	if len(configs) == 0 {
		return nil, nil
	}

	return &configs[0], nil
}

// ListVariables returns variables for a deployment.
func (c *Client) ListVariables(deployment string) ([]Variable, error) {
	body, err := c.doRequest("GET", "/deployments/"+deployment+"/variables", nil)
	if err != nil {
		return nil, err
	}

	var variables []Variable
	if err := json.Unmarshal(body, &variables); err != nil {
		return nil, err
	}

	return variables, nil
}

// ListLocks returns current deployment locks.
func (c *Client) ListLocks() ([]Lock, error) {
	body, err := c.doRequest("GET", "/locks", nil)
	if err != nil {
		return nil, err
	}

	var locks []Lock
	if err := json.Unmarshal(body, &locks); err != nil {
		return nil, err
	}

	return locks, nil
}
```

**Step 4: Run tests to verify they pass**

Run:
```bash
go test ./internal/bosh/... -v
```
Expected: All tests pass

**Step 5: Commit**

```bash
git add internal/bosh/client.go internal/bosh/client_test.go
git commit -m "Add infrastructure API methods to BOSH client"
```

---

## Task 3: Create Infrastructure Tool Handlers

**Files:**
- Create: `internal/tools/infrastructure.go`

**Step 1: Create infrastructure handlers file**

```go
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
```

**Step 2: Verify it compiles**

Run:
```bash
go build ./internal/tools/...
```
Expected: Success

**Step 3: Commit**

```bash
git add internal/tools/infrastructure.go
git commit -m "Add infrastructure tool handlers"
```

---

## Task 4: Register Infrastructure Tools

**Files:**
- Modify: `internal/tools/registry.go`

**Step 1: Add infrastructure tool registration**

Add a new method and call it from `RegisterTools`. After `registerDiagnosticTools(s)`, add:

```go
// RegisterTools registers all tools with the MCP server.
func (r *Registry) RegisterTools(s *server.MCPServer) {
	r.registerDiagnosticTools(s)
	r.registerInfrastructureTools(s)
}

func (r *Registry) registerInfrastructureTools(s *server.MCPServer) {
	// bosh_stemcells
	s.AddTool(mcp.NewTool("bosh_stemcells",
		mcp.WithDescription("List uploaded stemcells"),
		mcp.WithString("environment",
			mcp.Description("Named BOSH environment (optional)")),
	), r.handleBoshStemcells)

	// bosh_releases
	s.AddTool(mcp.NewTool("bosh_releases",
		mcp.WithDescription("List uploaded releases"),
		mcp.WithString("environment",
			mcp.Description("Named BOSH environment (optional)")),
	), r.handleBoshReleases)

	// bosh_deployments
	s.AddTool(mcp.NewTool("bosh_deployments",
		mcp.WithDescription("List all deployments"),
		mcp.WithString("environment",
			mcp.Description("Named BOSH environment (optional)")),
	), r.handleBoshDeployments)

	// bosh_cloud_config
	s.AddTool(mcp.NewTool("bosh_cloud_config",
		mcp.WithDescription("Get current cloud config"),
		mcp.WithString("environment",
			mcp.Description("Named BOSH environment (optional)")),
	), r.handleBoshCloudConfig)

	// bosh_runtime_config
	s.AddTool(mcp.NewTool("bosh_runtime_config",
		mcp.WithDescription("Get runtime configs"),
		mcp.WithString("environment",
			mcp.Description("Named BOSH environment (optional)")),
	), r.handleBoshRuntimeConfig)

	// bosh_cpi_config
	s.AddTool(mcp.NewTool("bosh_cpi_config",
		mcp.WithDescription("Get CPI config"),
		mcp.WithString("environment",
			mcp.Description("Named BOSH environment (optional)")),
	), r.handleBoshCPIConfig)

	// bosh_variables
	s.AddTool(mcp.NewTool("bosh_variables",
		mcp.WithDescription("List variables for a deployment"),
		mcp.WithString("deployment",
			mcp.Required(),
			mcp.Description("Name of the deployment")),
		mcp.WithString("environment",
			mcp.Description("Named BOSH environment (optional)")),
	), r.handleBoshVariables)

	// bosh_locks
	s.AddTool(mcp.NewTool("bosh_locks",
		mcp.WithDescription("Show current deployment locks"),
		mcp.WithString("environment",
			mcp.Description("Named BOSH environment (optional)")),
	), r.handleBoshLocks)
}
```

**Step 2: Verify it compiles**

Run:
```bash
go build ./cmd/bosh-mcp-server
```
Expected: Success

**Step 3: Clean up and commit**

```bash
rm -f bosh-mcp-server
git add internal/tools/registry.go
git commit -m "Register infrastructure tools"
```

---

## Task 5: Add Infrastructure Tool Handler Tests

**Files:**
- Create: `internal/tools/infrastructure_test.go`

**Step 1: Create test file**

```go
// ABOUTME: Tests for infrastructure tool handlers.
// ABOUTME: Uses httptest to mock BOSH Director API responses.

package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/malston/bosh-mcp-server/internal/auth"
	"github.com/malston/bosh-mcp-server/internal/bosh"
	"github.com/mark3labs/mcp-go/mcp"
)

func TestHandleBoshStemcells_Success(t *testing.T) {
	stemcells := []bosh.Stemcell{
		{Name: "bosh-vsphere-esxi-ubuntu-jammy-go_agent", Version: "1.200"},
	}

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stemcells)
	}))
	defer server.Close()

	t.Setenv("BOSH_ENVIRONMENT", server.URL)
	t.Setenv("BOSH_CLIENT", "admin")
	t.Setenv("BOSH_CLIENT_SECRET", "secret")

	authProvider := auth.NewProvider("")
	registry := NewRegistry(authProvider)

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{}

	result, err := registry.handleBoshStemcells(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "bosh-vsphere-esxi-ubuntu-jammy-go_agent") {
		t.Errorf("expected stemcell name in result, got %s", text)
	}
}

func TestHandleBoshStemcells_AuthFailure(t *testing.T) {
	t.Setenv("BOSH_ENVIRONMENT", "")
	t.Setenv("BOSH_CLIENT", "")
	t.Setenv("BOSH_CLIENT_SECRET", "")

	authProvider := auth.NewProvider("/nonexistent/path")
	registry := NewRegistry(authProvider)

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{}

	result, err := registry.handleBoshStemcells(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsError {
		t.Fatal("expected error result")
	}
}

func TestHandleBoshReleases_Success(t *testing.T) {
	releases := []bosh.Release{
		{Name: "cf", Version: "1.0.0"},
	}

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(releases)
	}))
	defer server.Close()

	t.Setenv("BOSH_ENVIRONMENT", server.URL)
	t.Setenv("BOSH_CLIENT", "admin")
	t.Setenv("BOSH_CLIENT_SECRET", "secret")

	authProvider := auth.NewProvider("")
	registry := NewRegistry(authProvider)

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{}

	result, err := registry.handleBoshReleases(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}
}

func TestHandleBoshDeployments_Success(t *testing.T) {
	deployments := []bosh.Deployment{
		{Name: "cf", CloudConfig: "latest"},
	}

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(deployments)
	}))
	defer server.Close()

	t.Setenv("BOSH_ENVIRONMENT", server.URL)
	t.Setenv("BOSH_CLIENT", "admin")
	t.Setenv("BOSH_CLIENT_SECRET", "secret")

	authProvider := auth.NewProvider("")
	registry := NewRegistry(authProvider)

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{}

	result, err := registry.handleBoshDeployments(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}
}

func TestHandleBoshVariables_Success(t *testing.T) {
	variables := []bosh.Variable{
		{ID: "1", Name: "admin_password"},
	}

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(variables)
	}))
	defer server.Close()

	t.Setenv("BOSH_ENVIRONMENT", server.URL)
	t.Setenv("BOSH_CLIENT", "admin")
	t.Setenv("BOSH_CLIENT_SECRET", "secret")

	authProvider := auth.NewProvider("")
	registry := NewRegistry(authProvider)

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"deployment": "cf",
	}

	result, err := registry.handleBoshVariables(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}
}

func TestHandleBoshVariables_MissingDeployment(t *testing.T) {
	authProvider := auth.NewProvider("")
	registry := NewRegistry(authProvider)

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{}

	result, err := registry.handleBoshVariables(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsError {
		t.Fatal("expected error for missing deployment")
	}
}

func TestHandleBoshLocks_Success(t *testing.T) {
	locks := []bosh.Lock{
		{Type: "deployment", Resource: "cf", TaskID: "123"},
	}

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(locks)
	}))
	defer server.Close()

	t.Setenv("BOSH_ENVIRONMENT", server.URL)
	t.Setenv("BOSH_CLIENT", "admin")
	t.Setenv("BOSH_CLIENT_SECRET", "secret")

	authProvider := auth.NewProvider("")
	registry := NewRegistry(authProvider)

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{}

	result, err := registry.handleBoshLocks(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}
}
```

**Step 2: Run all tests**

Run:
```bash
go test ./... -v
```
Expected: All tests pass

**Step 3: Commit**

```bash
git add internal/tools/infrastructure_test.go
git commit -m "Add infrastructure tool handler tests"
```

---

## Task 6: Final Verification and Push

**Step 1: Run all tests**

Run:
```bash
go test ./... -v
```
Expected: All tests pass (should be ~32+ tests now)

**Step 2: Build binary**

Run:
```bash
go build -o bosh-mcp-server ./cmd/bosh-mcp-server
./bosh-mcp-server --help 2>&1 || echo "Server runs (no --help flag)"
rm bosh-mcp-server
```
Expected: Binary builds successfully

**Step 3: Push feature branch**

Run:
```bash
git push -u origin feature/phase2-infrastructure
```

**Step 4: Report completion**

Phase 2 implementation complete. The server now supports 12 tools:
- Phase 1 (4): `bosh_vms`, `bosh_instances`, `bosh_tasks`, `bosh_task`
- Phase 2 (8): `bosh_stemcells`, `bosh_releases`, `bosh_deployments`, `bosh_cloud_config`, `bosh_runtime_config`, `bosh_cpi_config`, `bosh_variables`, `bosh_locks`
