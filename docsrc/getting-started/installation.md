# Installation

## Requirements

- Go 1.21 or later

## Install the SDK

```bash
go get github.com/grokify/go-comet-ml-opik
```

## Install the CLI (Optional)

```bash
go install github.com/grokify/go-comet-ml-opik/cmd/opik@latest
```

## Verify Installation

```go
package main

import (
    "fmt"

    opik "github.com/grokify/go-comet-ml-opik"
)

func main() {
    client, err := opik.NewClient()
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    fmt.Println("Opik client created successfully!")
    _ = client
}
```

## Opik Server Options

### Opik Cloud (Recommended)

Use the hosted Opik Cloud service:

1. Sign up at [comet.com](https://www.comet.com/)
2. Get your API key and workspace name
3. Set environment variables:

```bash
export OPIK_API_KEY="your-api-key"
export OPIK_WORKSPACE="your-workspace"
```

### Self-Hosted

Run Opik locally using Docker:

```bash
# Clone the Opik repository
git clone https://github.com/comet-ml/opik.git
cd opik

# Start with Docker Compose
docker-compose up -d
```

The local server runs at `http://localhost:5173` by default.

```bash
export OPIK_URL_OVERRIDE="http://localhost:5173/api"
```

## Next Steps

- [Configuration](configuration.md) - Configure the SDK for your environment
- [Testing](testing.md) - Run the test suite (no API key required)
- [Traces and Spans](../core-concepts/traces-and-spans.md) - Start tracing your LLM calls
