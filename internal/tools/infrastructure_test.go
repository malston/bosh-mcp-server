// ABOUTME: Tests for infrastructure tool handlers (stemcells, releases, configs).
// ABOUTME: Uses httptest to mock BOSH Director and verifies handler logic.

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
		{Name: "bosh-vsphere-esxi-ubuntu-jammy-go_agent", Version: "1.200", OperatingSystem: "ubuntu-jammy"},
		{Name: "bosh-aws-xen-hvm-ubuntu-jammy-go_agent", Version: "1.199", OperatingSystem: "ubuntu-jammy"},
	}

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/stemcells" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stemcells)
	}))
	defer server.Close()

	t.Setenv("BOSH_ENVIRONMENT", server.URL)
	t.Setenv("BOSH_CLIENT", "admin")
	t.Setenv("BOSH_CLIENT_SECRET", "secret")

	authProvider := auth.NewProvider("")
	registry := NewRegistry(authProvider)

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "bosh_stemcells",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := registry.handleBoshStemcells(context.Background(), request)
	if err != nil {
		t.Fatalf("handleBoshStemcells failed: %v", err)
	}

	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	if len(result.Content) == 0 {
		t.Fatal("expected content, got empty")
	}

	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatal("expected TextContent")
	}

	var response map[string]interface{}
	if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	stemcellsResponse, ok := response["stemcells"].([]interface{})
	if !ok {
		t.Fatal("expected stemcells array")
	}

	if len(stemcellsResponse) != 2 {
		t.Errorf("expected 2 stemcells, got %d", len(stemcellsResponse))
	}

	if !strings.Contains(textContent.Text, "bosh-vsphere-esxi-ubuntu-jammy-go_agent") {
		t.Errorf("expected stemcell name in response")
	}
}

func TestHandleBoshStemcells_AuthFailure(t *testing.T) {
	authProvider := auth.NewProvider("")
	registry := NewRegistry(authProvider)

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "bosh_stemcells",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := registry.handleBoshStemcells(context.Background(), request)
	if err != nil {
		t.Fatalf("handleBoshStemcells failed: %v", err)
	}

	if !result.IsError {
		t.Fatal("expected error for auth failure")
	}

	if len(result.Content) == 0 {
		t.Fatal("expected error content")
	}

	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatal("expected TextContent")
	}

	if !strings.Contains(textContent.Text, "auth failed") {
		t.Errorf("expected 'auth failed' error, got: %s", textContent.Text)
	}
}

func TestHandleBoshReleases_Success(t *testing.T) {
	releases := []bosh.Release{
		{Name: "cf", Version: "1.0.0", CommitHash: "abc123"},
		{Name: "diego", Version: "2.0.0", CommitHash: "def456"},
	}

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/releases" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(releases)
	}))
	defer server.Close()

	t.Setenv("BOSH_ENVIRONMENT", server.URL)
	t.Setenv("BOSH_CLIENT", "admin")
	t.Setenv("BOSH_CLIENT_SECRET", "secret")

	authProvider := auth.NewProvider("")
	registry := NewRegistry(authProvider)

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "bosh_releases",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := registry.handleBoshReleases(context.Background(), request)
	if err != nil {
		t.Fatalf("handleBoshReleases failed: %v", err)
	}

	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	if len(result.Content) == 0 {
		t.Fatal("expected content, got empty")
	}

	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatal("expected TextContent")
	}

	var response map[string]interface{}
	if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	releasesResponse, ok := response["releases"].([]interface{})
	if !ok {
		t.Fatal("expected releases array")
	}

	if len(releasesResponse) != 2 {
		t.Errorf("expected 2 releases, got %d", len(releasesResponse))
	}
}

func TestHandleBoshReleases_AuthFailure(t *testing.T) {
	authProvider := auth.NewProvider("")
	registry := NewRegistry(authProvider)

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "bosh_releases",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := registry.handleBoshReleases(context.Background(), request)
	if err != nil {
		t.Fatalf("handleBoshReleases failed: %v", err)
	}

	if !result.IsError {
		t.Fatal("expected error for auth failure")
	}

	if len(result.Content) == 0 {
		t.Fatal("expected error content")
	}

	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatal("expected TextContent")
	}

	if !strings.Contains(textContent.Text, "auth failed") {
		t.Errorf("expected 'auth failed' error, got: %s", textContent.Text)
	}
}

func TestHandleBoshDeployments_Success(t *testing.T) {
	deployments := []bosh.Deployment{
		{Name: "cf", CloudConfig: "latest"},
		{Name: "diego", CloudConfig: "latest"},
	}

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/deployments" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(deployments)
	}))
	defer server.Close()

	t.Setenv("BOSH_ENVIRONMENT", server.URL)
	t.Setenv("BOSH_CLIENT", "admin")
	t.Setenv("BOSH_CLIENT_SECRET", "secret")

	authProvider := auth.NewProvider("")
	registry := NewRegistry(authProvider)

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "bosh_deployments",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := registry.handleBoshDeployments(context.Background(), request)
	if err != nil {
		t.Fatalf("handleBoshDeployments failed: %v", err)
	}

	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	if len(result.Content) == 0 {
		t.Fatal("expected content, got empty")
	}

	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatal("expected TextContent")
	}

	var response map[string]interface{}
	if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	deploymentsResponse, ok := response["deployments"].([]interface{})
	if !ok {
		t.Fatal("expected deployments array")
	}

	if len(deploymentsResponse) != 2 {
		t.Errorf("expected 2 deployments, got %d", len(deploymentsResponse))
	}
}

func TestHandleBoshDeployments_AuthFailure(t *testing.T) {
	authProvider := auth.NewProvider("")
	registry := NewRegistry(authProvider)

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "bosh_deployments",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := registry.handleBoshDeployments(context.Background(), request)
	if err != nil {
		t.Fatalf("handleBoshDeployments failed: %v", err)
	}

	if !result.IsError {
		t.Fatal("expected error for auth failure")
	}

	if len(result.Content) == 0 {
		t.Fatal("expected error content")
	}

	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatal("expected TextContent")
	}

	if !strings.Contains(textContent.Text, "auth failed") {
		t.Errorf("expected 'auth failed' error, got: %s", textContent.Text)
	}
}

func TestHandleBoshCloudConfig_Success(t *testing.T) {
	configs := []bosh.CloudConfig{
		{Properties: "azs:\n- name: z1", CreatedAt: "2024-01-01T00:00:00Z"},
	}

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/configs" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("type") != "cloud" {
			t.Errorf("expected type=cloud query param")
		}
		if r.URL.Query().Get("latest") != "true" {
			t.Errorf("expected latest=true query param")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(configs)
	}))
	defer server.Close()

	t.Setenv("BOSH_ENVIRONMENT", server.URL)
	t.Setenv("BOSH_CLIENT", "admin")
	t.Setenv("BOSH_CLIENT_SECRET", "secret")

	authProvider := auth.NewProvider("")
	registry := NewRegistry(authProvider)

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "bosh_cloud_config",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := registry.handleBoshCloudConfig(context.Background(), request)
	if err != nil {
		t.Fatalf("handleBoshCloudConfig failed: %v", err)
	}

	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	if len(result.Content) == 0 {
		t.Fatal("expected content, got empty")
	}

	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatal("expected TextContent")
	}

	var response map[string]interface{}
	if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response["cloud_config"] == nil {
		t.Error("expected cloud_config field")
	}
}

func TestHandleBoshCloudConfig_AuthFailure(t *testing.T) {
	authProvider := auth.NewProvider("")
	registry := NewRegistry(authProvider)

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "bosh_cloud_config",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := registry.handleBoshCloudConfig(context.Background(), request)
	if err != nil {
		t.Fatalf("handleBoshCloudConfig failed: %v", err)
	}

	if !result.IsError {
		t.Fatal("expected error for auth failure")
	}

	if len(result.Content) == 0 {
		t.Fatal("expected error content")
	}

	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatal("expected TextContent")
	}

	if !strings.Contains(textContent.Text, "auth failed") {
		t.Errorf("expected 'auth failed' error, got: %s", textContent.Text)
	}
}

func TestHandleBoshRuntimeConfig_Success(t *testing.T) {
	configs := []bosh.RuntimeConfig{
		{Name: "default", Properties: "releases:\n- name: bpm", CreatedAt: "2024-01-01T00:00:00Z"},
	}

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/configs" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("type") != "runtime" {
			t.Errorf("expected type=runtime query param")
		}
		if r.URL.Query().Get("latest") != "true" {
			t.Errorf("expected latest=true query param")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(configs)
	}))
	defer server.Close()

	t.Setenv("BOSH_ENVIRONMENT", server.URL)
	t.Setenv("BOSH_CLIENT", "admin")
	t.Setenv("BOSH_CLIENT_SECRET", "secret")

	authProvider := auth.NewProvider("")
	registry := NewRegistry(authProvider)

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "bosh_runtime_config",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := registry.handleBoshRuntimeConfig(context.Background(), request)
	if err != nil {
		t.Fatalf("handleBoshRuntimeConfig failed: %v", err)
	}

	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	if len(result.Content) == 0 {
		t.Fatal("expected content, got empty")
	}

	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatal("expected TextContent")
	}

	var response map[string]interface{}
	if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	runtimeConfigs, ok := response["runtime_configs"].([]interface{})
	if !ok {
		t.Fatal("expected runtime_configs array")
	}

	if len(runtimeConfigs) != 1 {
		t.Errorf("expected 1 runtime config, got %d", len(runtimeConfigs))
	}
}

func TestHandleBoshRuntimeConfig_AuthFailure(t *testing.T) {
	authProvider := auth.NewProvider("")
	registry := NewRegistry(authProvider)

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "bosh_runtime_config",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := registry.handleBoshRuntimeConfig(context.Background(), request)
	if err != nil {
		t.Fatalf("handleBoshRuntimeConfig failed: %v", err)
	}

	if !result.IsError {
		t.Fatal("expected error for auth failure")
	}

	if len(result.Content) == 0 {
		t.Fatal("expected error content")
	}

	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatal("expected TextContent")
	}

	if !strings.Contains(textContent.Text, "auth failed") {
		t.Errorf("expected 'auth failed' error, got: %s", textContent.Text)
	}
}

func TestHandleBoshCPIConfig_Success(t *testing.T) {
	configs := []bosh.CPIConfig{
		{Properties: "cpis:\n- name: vsphere", CreatedAt: "2024-01-01T00:00:00Z"},
	}

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/configs" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("type") != "cpi" {
			t.Errorf("expected type=cpi query param")
		}
		if r.URL.Query().Get("latest") != "true" {
			t.Errorf("expected latest=true query param")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(configs)
	}))
	defer server.Close()

	t.Setenv("BOSH_ENVIRONMENT", server.URL)
	t.Setenv("BOSH_CLIENT", "admin")
	t.Setenv("BOSH_CLIENT_SECRET", "secret")

	authProvider := auth.NewProvider("")
	registry := NewRegistry(authProvider)

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "bosh_cpi_config",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := registry.handleBoshCPIConfig(context.Background(), request)
	if err != nil {
		t.Fatalf("handleBoshCPIConfig failed: %v", err)
	}

	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	if len(result.Content) == 0 {
		t.Fatal("expected content, got empty")
	}

	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatal("expected TextContent")
	}

	var response map[string]interface{}
	if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response["cpi_config"] == nil {
		t.Error("expected cpi_config field")
	}
}

func TestHandleBoshCPIConfig_AuthFailure(t *testing.T) {
	authProvider := auth.NewProvider("")
	registry := NewRegistry(authProvider)

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "bosh_cpi_config",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := registry.handleBoshCPIConfig(context.Background(), request)
	if err != nil {
		t.Fatalf("handleBoshCPIConfig failed: %v", err)
	}

	if !result.IsError {
		t.Fatal("expected error for auth failure")
	}

	if len(result.Content) == 0 {
		t.Fatal("expected error content")
	}

	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatal("expected TextContent")
	}

	if !strings.Contains(textContent.Text, "auth failed") {
		t.Errorf("expected 'auth failed' error, got: %s", textContent.Text)
	}
}

func TestHandleBoshVariables_Success(t *testing.T) {
	variables := []bosh.Variable{
		{ID: "1", Name: "admin_password"},
		{ID: "2", Name: "database_password"},
	}

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/deployments/cf/variables" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(variables)
	}))
	defer server.Close()

	t.Setenv("BOSH_ENVIRONMENT", server.URL)
	t.Setenv("BOSH_CLIENT", "admin")
	t.Setenv("BOSH_CLIENT_SECRET", "secret")

	authProvider := auth.NewProvider("")
	registry := NewRegistry(authProvider)

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "bosh_variables",
			Arguments: map[string]interface{}{
				"deployment": "cf",
			},
		},
	}

	result, err := registry.handleBoshVariables(context.Background(), request)
	if err != nil {
		t.Fatalf("handleBoshVariables failed: %v", err)
	}

	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	if len(result.Content) == 0 {
		t.Fatal("expected content, got empty")
	}

	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatal("expected TextContent")
	}

	var response map[string]interface{}
	if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response["deployment"] != "cf" {
		t.Errorf("expected deployment 'cf', got %v", response["deployment"])
	}

	variablesResponse, ok := response["variables"].([]interface{})
	if !ok {
		t.Fatal("expected variables array")
	}

	if len(variablesResponse) != 2 {
		t.Errorf("expected 2 variables, got %d", len(variablesResponse))
	}
}

func TestHandleBoshVariables_MissingDeployment(t *testing.T) {
	authProvider := auth.NewProvider("")
	registry := NewRegistry(authProvider)

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "bosh_variables",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := registry.handleBoshVariables(context.Background(), request)
	if err != nil {
		t.Fatalf("handleBoshVariables failed: %v", err)
	}

	if !result.IsError {
		t.Fatal("expected error for missing deployment")
	}

	if len(result.Content) == 0 {
		t.Fatal("expected error content")
	}

	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatal("expected TextContent")
	}

	if !strings.Contains(textContent.Text, "deployment is required") {
		t.Errorf("expected 'deployment is required' error, got: %s", textContent.Text)
	}
}

func TestHandleBoshVariables_AuthFailure(t *testing.T) {
	authProvider := auth.NewProvider("")
	registry := NewRegistry(authProvider)

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "bosh_variables",
			Arguments: map[string]interface{}{
				"deployment": "cf",
			},
		},
	}

	result, err := registry.handleBoshVariables(context.Background(), request)
	if err != nil {
		t.Fatalf("handleBoshVariables failed: %v", err)
	}

	if !result.IsError {
		t.Fatal("expected error for auth failure")
	}

	if len(result.Content) == 0 {
		t.Fatal("expected error content")
	}

	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatal("expected TextContent")
	}

	if !strings.Contains(textContent.Text, "auth failed") {
		t.Errorf("expected 'auth failed' error, got: %s", textContent.Text)
	}
}

func TestHandleBoshLocks_Success(t *testing.T) {
	locks := []bosh.Lock{
		{Type: "deployment", Resource: "cf", TaskID: "123", Timeout: "900"},
		{Type: "deployment", Resource: "diego", TaskID: "124", Timeout: "900"},
	}

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/locks" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(locks)
	}))
	defer server.Close()

	t.Setenv("BOSH_ENVIRONMENT", server.URL)
	t.Setenv("BOSH_CLIENT", "admin")
	t.Setenv("BOSH_CLIENT_SECRET", "secret")

	authProvider := auth.NewProvider("")
	registry := NewRegistry(authProvider)

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "bosh_locks",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := registry.handleBoshLocks(context.Background(), request)
	if err != nil {
		t.Fatalf("handleBoshLocks failed: %v", err)
	}

	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	if len(result.Content) == 0 {
		t.Fatal("expected content, got empty")
	}

	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatal("expected TextContent")
	}

	var response map[string]interface{}
	if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	locksResponse, ok := response["locks"].([]interface{})
	if !ok {
		t.Fatal("expected locks array")
	}

	if len(locksResponse) != 2 {
		t.Errorf("expected 2 locks, got %d", len(locksResponse))
	}
}

func TestHandleBoshLocks_AuthFailure(t *testing.T) {
	authProvider := auth.NewProvider("")
	registry := NewRegistry(authProvider)

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "bosh_locks",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := registry.handleBoshLocks(context.Background(), request)
	if err != nil {
		t.Fatalf("handleBoshLocks failed: %v", err)
	}

	if !result.IsError {
		t.Fatal("expected error for auth failure")
	}

	if len(result.Content) == 0 {
		t.Fatal("expected error content")
	}

	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatal("expected TextContent")
	}

	if !strings.Contains(textContent.Text, "auth failed") {
		t.Errorf("expected 'auth failed' error, got: %s", textContent.Text)
	}
}
