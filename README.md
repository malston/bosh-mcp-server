# bosh-mcp-server

An MCP (Model Context Protocol) server that provides BOSH Director operations to AI assistants like Claude.

## Features

- **18 BOSH tools** for diagnostics, infrastructure inspection, and deployment operations
- **Layered authentication**: environment variables → ~/.bosh/config → Ops Manager
- **Confirmation tokens** for destructive operations (configurable)
- **Async task handling**: deployment operations wait for completion by default

## Installation

### From Source

```bash
git clone https://github.com/malston/bosh-mcp-server.git
cd bosh-mcp-server
go build -o bosh-mcp-server ./cmd/bosh-mcp-server
```

### Binary

Download the latest release from the [releases page](https://github.com/malston/bosh-mcp-server/releases).

## Configuration

### Authentication

The server resolves BOSH credentials in order of precedence:

1. **Environment variables** (highest priority)
   ```bash
   export BOSH_ENVIRONMENT=https://10.0.0.5:25555
   export BOSH_CLIENT=admin
   export BOSH_CLIENT_SECRET=secret
   export BOSH_CA_CERT=/path/to/ca.crt
   ```

2. **BOSH config file** (`~/.bosh/config`)
   - Standard BOSH CLI configuration format
   - Supports named environments

3. **Ops Manager** (fallback)
   ```bash
   export OM_TARGET=https://opsman.example.com
   export OM_USERNAME=admin
   export OM_PASSWORD=secret
   ```
   The server calls `om bosh-env` and caches credentials for 5 minutes.

### Server Configuration

Optional configuration via `~/.bosh-mcp/config.yaml`:

```yaml
# Token TTL for confirmation tokens (seconds)
token_ttl: 300

# Operations requiring confirmation tokens
confirm_operations:
  - delete_deployment
  - recreate
  - stop
  - cck

# Operations blocked entirely
blocked_operations: []
```

Set `BOSH_MCP_CONFIG` to use a custom config path.

## Usage with Claude Desktop

Add to your Claude Desktop configuration (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "bosh": {
      "command": "/path/to/bosh-mcp-server",
      "env": {
        "BOSH_ENVIRONMENT": "https://10.0.0.5:25555",
        "BOSH_CLIENT": "admin",
        "BOSH_CLIENT_SECRET": "your-secret",
        "BOSH_CA_CERT": "/path/to/ca.crt"
      }
    }
  }
}
```

## Available Tools

### Diagnostic Tools

| Tool | Description |
|------|-------------|
| `bosh_vms` | List VMs for a deployment |
| `bosh_instances` | List instances with process details |
| `bosh_tasks` | List recent BOSH tasks |
| `bosh_task` | Get details of a specific task |
| `bosh_task_wait` | Wait for a task to complete |

### Infrastructure Tools

| Tool | Description |
|------|-------------|
| `bosh_stemcells` | List uploaded stemcells |
| `bosh_releases` | List uploaded releases |
| `bosh_deployments` | List all deployments |
| `bosh_cloud_config` | Get current cloud config |
| `bosh_runtime_config` | Get runtime configs |
| `bosh_cpi_config` | Get CPI config |
| `bosh_variables` | List variables for a deployment |
| `bosh_locks` | Show current deployment locks |

### Deployment Tools

| Tool | Description | Confirmation Required |
|------|-------------|----------------------|
| `bosh_delete_deployment` | Delete a deployment | Yes |
| `bosh_recreate` | Recreate VMs | Yes |
| `bosh_stop` | Stop jobs | Yes |
| `bosh_start` | Start jobs | No |
| `bosh_restart` | Restart jobs | No |

All deployment tools wait for task completion by default (configurable timeout).

## Confirmation Token Flow

Destructive operations require a two-step confirmation:

1. **Request operation** (without confirm parameter):
   ```
   bosh_delete_deployment(deployment: "my-app")
   → {"requires_confirmation": true, "confirmation_token": "tok_abc123", ...}
   ```

2. **Confirm operation** (with token):
   ```
   bosh_delete_deployment(deployment: "my-app", confirm: "tok_abc123")
   → {"task_id": 456, "state": "done", ...}
   ```

Tokens expire after 5 minutes (configurable) and are single-use.

## Development

### Prerequisites

- Go 1.21+

### Building

```bash
go build ./cmd/bosh-mcp-server
```

### Testing

```bash
go test ./... -v
```

### Manual Testing with Claude Code

A [mock BOSH Director](https://github.com/malston/bosh-mock-director) is available for manually testing the MCP server with Claude Code without needing a real BOSH environment.

**Setup:**

```bash
# Clone and build the mock director
git clone https://github.com/malston/bosh-mock-director.git
cd bosh-mock-director
go build -o mock-bosh-director ./cmd/mock-bosh-director

# Start the mock director (Terminal 1)
./mock-bosh-director

# Build the MCP server
cd /path/to/bosh-mcp-server
go build -o bosh-mcp-server ./cmd/bosh-mcp-server

# Start Claude Code from the mock-director directory (Terminal 2)
# This picks up the pre-configured .claude/mcp.json
cd /path/to/bosh-mock-director
claude
```

The mock director includes realistic sample data:
- 3 deployments (`cf`, `redis`, `mysql`) with VMs and instances
- Stemcells, releases, and configs
- Task simulation with state progression
- State mutations for destructive operations

**Example prompts to try:**
- "What deployments are running on BOSH?"
- "Show me the VMs in the cf deployment"
- "List the recent BOSH tasks"
- "Stop the router job in cf" (triggers confirmation)

### Project Structure

```
├── cmd/bosh-mcp-server/    # Entry point
├── internal/
│   ├── auth/               # Authentication providers
│   ├── bosh/               # BOSH API client
│   ├── config/             # Server configuration
│   ├── confirm/            # Confirmation token system
│   └── tools/              # MCP tool handlers
└── test/                   # Integration tests
```

## License

MIT License - see [LICENSE](LICENSE) for details.
