# BOSH MCP Server Design

A Go-based MCP server that provides BOSH Director operations to MCP clients, specifically the tanzu-cf-architect plugin.

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    MCP Protocol Layer                   │
│              (stdio transport, JSON-RPC)                │
└─────────────────────────────────────────────────────────┘
                            │
┌─────────────────────────────────────────────────────────┐
│                     Tool Registry                       │
│    (diagnostic, infrastructure, deployment tools)       │
└─────────────────────────────────────────────────────────┘
                            │
┌─────────────────────────────────────────────────────────┐
│                   Execution Engine                      │
│         ┌─────────────┬─────────────────┐               │
│         │  API Client │  CLI Executor   │               │
│         └─────────────┴─────────────────┘               │
└─────────────────────────────────────────────────────────┘
                            │
┌─────────────────────────────────────────────────────────┐
│                  Auth Provider                          │
│    (env vars → bosh config → om bosh-env)               │
└─────────────────────────────────────────────────────────┘
```

**Key decisions:**
- Single binary distribution (Go)
- stdio transport for MCP
- Layered auth resolution with clear precedence
- Hybrid execution: API for structured data, CLI for streaming/complex commands

## Authentication

The Auth Provider resolves BOSH credentials in order of precedence:

### 1. Environment Variables (highest priority)

```
BOSH_ENVIRONMENT    - Director URL
BOSH_CLIENT         - UAA client name
BOSH_CLIENT_SECRET  - UAA client secret
BOSH_CA_CERT        - CA certificate (path or inline PEM)
```

### 2. BOSH Config File (`~/.bosh/config`)

- Standard BOSH CLI config format (YAML)
- Uses the current environment from the config
- Supports named environments if specified in request

### 3. Ops Manager Integration (fallback)

- Requires `OM_TARGET`, `OM_USERNAME`, `OM_PASSWORD` (or `OM_CLIENT_ID`/`OM_CLIENT_SECRET`)
- Executes `om bosh-env` to retrieve BOSH credentials
- Caches the result for configurable TTL (default: 5 minutes)

### Auth Resolution Flow

1. Check if required BOSH_* env vars are set → use them
2. Else, check ~/.bosh/config for current environment → use it
3. Else, check if OM_* env vars are set → call om bosh-env → use result
4. Else, return authentication error

Per-request override: Tools accept an optional `environment` parameter to target a specific named environment.

## Tools

### Phase 1: Diagnostic Tools

| Tool | Description | Execution |
|------|-------------|-----------|
| `bosh_vms` | List VMs for a deployment | API |
| `bosh_instances` | List instances with process details | API |
| `bosh_tasks` | List recent tasks, filter by state/deployment | API |
| `bosh_task` | Get single task details and output | API |
| `bosh_logs` | Fetch logs for job/instance | CLI (streaming) |
| `bosh_ssh` | SSH to instance, run command | CLI (streaming) |

### Phase 2: Infrastructure Tools

| Tool | Description | Execution |
|------|-------------|-----------|
| `bosh_stemcells` | List uploaded stemcells | API |
| `bosh_releases` | List uploaded releases | API |
| `bosh_deployments` | List all deployments | API |
| `bosh_cloud_config` | Get current cloud config | API |
| `bosh_runtime_config` | Get runtime config(s) | API |
| `bosh_cpi_config` | Get CPI config | API |
| `bosh_variables` | List variables for deployment | API |
| `bosh_locks` | Show current deployment locks | API |

### Phase 3: Deployment Tools

| Tool | Description | Execution | Default Confirmation |
|------|-------------|-----------|---------------------|
| `bosh_deploy` | Deploy/update a deployment | CLI + Task | No |
| `bosh_delete_deployment` | Delete a deployment | API + Task | Yes |
| `bosh_recreate` | Recreate VMs | API + Task | Yes |
| `bosh_restart` | Restart jobs | API + Task | No |
| `bosh_stop` | Stop jobs/instances | API + Task | Yes |
| `bosh_start` | Start stopped jobs | API + Task | No |
| `bosh_cck` | Cloud check (scan & fix) | CLI + Task | Yes |

## Confirmation Token Pattern

Destructive operations use a two-phase confirmation pattern.

### Flow

1. Client calls: `bosh_delete_deployment(deployment: "old-cf")`

2. Server returns (no execution yet):
   ```json
   {
     "requires_confirmation": true,
     "confirmation_token": "tok_abc123",
     "operation": "delete_deployment",
     "deployment": "old-cf",
     "expires_in_seconds": 300
   }
   ```

3. Client confirms: `bosh_delete_deployment(deployment: "old-cf", confirm: "tok_abc123")`

4. Server executes, returns task ID:
   ```json
   {
     "task_id": 4521,
     "state": "queued"
   }
   ```

Token expiry is configurable (default: 5 minutes). Expired tokens are rejected.

## Long-Running Operations

The server supports two modes for long-running operations, chosen by the caller:

- **Task-based**: Return a BOSH task ID immediately, poll with `bosh_task` for status
- **Streaming**: Stream output back as it happens (for logs, ssh sessions)

## Configuration

### Environment Variables

```
BOSH_MCP_CONFIG        - Path to config file (default: ~/.bosh-mcp/config.yaml)
BOSH_MCP_LOG_LEVEL     - debug, info, warn, error (default: info)
BOSH_MCP_TOKEN_TTL     - Confirmation token expiry in seconds (default: 300)
```

### Config File (`~/.bosh-mcp/config.yaml`)

```yaml
# Operations requiring confirmation tokens
confirm_operations:
  - delete_deployment
  - recreate
  - stop
  - cck

# Operations blocked entirely (optional)
blocked_operations: []

# Cache TTL for om bosh-env credentials (seconds)
om_credential_cache_ttl: 300

# Named environments (supplement ~/.bosh/config)
environments:
  sandbox:
    url: https://10.0.0.5:25555
    client: admin
    client_secret: ${SANDBOX_BOSH_SECRET}
    ca_cert: /path/to/ca.crt
```

### Precedence

1. Environment variables override config file values
2. Config file values override defaults
3. `confirm_operations` is additive

## Error Handling

### Error Response Format

```json
{
  "error": {
    "code": "BOSH_AUTH_FAILED",
    "message": "Failed to authenticate to BOSH Director",
    "details": {
      "environment": "https://10.0.0.5:25555",
      "auth_method": "environment_variables",
      "underlying": "UAA token refresh failed: invalid client credentials"
    },
    "recoverable": true,
    "suggestion": "Check BOSH_CLIENT and BOSH_CLIENT_SECRET values"
  }
}
```

### Error Codes

| Code | Description |
|------|-------------|
| `BOSH_AUTH_FAILED` | Authentication/authorization failure |
| `BOSH_NOT_FOUND` | Deployment, VM, or resource not found |
| `BOSH_TASK_FAILED` | BOSH task completed with error |
| `BOSH_TIMEOUT` | Operation timed out |
| `BOSH_UNAVAILABLE` | Director unreachable |
| `CONFIRMATION_REQUIRED` | Operation needs confirmation token |
| `CONFIRMATION_EXPIRED` | Token expired, must re-request |
| `OPERATION_BLOCKED` | Operation disabled by configuration |

## Project Structure

```
bosh-mcp-server/
├── cmd/
│   └── bosh-mcp-server/
│       └── main.go              # Entry point, MCP server setup
├── internal/
│   ├── auth/
│   │   ├── provider.go          # Auth resolution (env → config → om)
│   │   ├── env.go               # Environment variable auth
│   │   ├── config.go            # ~/.bosh/config parser
│   │   └── om.go                # om bosh-env integration
│   ├── bosh/
│   │   ├── client.go            # BOSH API client
│   │   ├── cli.go               # CLI executor (ssh, logs, deploy)
│   │   └── types.go             # Shared types (VM, Task, etc.)
│   ├── tools/
│   │   ├── registry.go          # Tool registration
│   │   ├── diagnostic.go        # vms, instances, tasks, logs, ssh
│   │   ├── infrastructure.go    # stemcells, releases, configs
│   │   └── deployment.go        # deploy, delete, recreate, etc.
│   ├── confirm/
│   │   └── tokens.go            # Confirmation token generation/validation
│   └── config/
│       └── config.go            # Server configuration loading
├── go.mod
├── go.sum
└── README.md
```

### Dependencies

- `github.com/mark3labs/mcp-go` - MCP protocol implementation
- Standard library for HTTP, JSON, YAML, subprocess execution
