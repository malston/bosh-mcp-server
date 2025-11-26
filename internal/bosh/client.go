// ABOUTME: HTTP client for BOSH Director REST API.
// ABOUTME: Handles authentication, TLS, and request construction.

package bosh

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/malston/bosh-mcp-server/internal/auth"
)

// Client communicates with the BOSH Director API.
type Client struct {
	baseURL    string
	httpClient *http.Client
	creds      *auth.Credentials
}

// TaskFilter specifies task list filters.
type TaskFilter struct {
	State      string // Filter by state (queued, processing, done, error, etc.)
	Deployment string // Filter by deployment name
	Limit      int    // Maximum number of tasks to return
}

// NewClient creates a new BOSH API client.
func NewClient(creds *auth.Credentials) (*Client, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // Default for test servers
	}

	// Load CA cert if provided
	if creds.CACert != "" {
		caCertPool := x509.NewCertPool()
		var caCert []byte
		var err error

		if strings.HasPrefix(creds.CACert, "-----BEGIN") {
			caCert = []byte(creds.CACert)
		} else {
			caCert, err = os.ReadFile(creds.CACert)
			if err != nil {
				return nil, fmt.Errorf("failed to read CA cert: %w", err)
			}
		}

		if ok := caCertPool.AppendCertsFromPEM(caCert); ok {
			tlsConfig = &tls.Config{
				RootCAs: caCertPool,
			}
		}
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	return &Client{
		baseURL:    strings.TrimSuffix(creds.Environment, "/"),
		httpClient: httpClient,
		creds:      creds,
	}, nil
}

func (c *Client) doRequest(method, path string, query url.Values) ([]byte, error) {
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	req, err := http.NewRequest(method, u, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.creds.Client, c.creds.ClientSecret)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// ListVMs returns VMs for a deployment.
func (c *Client) ListVMs(deployment string) ([]VM, error) {
	body, err := c.doRequest("GET", "/deployments/"+deployment+"/vms", nil)
	if err != nil {
		return nil, err
	}

	var vms []VM
	if err := json.Unmarshal(body, &vms); err != nil {
		return nil, err
	}

	return vms, nil
}

// ListInstances returns instances with process details for a deployment.
func (c *Client) ListInstances(deployment string) ([]Instance, error) {
	query := url.Values{"format": {"full"}}
	body, err := c.doRequest("GET", "/deployments/"+deployment+"/instances", query)
	if err != nil {
		return nil, err
	}

	var instances []Instance
	if err := json.Unmarshal(body, &instances); err != nil {
		return nil, err
	}

	return instances, nil
}

// ListTasks returns tasks matching the filter.
func (c *Client) ListTasks(filter TaskFilter) ([]Task, error) {
	query := url.Values{}
	if filter.State != "" {
		query.Set("state", filter.State)
	}
	if filter.Deployment != "" {
		query.Set("deployment", filter.Deployment)
	}
	if filter.Limit > 0 {
		query.Set("limit", strconv.Itoa(filter.Limit))
	}

	body, err := c.doRequest("GET", "/tasks", query)
	if err != nil {
		return nil, err
	}

	var tasks []Task
	if err := json.Unmarshal(body, &tasks); err != nil {
		return nil, err
	}

	return tasks, nil
}

// GetTask returns a single task by ID.
func (c *Client) GetTask(id int) (*Task, error) {
	body, err := c.doRequest("GET", "/tasks/"+strconv.Itoa(id), nil)
	if err != nil {
		return nil, err
	}

	var task Task
	if err := json.Unmarshal(body, &task); err != nil {
		return nil, err
	}

	return &task, nil
}

// GetTaskOutput returns the output of a task.
func (c *Client) GetTaskOutput(id int, outputType string) (string, error) {
	if outputType == "" {
		outputType = "result"
	}
	query := url.Values{"type": {outputType}}
	body, err := c.doRequest("GET", "/tasks/"+strconv.Itoa(id)+"/output", query)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// ListDeployments returns all deployments.
func (c *Client) ListDeployments() ([]Deployment, error) {
	body, err := c.doRequest("GET", "/deployments", nil)
	if err != nil {
		return nil, err
	}

	var deployments []Deployment
	if err := json.Unmarshal(body, &deployments); err != nil {
		return nil, err
	}

	return deployments, nil
}

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
