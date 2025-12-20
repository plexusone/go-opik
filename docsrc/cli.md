# CLI Reference

The Opik CLI provides command-line access to manage projects, traces, datasets, and experiments.

## Installation

```bash
go install github.com/grokify/go-comet-ml-opik/cmd/opik@latest
```

## Configuration

### Configure Credentials

```bash
opik configure -api-key=your-key -workspace=your-workspace
```

### Options

| Flag | Description |
|------|-------------|
| `-api-key` | API key for Opik Cloud |
| `-workspace` | Workspace name |
| `-url` | Custom API endpoint URL |

Configuration is saved to `~/.opik.config`.

## Commands

### Projects

List and manage projects.

```bash
# List all projects
opik projects -list

# Create a new project
opik projects -create="New Project"

# Output as JSON
opik projects -list -format=json
```

| Flag | Description |
|------|-------------|
| `-list` | List all projects |
| `-create` | Create a project with the given name |
| `-format` | Output format: `text` (default) or `json` |

### Traces

View recent traces.

```bash
# List recent traces
opik traces -list

# Filter by project
opik traces -list -project="My Project"

# Limit results
opik traces -list -limit=20

# Output as JSON
opik traces -list -format=json
```

| Flag | Description |
|------|-------------|
| `-list` | List recent traces |
| `-project` | Filter by project name |
| `-limit` | Maximum traces to show (default: 10) |
| `-format` | Output format: `text` (default) or `json` |

### Datasets

Manage evaluation datasets.

```bash
# List all datasets
opik datasets -list

# Create a new dataset
opik datasets -create="evaluation-data"

# Get dataset by name
opik datasets -get="my-dataset"

# Delete a dataset
opik datasets -delete="old-dataset"

# Output as JSON
opik datasets -list -format=json
```

| Flag | Description |
|------|-------------|
| `-list` | List all datasets |
| `-create` | Create a dataset with the given name |
| `-get` | Get a dataset by name |
| `-delete` | Delete a dataset by name |
| `-format` | Output format: `text` (default) or `json` |

### Experiments

View experiments.

```bash
# List experiments for a dataset
opik experiments -list -dataset="my-dataset"

# Output as JSON
opik experiments -list -dataset="my-dataset" -format=json
```

| Flag | Description |
|------|-------------|
| `-list` | List experiments |
| `-dataset` | Dataset name (required for listing) |
| `-format` | Output format: `text` (default) or `json` |

### Help

```bash
# Show general help
opik help

# Show command-specific help
opik projects -h
opik traces -h
opik datasets -h
opik experiments -h
```

## Environment Variables

The CLI respects these environment variables:

| Variable | Description |
|----------|-------------|
| `OPIK_API_KEY` | API key for Opik Cloud |
| `OPIK_WORKSPACE` | Workspace name |
| `OPIK_URL_OVERRIDE` | Custom API endpoint |
| `OPIK_PROJECT_NAME` | Default project name |

## Examples

### Quick Setup

```bash
# Configure credentials
opik configure -api-key=your-key -workspace=your-workspace

# Verify by listing projects
opik projects -list
```

### Daily Workflow

```bash
# Check recent traces
opik traces -list -project="production" -limit=20

# Review datasets
opik datasets -list

# Check experiment results
opik experiments -list -dataset="qa-eval"
```

### Scripting

```bash
# Export traces as JSON for analysis
opik traces -list -format=json > traces.json

# Create dataset from script
opik datasets -create="$(date +%Y%m%d)-eval"
```
