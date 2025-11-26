// ABOUTME: Tests for diagnostic tool handlers (vms, instances, tasks, task).
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

func TestHandleBoshVMs_Success(t *testing.T) {
	vms := []bosh.VM{
		{Job: "diego_cell", Index: 0, ProcessState: "running", IPs: []string{"10.0.1.5"}},
		{Job: "diego_cell", Index: 1, ProcessState: "running", IPs: []string{"10.0.1.6"}},
	}

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/deployments/cf/vms" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(vms)
	}))
	defer server.Close()

	t.Setenv("BOSH_ENVIRONMENT", server.URL)
	t.Setenv("BOSH_CLIENT", "admin")
	t.Setenv("BOSH_CLIENT_SECRET", "secret")

	authProvider := auth.NewProvider("")
	registry := NewRegistry(authProvider)

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "bosh_vms",
			Arguments: map[string]interface{}{
				"deployment": "cf",
			},
		},
	}

	result, err := registry.handleBoshVMs(context.Background(), request)
	if err != nil {
		t.Fatalf("handleBoshVMs failed: %v", err)
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

	vmsResponse, ok := response["vms"].([]interface{})
	if !ok {
		t.Fatal("expected vms array")
	}

	if len(vmsResponse) != 2 {
		t.Errorf("expected 2 VMs, got %d", len(vmsResponse))
	}
}

func TestHandleBoshVMs_MissingDeployment(t *testing.T) {
	authProvider := auth.NewProvider("")
	registry := NewRegistry(authProvider)

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "bosh_vms",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := registry.handleBoshVMs(context.Background(), request)
	if err != nil {
		t.Fatalf("handleBoshVMs failed: %v", err)
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

func TestHandleBoshVMs_AuthFailure(t *testing.T) {
	authProvider := auth.NewProvider("")
	registry := NewRegistry(authProvider)

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "bosh_vms",
			Arguments: map[string]interface{}{
				"deployment": "cf",
			},
		},
	}

	result, err := registry.handleBoshVMs(context.Background(), request)
	if err != nil {
		t.Fatalf("handleBoshVMs failed: %v", err)
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

func TestHandleBoshInstances_Success(t *testing.T) {
	instances := []bosh.Instance{
		{
			Job:   "diego_cell",
			Index: 0,
			State: "running",
			IPs:   []string{"10.0.1.5"},
			Processes: []bosh.Process{
				{Name: "rep", State: "running"},
			},
		},
	}

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/deployments/cf/instances" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("format") != "full" {
			t.Errorf("expected format=full, got %s", r.URL.Query().Get("format"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(instances)
	}))
	defer server.Close()

	t.Setenv("BOSH_ENVIRONMENT", server.URL)
	t.Setenv("BOSH_CLIENT", "admin")
	t.Setenv("BOSH_CLIENT_SECRET", "secret")

	authProvider := auth.NewProvider("")
	registry := NewRegistry(authProvider)

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "bosh_instances",
			Arguments: map[string]interface{}{
				"deployment": "cf",
			},
		},
	}

	result, err := registry.handleBoshInstances(context.Background(), request)
	if err != nil {
		t.Fatalf("handleBoshInstances failed: %v", err)
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
}

func TestHandleBoshInstances_MissingDeployment(t *testing.T) {
	authProvider := auth.NewProvider("")
	registry := NewRegistry(authProvider)

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "bosh_instances",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := registry.handleBoshInstances(context.Background(), request)
	if err != nil {
		t.Fatalf("handleBoshInstances failed: %v", err)
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

func TestHandleBoshInstances_AuthFailure(t *testing.T) {
	authProvider := auth.NewProvider("")
	registry := NewRegistry(authProvider)

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "bosh_instances",
			Arguments: map[string]interface{}{
				"deployment": "cf",
			},
		},
	}

	result, err := registry.handleBoshInstances(context.Background(), request)
	if err != nil {
		t.Fatalf("handleBoshInstances failed: %v", err)
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

func TestHandleBoshTasks_Success(t *testing.T) {
	tasks := []bosh.Task{
		{ID: 100, State: "done", Description: "create deployment"},
		{ID: 99, State: "done", Description: "update cloud config"},
	}

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tasks" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tasks)
	}))
	defer server.Close()

	t.Setenv("BOSH_ENVIRONMENT", server.URL)
	t.Setenv("BOSH_CLIENT", "admin")
	t.Setenv("BOSH_CLIENT_SECRET", "secret")

	authProvider := auth.NewProvider("")
	registry := NewRegistry(authProvider)

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "bosh_tasks",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := registry.handleBoshTasks(context.Background(), request)
	if err != nil {
		t.Fatalf("handleBoshTasks failed: %v", err)
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

	tasksResponse, ok := response["tasks"].([]interface{})
	if !ok {
		t.Fatal("expected tasks array")
	}

	if len(tasksResponse) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(tasksResponse))
	}
}

func TestHandleBoshTasks_AuthFailure(t *testing.T) {
	authProvider := auth.NewProvider("")
	registry := NewRegistry(authProvider)

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "bosh_tasks",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := registry.handleBoshTasks(context.Background(), request)
	if err != nil {
		t.Fatalf("handleBoshTasks failed: %v", err)
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

func TestHandleBoshTask_Success(t *testing.T) {
	task := bosh.Task{
		ID:          42,
		State:       "done",
		Description: "create deployment",
		User:        "admin",
	}

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tasks/42" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(task)
	}))
	defer server.Close()

	t.Setenv("BOSH_ENVIRONMENT", server.URL)
	t.Setenv("BOSH_CLIENT", "admin")
	t.Setenv("BOSH_CLIENT_SECRET", "secret")

	authProvider := auth.NewProvider("")
	registry := NewRegistry(authProvider)

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "bosh_task",
			Arguments: map[string]interface{}{
				"id": float64(42),
			},
		},
	}

	result, err := registry.handleBoshTask(context.Background(), request)
	if err != nil {
		t.Fatalf("handleBoshTask failed: %v", err)
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

	taskResponse, ok := response["task"].(map[string]interface{})
	if !ok {
		t.Fatal("expected task object")
	}

	if taskResponse["id"] != float64(42) {
		t.Errorf("expected task id 42, got %v", taskResponse["id"])
	}
}

func TestHandleBoshTask_MissingID(t *testing.T) {
	authProvider := auth.NewProvider("")
	registry := NewRegistry(authProvider)

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "bosh_task",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := registry.handleBoshTask(context.Background(), request)
	if err != nil {
		t.Fatalf("handleBoshTask failed: %v", err)
	}

	if !result.IsError {
		t.Fatal("expected error for missing id")
	}

	if len(result.Content) == 0 {
		t.Fatal("expected error content")
	}

	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatal("expected TextContent")
	}

	if !strings.Contains(textContent.Text, "id is required") {
		t.Errorf("expected 'id is required' error, got: %s", textContent.Text)
	}
}

func TestHandleBoshTask_AuthFailure(t *testing.T) {
	authProvider := auth.NewProvider("")
	registry := NewRegistry(authProvider)

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "bosh_task",
			Arguments: map[string]interface{}{
				"id": float64(42),
			},
		},
	}

	result, err := registry.handleBoshTask(context.Background(), request)
	if err != nil {
		t.Fatalf("handleBoshTask failed: %v", err)
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
