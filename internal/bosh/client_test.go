// ABOUTME: Tests for BOSH Director API client.
// ABOUTME: Uses httptest to mock Director responses.

package bosh

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/malston/bosh-mcp-server/internal/auth"
)

func TestClient_ListVMs(t *testing.T) {
	vms := []VM{
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

	creds := &auth.Credentials{
		Environment:  server.URL,
		Client:       "admin",
		ClientSecret: "secret",
	}

	client, err := NewClient(creds)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	result, err := client.ListVMs("cf")
	if err != nil {
		t.Fatalf("ListVMs failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 VMs, got %d", len(result))
	}
	if result[0].Job != "diego_cell" {
		t.Errorf("expected diego_cell, got %s", result[0].Job)
	}
}

func TestClient_ListTasks(t *testing.T) {
	tasks := []Task{
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

	creds := &auth.Credentials{
		Environment:  server.URL,
		Client:       "admin",
		ClientSecret: "secret",
	}

	client, err := NewClient(creds)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	result, err := client.ListTasks(TaskFilter{})
	if err != nil {
		t.Fatalf("ListTasks failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(result))
	}
}

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

func TestClient_WaitForTask(t *testing.T) {
	callCount := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		task := Task{ID: 123, State: "processing", Description: "test task"}
		if callCount >= 3 {
			task.State = "done"
			task.Result = "success"
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(task)
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

	task, err := client.WaitForTask(123, 10*time.Second, 10*time.Millisecond)
	if err != nil {
		t.Fatalf("WaitForTask failed: %v", err)
	}

	if task.State != "done" {
		t.Errorf("expected state done, got %s", task.State)
	}
	if callCount < 3 {
		t.Errorf("expected at least 3 poll calls, got %d", callCount)
	}
}
