// ABOUTME: Tests for BOSH Director API client.
// ABOUTME: Uses httptest to mock Director responses.

package bosh

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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
