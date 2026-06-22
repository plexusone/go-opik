# OmniAgent Integration

The `opik-go` SDK provides observability support for [omniagent](https://github.com/plexusone/omniagent) - a flexible agent framework for building conversational AI applications in Go.

## Overview

The integration consists of two packages:

| Package | Description |
|---------|-------------|
| `opik-go/agentobs` | Generic agent observability interfaces (media extraction, sanitization, trace management) |
| Your app's integration | Hook implementation that bridges omniagent events to Opik traces |

OmniAgent itself is **observability-agnostic**—it doesn't bundle any specific telemetry SDK. Instead, integrations are built in your application code. See the [omniagent observability guide](https://github.com/plexusone/omniagent/blob/main/docs/guides/observability.md) for details.

## Installation

```bash
# Install opik-go (includes agentobs)
go get github.com/plexusone/opik-go

# Install omniagent
go get github.com/plexusone/omniagent
```

## Quick Start

Build an Opik integration in your application:

```go
package main

import (
    "context"

    "github.com/plexusone/omniagent/agent"
    opikintegration "my-agent/integrations/opik"  // Your app's integration
    "github.com/plexusone/opik-go"
)

func main() {
    ctx := context.Background()

    // Create Opik client
    opikClient, _ := opik.NewClient(opik.WithProjectName("my-agent"))

    // Create Opik hook
    opikHook, _ := opikintegration.New(
        opikintegration.WithClient(opikClient),
        opikintegration.WithTags("env:production"),
    )

    // Create agent with hook
    myAgent, _ := agent.New(config,
        agent.WithCompiledHook(opikHook),
    )

    // Initialize hooks
    myAgent.InitHooks(ctx)
    defer myAgent.HookRegistry().Close()

    // Processing is automatically traced
    myAgent.ProcessWithSession(ctx, "session-123", "Hello!")
}
```

## Event Mapping

The integration automatically maps omniagent events to Opik traces and spans:

| OmniAgent Event | Opik Action |
|-----------------|-------------|
| `EventSessionCreated` | Create trace `agent.session.{id}` |
| `EventMessageReceived` | Create span `message.user` |
| `EventToolCalled` | Create span `tool.{name}` |
| `EventToolCompleted` | End tool span with result/error |
| `EventMessageSent` | Create span `message.assistant` |
| `EventSessionUpdated` | Touch trace (reset stale timer) |
| `EventJobExecuted` | Create span `job.{name}` |

## Configuration Options

```go
opikHook, _ := opikintegration.New(
    // Required: Opik client
    opikintegration.WithClient(opikClient),

    // Override project name (defaults to client's project)
    opikintegration.WithProjectName("my-agent"),

    // Add tags to all traces
    opikintegration.WithTags("env:prod", "version:1.0"),

    // Stale trace cleanup (default: 5min timeout, 1min sweep)
    opikintegration.WithStaleTimeout(10 * time.Minute),
    opikintegration.WithSweepInterval(2 * time.Minute),

    // Content sanitization (default: enabled)
    opikintegration.WithSanitization(true),

    // Media extraction from content (default: enabled)
    opikintegration.WithMediaExtraction(true),

    // Share trace with LLM clients for correlation (default: false)
    opikintegration.WithShareTraceWithLLM(true),

    // Custom trace name prefix (default: "agent.session.")
    opikintegration.WithTraceNamePrefix("custom."),

    // Include tool parameters in spans (default: true)
    opikintegration.WithToolParams(true),

    // Include tool results in spans (default: true)
    opikintegration.WithToolResults(true),

    // Truncate content at this length (default: 0 = no limit)
    opikintegration.WithMaxContentLength(10000),
)
```

## Stale Trace Cleanup

The integration automatically cleans up inactive traces to prevent resource leaks:

- **Stale Timeout**: Traces are closed after 5 minutes of inactivity (configurable)
- **Sweep Interval**: A background goroutine checks every 1 minute (configurable)
- **Graceful Shutdown**: All active traces are closed when the hook is closed

```go
// Configure cleanup behavior
opikHook, _ := opikintegration.New(
    opikintegration.WithClient(client),
    opikintegration.WithStaleTimeout(10 * time.Minute),  // Close after 10min idle
    opikintegration.WithSweepInterval(30 * time.Second), // Check every 30s
)
```

## Content Sanitization

The integration sanitizes message content before sending to Opik:

- Removes `<system-reminder>` and other metadata blocks
- Redacts patterns that look like API keys or secrets
- Removes ANSI escape codes from terminal output
- Collapses excessive whitespace

```go
// Disable sanitization if needed
opikHook, _ := opikintegration.New(
    opikintegration.WithClient(client),
    opikintegration.WithSanitization(false),
)
```

## Media Extraction

The integration extracts media references from message content:

- Markdown images: `![alt](url)`
- Data URLs: `data:image/png;base64,...`
- File URLs: `file:///path/to/file`
- HTTP media URLs with common extensions

```go
// Disable media extraction if not needed
opikHook, _ := opikintegration.New(
    opikintegration.WithClient(client),
    opikintegration.WithMediaExtraction(false),
)
```

## LLM Trace Correlation

When using omnillm's `TracingClient`, you can correlate LLM traces with agent traces:

```go
// Enable trace sharing
opikHook, _ := opikintegration.New(
    opikintegration.WithClient(client),
    opikintegration.WithShareTraceWithLLM(true),
)

// In your message handler, inject the trace context
func (a *MyAgent) handleMessage(ctx context.Context, sessionID, message string) {
    // Inject trace context for LLM calls
    ctx = opikHook.InjectTraceContext(ctx, sessionID)

    // Now LLM calls will appear as child spans
    response, _ := a.llmClient.Chat(ctx, message)
}
```

This creates a trace hierarchy:

```
agent.session.{id} (trace)
├── message.user (span)
├── tool.search (span)
│   └── llm-call (span, from omnillm TracingClient)
├── tool.search completed (span end)
└── message.assistant (span)
    └── llm-call (span, from omnillm TracingClient)
```

## Getting Trace Context

You can retrieve the current trace context for a session:

```go
traceID, spanID, ok := opikHook.GetTraceContext("session-123")
if ok {
    fmt.Printf("Trace: %s, Current Span: %s\n", traceID, spanID)
}
```

## The agentobs Package

The `opik-go/agentobs` package provides reusable utilities for agent observability:

### Event Types

```go
import "github.com/plexusone/opik-go/agentobs"

// Create an event
event := agentobs.NewEvent(agentobs.EventMessageReceived).
    WithSession("session-123").
    WithData("content", "Hello!")
```

### Sanitization

```go
// Sanitize content
cleaned := agentobs.Sanitize(content, agentobs.DefaultSanitizeConfig())

// Sanitize a map recursively
cleanedData := agentobs.SanitizeMap(data, agentobs.DefaultSanitizeConfig())

// Redact sensitive keys
safeData := agentobs.SanitizeMapKeys(data)
```

### Media Extraction

```go
// Extract media references from content
refs := agentobs.ExtractMedia(messageContent)
for _, ref := range refs {
    fmt.Printf("Found %s: %s\n", ref.Type, ref.URL)
}

// Strip media from content
stripped := agentobs.StripMedia(content)
```

### TraceManager

The `TraceManager` provides session-based trace lifecycle management:

```go
// Create a custom trace client adapter (implements agentobs.TraceClient)
adapter := &myTraceClientAdapter{client: opikClient}

// Create manager
manager := agentobs.NewTraceManager(adapter,
    agentobs.WithStaleTimeout(5 * time.Minute),
    agentobs.WithSweepInterval(1 * time.Minute),
)
defer manager.Close()

// Start a trace for a session
state, _ := manager.StartTrace(ctx, "session-123", "my-agent", nil)

// Create spans
span, _ := manager.StartSpan(ctx, "session-123", "tool.search", "tool", input)
manager.EndSpan(ctx, "session-123", output, nil)

// End the trace
manager.EndTrace(ctx, "session-123", finalOutput)
```

## Viewing Traces

After running your agent with the Opik integration, traces appear in the Opik dashboard:

```
agent.session.abc123 (trace)
├── message.user (span)
│   └── input: "Search for climate data"
├── tool.web_search (span)
│   ├── input: {query: "climate data 2024"}
│   └── output: {results: [...]}
├── tool.analyze (span)
│   ├── input: {data: [...]}
│   └── output: {summary: "..."}
└── message.assistant (span)
    └── output: "Here's what I found..."
```

## Best Practices

1. **Use meaningful session IDs**: Session IDs become part of trace names, so use descriptive identifiers.

2. **Close hooks properly**: Always call `hook.Close()` or `registry.Close()` to flush pending traces.

3. **Enable LLM correlation**: If using omnillm, enable `WithShareTraceWithLLM(true)` for full visibility.

4. **Configure stale timeout**: Set based on expected conversation length to avoid premature trace closure.

5. **Use tags**: Add environment and version tags for filtering in the dashboard.

## See Also

- [omniagent Documentation](https://github.com/plexusone/omniagent)
- [omnillm Integration](omnillm.md)
- [Agentic Observability Tutorial](../tutorials/agentic-observability.md)
- [Context Propagation](../core-concepts/context-propagation.md)
