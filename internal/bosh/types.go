// ABOUTME: Defines types for BOSH Director API responses.
// ABOUTME: Used by API client and tool implementations.

package bosh

// VM represents a BOSH VM from the /deployments/:name/vms endpoint.
type VM struct {
	VMCID        string   `json:"vm_cid"`
	Active       bool     `json:"active"`
	AgentID      string   `json:"agent_id"`
	AZ           string   `json:"az"`
	Bootstrap    bool     `json:"bootstrap"`
	Deployment   string   `json:"deployment"`
	IPs          []string `json:"ips"`
	Job          string   `json:"job"`
	Index        int      `json:"index"`
	ID           string   `json:"id"`
	ProcessState string   `json:"process_state"`
	State        string   `json:"state"`
	VMType       string   `json:"vm_type"`
	Ignore       bool     `json:"ignore"`
}

// Instance represents a BOSH instance with process details.
type Instance struct {
	AgentID    string    `json:"agent_id"`
	AZ         string    `json:"az"`
	Bootstrap  bool      `json:"bootstrap"`
	Deployment string    `json:"deployment"`
	Disk       string    `json:"disk_cid,omitempty"`
	Expects    VMState   `json:"expects_vm"`
	ID         string    `json:"id"`
	IPs        []string  `json:"ips"`
	Job        string    `json:"job"`
	Index      int       `json:"index"`
	State      string    `json:"state"`
	VMType     string    `json:"vm_type"`
	VMCID      string    `json:"vm_cid"`
	Processes  []Process `json:"processes,omitempty"`
}

// VMState represents expected VM state.
type VMState bool

// Process represents a process running on a BOSH instance.
type Process struct {
	Name   string         `json:"name"`
	State  string         `json:"state"`
	Uptime *Uptime        `json:"uptime,omitempty"`
	Memory *ResourceUsage `json:"mem,omitempty"`
	CPU    *CPUUsage      `json:"cpu,omitempty"`
}

// Uptime represents process uptime.
type Uptime struct {
	Seconds int `json:"secs"`
}

// ResourceUsage represents memory usage.
type ResourceUsage struct {
	Percent float64 `json:"percent"`
	KB      int     `json:"kb"`
}

// CPUUsage represents CPU usage.
type CPUUsage struct {
	Total float64 `json:"total"`
}

// Task represents a BOSH task.
type Task struct {
	ID          int    `json:"id"`
	State       string `json:"state"`
	Description string `json:"description"`
	Timestamp   int64  `json:"timestamp"`
	Result      string `json:"result,omitempty"`
	User        string `json:"user"`
	Deployment  string `json:"deployment,omitempty"`
	ContextID   string `json:"context_id,omitempty"`
}

// Deployment represents a BOSH deployment.
type Deployment struct {
	Name        string        `json:"name"`
	CloudConfig string        `json:"cloud_config"`
	Releases    []NameVersion `json:"releases"`
	Stemcells   []NameVersion `json:"stemcells"`
}

// NameVersion represents a name/version pair.
type NameVersion struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Stemcell represents an uploaded stemcell.
type Stemcell struct {
	Name            string   `json:"name"`
	OperatingSystem string   `json:"operating_system"`
	Version         string   `json:"version"`
	CID             string   `json:"cid"`
	Deployments     []string `json:"deployments"`
}

// Release represents an uploaded release.
type Release struct {
	Name               string `json:"name"`
	Version            string `json:"version"`
	CommitHash         string `json:"commit_hash"`
	UncommittedChanges bool   `json:"uncommitted_changes"`
}

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
