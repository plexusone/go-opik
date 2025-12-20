# Configuration

The SDK supports multiple configuration methods. They are applied in this order (later overrides earlier):

1. Config file (`~/.opik.config`)
2. Environment variables
3. Programmatic options

## Environment Variables

| Variable | Description |
|----------|-------------|
| `OPIK_URL_OVERRIDE` | API endpoint URL |
| `OPIK_API_KEY` | API key for Opik Cloud |
| `OPIK_WORKSPACE` | Workspace name for Opik Cloud |
| `OPIK_PROJECT_NAME` | Default project name |
| `OPIK_TRACK_DISABLE` | Set to `true` to disable tracing |

### Example

```bash
export OPIK_API_KEY="your-api-key"
export OPIK_WORKSPACE="your-workspace"
export OPIK_PROJECT_NAME="my-project"
```

## Config File

Create `~/.opik.config` with INI format:

```ini
[opik]
url_override = https://www.comet.com/opik/api
api_key = your-api-key
workspace = your-workspace
project_name = My Project
```

## Programmatic Configuration

Override any configuration using functional options:

```go
client, err := opik.NewClient(
    opik.WithURL("https://www.comet.com/opik/api"),
    opik.WithAPIKey("your-api-key"),
    opik.WithWorkspace("your-workspace"),
    opik.WithProjectName("My Project"),
)
```

### Available Options

| Option | Description |
|--------|-------------|
| `WithURL(url)` | Set the API endpoint URL |
| `WithAPIKey(key)` | Set the API key |
| `WithWorkspace(name)` | Set the workspace name |
| `WithProjectName(name)` | Set the default project name |
| `WithHTTPClient(client)` | Use a custom HTTP client |

## Configure via CLI

Use the CLI to save configuration:

```bash
opik configure -api-key=your-key -workspace=your-workspace
```

This saves to `~/.opik.config`.

## Disable Tracing

To disable tracing (useful in tests or local development):

```bash
export OPIK_TRACK_DISABLE=true
```

Or programmatically:

```go
client := opik.RecordTracesLocally("my-project")
```

This records traces in memory without sending to the server.

## Load and Check Configuration

```go
// Load current configuration
cfg := opik.LoadConfig()

fmt.Printf("URL: %s\n", cfg.URL)
fmt.Printf("Workspace: %s\n", cfg.Workspace)
fmt.Printf("Project: %s\n", cfg.ProjectName)
```
