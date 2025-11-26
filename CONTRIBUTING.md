# Contributing to bosh-mcp-server

Thank you for your interest in contributing to bosh-mcp-server!

## Getting Started

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/bosh-mcp-server.git
   cd bosh-mcp-server
   ```
3. Add the upstream remote:
   ```bash
   git remote add upstream https://github.com/malston/bosh-mcp-server.git
   ```

## Development Setup

### Prerequisites

- Go 1.21 or later
- Access to a BOSH Director (for integration testing)

### Building

```bash
go build ./cmd/bosh-mcp-server
```

### Running Tests

```bash
go test ./... -v
```

All tests must pass before submitting a pull request.

## Making Changes

### Branch Naming

Create a feature branch from `main`:

```bash
git checkout -b feature/your-feature-name
```

### Code Style

- Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
- Follow [Google Go Style Guide](https://google.github.io/styleguide/go/)
- All code files must start with a 2-line ABOUTME comment:
  ```go
  // ABOUTME: Brief description of what this file does.
  // ABOUTME: Additional context about its purpose.
  ```
- Run `go fmt` before committing

### Testing

- Write tests for new functionality
- Follow TDD: write failing test first, then implement
- Use `httptest.NewTLSServer` for mocking BOSH Director
- Use `t.Setenv()` for environment variables in tests

### Commit Messages

Write clear, concise commit messages:

```
Add bosh_task_wait tool for polling task completion

- Implement WaitForTask method in BOSH client
- Add handleBoshTaskWait handler
- Register tool with timeout parameter
```

## Pull Request Process

1. Ensure all tests pass
2. Update documentation if needed
3. Create a pull request against `main`
4. Fill out the PR template with:
   - Summary of changes
   - Test plan
5. Wait for review

## Adding New Tools

When adding a new BOSH tool:

1. **Add API method** to `internal/bosh/client.go` (if needed)
2. **Add handler** to appropriate file in `internal/tools/`
3. **Register tool** in `internal/tools/registry.go`
4. **Add tests** for both client and handler
5. **Update README.md** with the new tool

### Tool Handler Pattern

```go
func (r *Registry) handleBoshExample(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    // 1. Extract parameters
    param := request.GetString("param", "")
    environment := request.GetString("environment", "")

    // 2. Validate required parameters
    if param == "" {
        return mcp.NewToolResultError("param is required"), nil
    }

    // 3. Get authenticated client
    client, err := r.GetClient(environment)
    if err != nil {
        return mcp.NewToolResultError(fmt.Sprintf("auth failed: %v", err)), nil
    }

    // 4. Call BOSH API
    result, err := client.SomeMethod(param)
    if err != nil {
        return mcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
    }

    // 5. Return JSON result
    jsonBytes, _ := json.MarshalIndent(result, "", "  ")
    return mcp.NewToolResultText(string(jsonBytes)), nil
}
```

## Questions?

Open an issue for questions or discussion.
