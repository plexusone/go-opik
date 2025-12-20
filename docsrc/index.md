# Go Opik SDK

Go SDK for [Opik](https://github.com/comet-ml/opik) - an open-source LLM observability platform by Comet ML.

## Features

- **Tracing**: Create traces and spans to monitor LLM application execution
- **Context Propagation**: Automatic parent-child relationships using Go context
- **Distributed Tracing**: Propagate traces across service boundaries
- **Datasets**: Manage evaluation datasets with CRUD operations
- **Experiments**: Run and track evaluation experiments
- **Prompts**: Version-controlled prompt templates with variable substitution
- **Evaluation Framework**: Heuristic and LLM-based metrics for output evaluation
- **LLM Integrations**: Auto-tracing for OpenAI, Anthropic, and gollm
- **CLI Tool**: Command-line interface for management tasks

## Quick Start

```go
package main

import (
    "context"
    "log"

    opik "github.com/grokify/go-comet-ml-opik"
)

func main() {
    // Create client (uses OPIK_API_KEY and OPIK_WORKSPACE env vars)
    client, err := opik.NewClient(
        opik.WithProjectName("My Project"),
    )
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // Create a trace
    trace, _ := client.Trace(ctx, "my-trace",
        opik.WithTraceInput(map[string]any{"prompt": "Hello"}),
    )

    // Create a span for an LLM call
    span, _ := trace.Span(ctx, "llm-call",
        opik.WithSpanType(opik.SpanTypeLLM),
        opik.WithSpanModel("gpt-4"),
    )

    // Do your LLM call here...

    // End span with output
    span.End(ctx, opik.WithSpanOutput(map[string]any{"response": "Hi!"}))

    // End trace
    trace.End(ctx)
}
```

## Installation

```bash
go get github.com/grokify/go-comet-ml-opik
```

## Next Steps

- [Installation & Configuration](getting-started/installation.md) - Set up the SDK
- [Traces and Spans](core-concepts/traces-and-spans.md) - Learn the core concepts
- [Evaluation Framework](evaluation/overview.md) - Evaluate LLM outputs
- [Integrations](integrations/openai.md) - Auto-trace OpenAI, Anthropic, and more
- [Testing](getting-started/testing.md) - Run the test suite (no API key required)
