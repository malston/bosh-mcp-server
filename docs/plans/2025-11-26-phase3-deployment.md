# Phase 3: Deployment Operations Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add deployment operation tools with confirmation tokens for destructive operations.

**Architecture:** Implement confirmation token system for safety. Add CLI executor for `bosh deploy` and `bosh cck`. Add API methods for other deployment operations. All operations return BOSH task IDs for async polling.

**Tech Stack:** Go 1.21+, crypto/rand for token generation, os/exec for CLI, existing auth/client infrastructure

---

## Task 1: Implement Confirmation Token System

**Files:**
- Create: `internal/confirm/tokens.go`
- Create: `internal/confirm/tokens_test.go`

**Step 1: Write failing tests**

Create `internal/confirm/tokens_test.go`:

```go
// ABOUTME: Tests for confirmation token generation and validation.
// ABOUTME: Verifies token creation, validation, and expiry behavior.

package confirm

import (
	"testing"
	"time"
)

func TestTokenStore_GenerateAndValidate(t *testing.T) {
	store := NewTokenStore(5 * time.Minute)

	token := store.Generate("delete_deployment", "cf")

	if token == "" {
		t.Fatal("expected non-empty token")
	}

	valid := store.Validate(token, "delete_deployment", "cf")
	if !valid {
		t.Error("expected token to be valid")
	}

	// Token should be consumed after validation
	valid = store.Validate(token, "delete_deployment", "cf")
	if valid {
		t.Error("expected token to be consumed after use")
	}
}

func TestTokenStore_WrongOperation(t *testing.T) {
	store := NewTokenStore(5 * time.Minute)

	token := store.Generate("delete_deployment", "cf")

	valid := store.Validate(token, "recreate", "cf")
	if valid {
		t.Error("expected token to be invalid for wrong operation")
	}
}

func TestTokenStore_WrongResource(t *testing.T) {
	store := NewTokenStore(5 * time.Minute)

	token := store.Generate("delete_deployment", "cf")

	valid := store.Validate(token, "delete_deployment", "other-deployment")
	if valid {
		t.Error("expected token to be invalid for wrong resource")
	}
}

func TestTokenStore_Expiry(t *testing.T) {
	store := NewTokenStore(50 * time.Millisecond)

	token := store.Generate("delete_deployment", "cf")

	time.Sleep(100 * time.Millisecond)

	valid := store.Validate(token, "delete_deployment", "cf")
	if valid {
		t.Error("expected token to be expired")
	}
}

func TestTokenStore_GetPendingToken(t *testing.T) {
	store := NewTokenStore(5 * time.Minute)

	token := store.Generate("delete_deployment", "cf")

	pending := store.GetPending(token)
	if pending == nil {
		t.Fatal("expected pending token info")
	}
	if pending.Operation != "delete_deployment" {
		t.Errorf("expected operation delete_deployment, got %s", pending.Operation)
	}
	if pending.Resource != "cf" {
		t.Errorf("expected resource cf, got %s", pending.Resource)
	}
}
```

**Step 2: Run tests to verify they fail**

Run:
```bash
go test ./internal/confirm/... -v
```
Expected: FAIL - package doesn't exist

**Step 3: Implement token store**

Create `internal/confirm/tokens.go`:

```go
// ABOUTME: Generates and validates confirmation tokens for destructive operations.
// ABOUTME: Tokens are single-use, time-limited, and tied to specific operations.

package confirm

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

// PendingToken holds information about a pending confirmation.
type PendingToken struct {
	Operation string
	Resource  string
	ExpiresAt time.Time
}

// TokenStore manages confirmation tokens.
type TokenStore struct {
	ttl    time.Duration
	mu     sync.Mutex
	tokens map[string]*PendingToken
}

// NewTokenStore creates a new token store with the given TTL.
func NewTokenStore(ttl time.Duration) *TokenStore {
	return &TokenStore{
		ttl:    ttl,
		tokens: make(map[string]*PendingToken),
	}
}

// Generate creates a new confirmation token for an operation.
func (s *TokenStore) Generate(operation, resource string) string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	token := "tok_" + hex.EncodeToString(bytes)

	s.mu.Lock()
	defer s.mu.Unlock()

	s.tokens[token] = &PendingToken{
		Operation: operation,
		Resource:  resource,
		ExpiresAt: time.Now().Add(s.ttl),
	}

	return token
}

// Validate checks if a token is valid for the given operation and resource.
// Valid tokens are consumed (single-use).
func (s *TokenStore) Validate(token, operation, resource string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	pending, exists := s.tokens[token]
	if !exists {
		return false
	}

	// Check expiry
	if time.Now().After(pending.ExpiresAt) {
		delete(s.tokens, token)
		return false
	}

	// Check operation and resource match
	if pending.Operation != operation || pending.Resource != resource {
		return false
	}

	// Consume token
	delete(s.tokens, token)
	return true
}

// GetPending returns information about a pending token without consuming it.
func (s *TokenStore) GetPending(token string) *PendingToken {
	s.mu.Lock()
	defer s.mu.Unlock()

	pending, exists := s.tokens[token]
	if !exists {
		return nil
	}

	if time.Now().After(pending.ExpiresAt) {
		delete(s.tokens, token)
		return nil
	}

	return pending
}

// Cleanup removes expired tokens. Call periodically if needed.
func (s *TokenStore) Cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for token, pending := range s.tokens {
		if now.After(pending.ExpiresAt) {
			delete(s.tokens, token)
		}
	}
}
```

**Step 4: Run tests to verify they pass**

Run:
```bash
go test ./internal/confirm/... -v
```
Expected: All tests pass

**Step 5: Commit**

```bash
git add internal/confirm/
git commit -m "Add confirmation token system"
```

---

## Task 2: Add Configuration System

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`

**Step 1: Write failing tests**

Create `internal/config/config_test.go`:

```go
// ABOUTME: Tests for server configuration loading.
// ABOUTME: Verifies default values and config file parsing.

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfig_Defaults(t *testing.T) {
	cfg := Load("")

	if cfg.TokenTTL != 300 {
		t.Errorf("expected default TokenTTL 300, got %d", cfg.TokenTTL)
	}

	if !cfg.RequiresConfirmation("delete_deployment") {
		t.Error("expected delete_deployment to require confirmation by default")
	}

	if !cfg.RequiresConfirmation("recreate") {
		t.Error("expected recreate to require confirmation by default")
	}

	if cfg.RequiresConfirmation("restart") {
		t.Error("expected restart to NOT require confirmation by default")
	}
}

func TestConfig_FromFile(t *testing.T) {
	// Create temp config file
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	content := []byte(`
token_ttl: 600
confirm_operations:
  - delete_deployment
  - stop
blocked_operations:
  - cck
`)
	if err := os.WriteFile(configPath, content, 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg := Load(configPath)

	if cfg.TokenTTL != 600 {
		t.Errorf("expected TokenTTL 600, got %d", cfg.TokenTTL)
	}

	if !cfg.RequiresConfirmation("delete_deployment") {
		t.Error("expected delete_deployment to require confirmation")
	}

	if !cfg.RequiresConfirmation("stop") {
		t.Error("expected stop to require confirmation")
	}

	if cfg.RequiresConfirmation("recreate") {
		t.Error("expected recreate to NOT require confirmation (not in list)")
	}

	if !cfg.IsBlocked("cck") {
		t.Error("expected cck to be blocked")
	}
}
```

**Step 2: Run tests to verify they fail**

Run:
```bash
go test ./internal/config/... -v
```
Expected: FAIL - package doesn't exist

**Step 3: Implement configuration**

Create `internal/config/config.go`:

```go
// ABOUTME: Loads server configuration from file or defaults.
// ABOUTME: Controls which operations require confirmation tokens.

package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds server configuration.
type Config struct {
	TokenTTL          int      `yaml:"token_ttl"`
	ConfirmOperations []string `yaml:"confirm_operations"`
	BlockedOperations []string `yaml:"blocked_operations"`
}

// DefaultConfirmOperations lists operations requiring confirmation by default.
var DefaultConfirmOperations = []string{
	"delete_deployment",
	"recreate",
	"stop",
	"cck",
}

// Load reads configuration from file or returns defaults.
func Load(path string) *Config {
	cfg := &Config{
		TokenTTL:          300,
		ConfirmOperations: DefaultConfirmOperations,
		BlockedOperations: []string{},
	}

	if path == "" {
		return cfg
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg
	}

	var fileCfg Config
	if err := yaml.Unmarshal(data, &fileCfg); err != nil {
		return cfg
	}

	if fileCfg.TokenTTL > 0 {
		cfg.TokenTTL = fileCfg.TokenTTL
	}
	if len(fileCfg.ConfirmOperations) > 0 {
		cfg.ConfirmOperations = fileCfg.ConfirmOperations
	}
	if len(fileCfg.BlockedOperations) > 0 {
		cfg.BlockedOperations = fileCfg.BlockedOperations
	}

	return cfg
}

// RequiresConfirmation returns true if the operation needs a confirmation token.
func (c *Config) RequiresConfirmation(operation string) bool {
	for _, op := range c.ConfirmOperations {
		if op == operation {
			return true
		}
	}
	return false
}

// IsBlocked returns true if the operation is blocked.
func (c *Config) IsBlocked(operation string) bool {
	for _, op := range c.BlockedOperations {
		if op == operation {
			return true
		}
	}
	return false
}
```

**Step 4: Run tests to verify they pass**

Run:
```bash
go test ./internal/config/... -v
```
Expected: All tests pass

**Step 5: Commit**

```bash
git add internal/config/
git commit -m "Add configuration system for confirmation settings"
```

---

## Task 3: Add Deployment API Methods to Client

**Files:**
- Modify: `internal/bosh/client.go`
- Modify: `internal/bosh/client_test.go`

**Step 1: Write failing tests**

Add to `internal/bosh/client_test.go`:

```go
func TestClient_DeleteDeployment(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/deployments/cf" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Location", "/tasks/123")
		w.WriteHeader(http.StatusFound)
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

	taskID, err := client.DeleteDeployment("cf", false)
	if err != nil {
		t.Fatalf("DeleteDeployment failed: %v", err)
	}

	if taskID != 123 {
		t.Errorf("expected task ID 123, got %d", taskID)
	}
}

func TestClient_StopInstance(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		w.Header().Set("Location", "/tasks/456")
		w.WriteHeader(http.StatusFound)
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

	taskID, err := client.ChangeJobState("cf", "diego_cell", "stopped")
	if err != nil {
		t.Fatalf("ChangeJobState failed: %v", err)
	}

	if taskID != 456 {
		t.Errorf("expected task ID 456, got %d", taskID)
	}
}

func TestClient_Recreate(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		w.Header().Set("Location", "/tasks/789")
		w.WriteHeader(http.StatusFound)
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

	taskID, err := client.Recreate("cf", "", "")
	if err != nil {
		t.Fatalf("Recreate failed: %v", err)
	}

	if taskID != 789 {
		t.Errorf("expected task ID 789, got %d", taskID)
	}
}
```

**Step 2: Run tests to verify they fail**

Run:
```bash
go test ./internal/bosh/... -v -run "TestClient_(DeleteDeployment|StopInstance|Recreate)"
```
Expected: FAIL - methods not defined

**Step 3: Implement deployment methods**

Add to `internal/bosh/client.go`:

```go
// DeleteDeployment deletes a deployment. Returns task ID.
func (c *Client) DeleteDeployment(deployment string, force bool) (int, error) {
	query := url.Values{}
	if force {
		query.Set("force", "true")
	}

	path := "/deployments/" + deployment
	return c.doAsyncRequest("DELETE", path, query)
}

// ChangeJobState changes the state of a job (start, stop, restart, detach).
// Job can be empty to target all jobs, or "job_name" or "job_name/index".
func (c *Client) ChangeJobState(deployment, job, state string) (int, error) {
	path := "/deployments/" + deployment + "/jobs"
	if job != "" {
		path += "/" + job
	}
	query := url.Values{"state": {state}}
	return c.doAsyncRequest("PUT", path, query)
}

// Recreate recreates VMs for a deployment.
// Job and index can be empty to target all, or specific job/instance.
func (c *Client) Recreate(deployment, job, index string) (int, error) {
	path := "/deployments/" + deployment
	if job != "" {
		path += "/jobs/" + job
		if index != "" {
			path += "/" + index
		}
	}
	query := url.Values{"state": {"recreate"}}
	return c.doAsyncRequest("PUT", path, query)
}

// doAsyncRequest performs a request that returns a task ID in the Location header.
func (c *Client) doAsyncRequest(method, path string, query url.Values) (int, error) {
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	req, err := http.NewRequest(method, u, nil)
	if err != nil {
		return 0, err
	}

	req.SetBasicAuth(c.creds.Client, c.creds.ClientSecret)
	req.Header.Set("Content-Type", "application/json")

	// Don't follow redirects - we need the Location header
	c.httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound && resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	location := resp.Header.Get("Location")
	if location == "" {
		return 0, fmt.Errorf("no task location in response")
	}

	// Extract task ID from location like "/tasks/123"
	var taskID int
	if _, err := fmt.Sscanf(location, "/tasks/%d", &taskID); err != nil {
		return 0, fmt.Errorf("failed to parse task ID from %s: %w", location, err)
	}

	return taskID, nil
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
git commit -m "Add deployment operation API methods"
```

---

## Task 4: Create Deployment Tool Handlers

**Files:**
- Create: `internal/tools/deployment.go`

**Step 1: Create deployment handlers file**

```go
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
```

**Step 2: Verify it compiles**

Run:
```bash
go build ./internal/tools/...
```
Expected: Success

**Step 3: Commit**

```bash
git add internal/tools/deployment.go
git commit -m "Add deployment tool handlers with confirmation support"
```

---

## Task 5: Register Deployment Tools

**Files:**
- Modify: `internal/tools/registry.go`
- Modify: `cmd/bosh-mcp-server/main.go`

**Step 1: Add deployment tool registration**

Add to `internal/tools/registry.go`:

```go
// RegisterDeploymentTools registers deployment operation tools.
func (r *DeploymentRegistry) RegisterDeploymentTools(s *server.MCPServer) {
	// bosh_delete_deployment
	s.AddTool(mcp.NewTool("bosh_delete_deployment",
		mcp.WithDescription("Delete a BOSH deployment"),
		mcp.WithString("deployment",
			mcp.Required(),
			mcp.Description("Name of the deployment to delete")),
		mcp.WithString("confirm",
			mcp.Description("Confirmation token (required for destructive operation)")),
		mcp.WithBoolean("force",
			mcp.Description("Force delete even if instances are running")),
		mcp.WithString("environment",
			mcp.Description("Named BOSH environment (optional)")),
	), r.handleBoshDeleteDeployment)

	// bosh_recreate
	s.AddTool(mcp.NewTool("bosh_recreate",
		mcp.WithDescription("Recreate VMs for a deployment"),
		mcp.WithString("deployment",
			mcp.Required(),
			mcp.Description("Name of the deployment")),
		mcp.WithString("job",
			mcp.Description("Job name to recreate (optional, all if not specified)")),
		mcp.WithString("index",
			mcp.Description("Instance index to recreate (optional)")),
		mcp.WithString("confirm",
			mcp.Description("Confirmation token (required for destructive operation)")),
		mcp.WithString("environment",
			mcp.Description("Named BOSH environment (optional)")),
	), r.handleBoshRecreate)

	// bosh_stop
	s.AddTool(mcp.NewTool("bosh_stop",
		mcp.WithDescription("Stop jobs in a deployment"),
		mcp.WithString("deployment",
			mcp.Required(),
			mcp.Description("Name of the deployment")),
		mcp.WithString("job",
			mcp.Description("Job name to stop (optional, all if not specified)")),
		mcp.WithString("confirm",
			mcp.Description("Confirmation token (required for destructive operation)")),
		mcp.WithString("environment",
			mcp.Description("Named BOSH environment (optional)")),
	), r.handleBoshStop)

	// bosh_start
	s.AddTool(mcp.NewTool("bosh_start",
		mcp.WithDescription("Start stopped jobs in a deployment"),
		mcp.WithString("deployment",
			mcp.Required(),
			mcp.Description("Name of the deployment")),
		mcp.WithString("job",
			mcp.Description("Job name to start (optional, all if not specified)")),
		mcp.WithString("environment",
			mcp.Description("Named BOSH environment (optional)")),
	), r.handleBoshStart)

	// bosh_restart
	s.AddTool(mcp.NewTool("bosh_restart",
		mcp.WithDescription("Restart jobs in a deployment"),
		mcp.WithString("deployment",
			mcp.Required(),
			mcp.Description("Name of the deployment")),
		mcp.WithString("job",
			mcp.Description("Job name to restart (optional, all if not specified)")),
		mcp.WithString("environment",
			mcp.Description("Named BOSH environment (optional)")),
	), r.handleBoshRestart)
}
```

**Step 2: Update main.go to use DeploymentRegistry**

Update `cmd/bosh-mcp-server/main.go`:

```go
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

func main() {
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
		"0.2.0",
		server.WithToolCapabilities(true),
	)

	// Register tools
	registry.RegisterTools(s)
	deploymentRegistry.RegisterDeploymentTools(s)

	// Run server with stdio transport
	return server.ServeStdio(s)
}
```

**Step 3: Verify it compiles**

Run:
```bash
go build ./cmd/bosh-mcp-server
```
Expected: Success

**Step 4: Clean up and commit**

```bash
rm -f bosh-mcp-server
git add internal/tools/registry.go cmd/bosh-mcp-server/main.go
git commit -m "Register deployment tools with confirmation support"
```

---

## Task 6: Add Deployment Tool Handler Tests

**Files:**
- Create: `internal/tools/deployment_test.go`

**Step 1: Create test file**

```go
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
		w.Header().Set("Location", "/tasks/123")
		w.WriteHeader(http.StatusFound)
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
	request.Params.Arguments["confirm"] = token

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
		w.Header().Set("Location", "/tasks/456")
		w.WriteHeader(http.StatusFound)
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
		w.Header().Set("Location", "/tasks/789")
		w.WriteHeader(http.StatusFound)
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
```

**Step 2: Run all tests**

Run:
```bash
go test ./... -v
```
Expected: All tests pass

**Step 3: Commit**

```bash
git add internal/tools/deployment_test.go
git commit -m "Add deployment tool handler tests"
```

---

## Task 7: Final Verification and Push

**Step 1: Run all tests**

Run:
```bash
go test ./... -v
```
Expected: All tests pass (should be ~55+ tests)

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
git push -u origin feature/phase3-deployment
```

**Step 4: Report completion**

Phase 3 implementation complete. The server now supports 17 tools:
- Phase 1 (4): `bosh_vms`, `bosh_instances`, `bosh_tasks`, `bosh_task`
- Phase 2 (8): `bosh_stemcells`, `bosh_releases`, `bosh_deployments`, `bosh_cloud_config`, `bosh_runtime_config`, `bosh_cpi_config`, `bosh_variables`, `bosh_locks`
- Phase 3 (5): `bosh_delete_deployment`, `bosh_recreate`, `bosh_stop`, `bosh_start`, `bosh_restart`

Confirmation token system active for: `delete_deployment`, `recreate`, `stop`, `cck`
