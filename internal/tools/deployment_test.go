// ABOUTME: Tests for deployment tool handlers.
// ABOUTME: Verifies confirmation token flow and operation execution.

package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/malston/bosh-mcp-server/internal/auth"
	"github.com/malston/bosh-mcp-server/internal/config"
	"github.com/mark3labs/mcp-go/mcp"
)

func TestHandleBoshDeleteDeployment_RequiresConfirmation(t *testing.T) {
	cfg := config.Load("")
	authProvider := auth.NewProvider("")
	registry := NewRegistry(authProvider)
	deploymentRegistry := NewDeploymentRegistry(registry, cfg)

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"deployment": "cf",
	}

	result, err := deploymentRegistry.handleBoshDeleteDeployment(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Fatalf("expected success, got error")
	}

	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "requires_confirmation") {
		t.Error("expected confirmation required response")
	}
	if !strings.Contains(text, "confirmation_token") {
		t.Error("expected confirmation token in response")
	}
}

func TestHandleBoshDeleteDeployment_WithValidToken(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" {
			w.Header().Set("Location", "/tasks/123")
			w.WriteHeader(http.StatusFound)
		} else if r.Method == "GET" && strings.Contains(r.URL.Path, "/tasks/") {
			// Handle task status polling
			task := map[string]interface{}{
				"id":    123,
				"state": "done",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(task)
		}
	}))
	defer server.Close()

	t.Setenv("BOSH_ENVIRONMENT", server.URL)
	t.Setenv("BOSH_CLIENT", "admin")
	t.Setenv("BOSH_CLIENT_SECRET", "secret")

	cfg := config.Load("")
	authProvider := auth.NewProvider("")
	registry := NewRegistry(authProvider)
	deploymentRegistry := NewDeploymentRegistry(registry, cfg)

	// First call to get token
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"deployment": "cf",
	}

	result, _ := deploymentRegistry.handleBoshDeleteDeployment(context.Background(), request)
	text := result.Content[0].(mcp.TextContent).Text

	var response map[string]interface{}
	json.Unmarshal([]byte(text), &response)
	token := response["confirmation_token"].(string)

	// Second call with token
	args := request.Params.Arguments.(map[string]interface{})
	args["confirm"] = token

	result, err := deploymentRegistry.handleBoshDeleteDeployment(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	text = result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "task_id") {
		t.Error("expected task_id in response")
	}
}

func TestHandleBoshDeleteDeployment_InvalidToken(t *testing.T) {
	cfg := config.Load("")
	authProvider := auth.NewProvider("")
	registry := NewRegistry(authProvider)
	deploymentRegistry := NewDeploymentRegistry(registry, cfg)

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"deployment": "cf",
		"confirm":    "invalid_token",
	}

	result, err := deploymentRegistry.handleBoshDeleteDeployment(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsError {
		t.Fatal("expected error for invalid token")
	}
}

func TestHandleBoshStart_NoConfirmationRequired(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" {
			w.Header().Set("Location", "/tasks/456")
			w.WriteHeader(http.StatusFound)
		} else if r.Method == "GET" && strings.Contains(r.URL.Path, "/tasks/") {
			// Handle task status polling
			task := map[string]interface{}{
				"id":    456,
				"state": "done",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(task)
		}
	}))
	defer server.Close()

	t.Setenv("BOSH_ENVIRONMENT", server.URL)
	t.Setenv("BOSH_CLIENT", "admin")
	t.Setenv("BOSH_CLIENT_SECRET", "secret")

	cfg := config.Load("")
	authProvider := auth.NewProvider("")
	registry := NewRegistry(authProvider)
	deploymentRegistry := NewDeploymentRegistry(registry, cfg)

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"deployment": "cf",
	}

	result, err := deploymentRegistry.handleBoshStart(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "task_id") {
		t.Error("expected task_id in response (no confirmation needed)")
	}
}

func TestHandleBoshRestart_Success(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" {
			w.Header().Set("Location", "/tasks/789")
			w.WriteHeader(http.StatusFound)
		} else if r.Method == "GET" && strings.Contains(r.URL.Path, "/tasks/") {
			// Handle task status polling
			task := map[string]interface{}{
				"id":    789,
				"state": "done",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(task)
		}
	}))
	defer server.Close()

	t.Setenv("BOSH_ENVIRONMENT", server.URL)
	t.Setenv("BOSH_CLIENT", "admin")
	t.Setenv("BOSH_CLIENT_SECRET", "secret")

	cfg := config.Load("")
	authProvider := auth.NewProvider("")
	registry := NewRegistry(authProvider)
	deploymentRegistry := NewDeploymentRegistry(registry, cfg)

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"deployment": "cf",
		"job":        "diego_cell",
	}

	result, err := deploymentRegistry.handleBoshRestart(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}
}

func TestHandleBoshRecreate_MissingDeployment(t *testing.T) {
	cfg := config.Load("")
	authProvider := auth.NewProvider("")
	registry := NewRegistry(authProvider)
	deploymentRegistry := NewDeploymentRegistry(registry, cfg)

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{}

	result, err := deploymentRegistry.handleBoshRecreate(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsError {
		t.Fatal("expected error for missing deployment")
	}
}
