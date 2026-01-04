---
marp: true
theme: vibeminds
paginate: true
style: |
  .mermaid-container {
    display: flex;
    justify-content: center;
    align-items: center;
    width: 100%;
    margin: 0.5em 0;
  }

  .mermaid svg {
    max-height: 320px;
    width: auto;
  }

  section.section-divider {
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    text-align: center;
    background: linear-gradient(135deg, #1a1a3e 0%, #4a3f8a 50%, #2d2d5a 100%);
  }

  section.section-divider h1 {
    font-size: 3.5em;
    margin-bottom: 0.2em;
  }

  section.section-divider h2 {
    font-size: 1.5em;
    color: #b39ddb;
    font-weight: 400;
  }
---

<script type="module">
import mermaid from 'https://cdn.jsdelivr.net/npm/mermaid@10/dist/mermaid.esm.min.mjs';
mermaid.initialize({
  startOnLoad: true,
  theme: 'dark',
  themeVariables: {
    background: 'transparent',
    primaryColor: '#7c4dff',
    primaryTextColor: '#e8eaf6',
    primaryBorderColor: '#667eea',
    lineColor: '#b39ddb',
    secondaryColor: '#302b63',
    tertiaryColor: '#24243e'
  }
});
</script>

<!-- _paginate: false -->

# go-opik

## A Go SDK for LLM Observability

**Comet Opik** | Open Source | Production Ready

---

# What is Opik?

**Opik** is an open-source LLM observability platform by Comet ML

- **Traces & Spans** - Track LLM calls and application flow
- **Datasets** - Store test data for evaluation
- **Experiments** - Run and compare model evaluations
- **Prompts** - Version and manage prompt templates
- **Feedback** - Capture quality scores and user feedback

**go-opik** provides full Go SDK support for all Opik features

---

# Quick Start

```go
import opik "github.com/agentplexus/go-opik"

client, _ := opik.NewClient(
    opik.WithAPIKey("your-api-key"),
    opik.WithProjectName("my-project"),
)

// Create a trace
trace, _ := client.Trace(ctx, "chat-request",
    opik.WithTraceInput(map[string]any{"query": "Hello!"}),
)

// Create an LLM span
span, _ := trace.Span(ctx, "gpt-4-call",
    opik.WithSpanType(opik.SpanTypeLLM),
    opik.WithSpanModel("gpt-4"),
)

// End with output and usage
span.End(ctx, opik.WithSpanOutput(response))
trace.End(ctx)
```

---

# Architecture

```
go-opik/
├── *.go                    # Core SDK (client, trace, span, config)
├── llmops/                 # OmniObserve provider adapter
├── evaluation/
│   ├── heuristic/          # BLEU, ROUGE, Levenshtein, etc.
│   └── llm/                # LLM-as-judge metrics
├── integrations/
│   ├── openai/             # OpenAI tracing
│   ├── anthropic/          # Anthropic tracing
│   └── omnillm/            # OmniLLM adapter
├── middleware/             # HTTP tracing middleware
└── cmd/opik/               # CLI tool
```

---

# Core Features

<div style="display: grid; grid-template-columns: 1fr 1fr; gap: 2rem;">
<div>

**Tracing**
- Traces and nested spans
- LLM, Tool, General span types
- Distributed trace propagation
- Streaming support
- Automatic batching

**Data Management**
- Dataset CRUD + items
- Experiment tracking
- Prompt versioning
- Template rendering

</div>
<div>

**Evaluation**
- 10+ heuristic metrics
- 8+ LLM judge metrics
- G-EVAL implementation
- Custom judge support

**Integrations**
- OpenAI tracing transport
- Anthropic tracing transport
- OmniLLM adapter
- OmniObserve llmops provider
- HTTP middleware

</div>
</div>

---

# OmniObserve Integration

Use go-opik through the unified OmniObserve abstraction:

```go
import (
    "github.com/agentplexus/omniobserve/llmops"
    _ "github.com/agentplexus/go-opik/llmops"  // Register provider
)

// Open via OmniObserve
provider, _ := llmops.Open("opik",
    llmops.WithAPIKey("your-api-key"),
    llmops.WithProjectName("my-project"),
)

// Use unified interface
ctx, trace, _ := provider.StartTrace(ctx, "workflow")
ctx, span, _ := provider.StartSpan(ctx, "llm-call",
    llmops.WithSpanType(llmops.SpanTypeLLM),
)
```

Switch between Opik, Langfuse, and Phoenix without code changes.

---

# Evaluation Metrics

### Heuristic Metrics (No LLM Required)

| Metric | Description |
|--------|-------------|
| BLEU | N-gram precision for translation |
| ROUGE | Recall-oriented summary evaluation |
| Levenshtein | Edit distance similarity |
| Cosine | Vector similarity |
| Exact Match | String equality |

### LLM Judge Metrics

| Metric | Description |
|--------|-------------|
| Hallucination | Detects unsupported claims |
| Relevance | Query-response relevance |
| Coherence | Logical flow and structure |
| G-EVAL | Customizable LLM evaluation |

---

# Feature Comparison

| Feature | Opik (Python) | go-opik | omniobserve |
|---------|:-------------:|:-------:|:-----------:|
| Traces & Spans | :white_check_mark: | :white_check_mark: | :white_check_mark: |
| Feedback Scores | :white_check_mark: | :white_check_mark: | :white_check_mark: |
| Datasets | :white_check_mark: | :white_check_mark: | :white_check_mark: |
| Experiments | :white_check_mark: | :white_check_mark: | Partial |
| Prompts | :white_check_mark: | :white_check_mark: | :white_check_mark: |
| Distributed Tracing | :white_check_mark: | :white_check_mark: | :x: |
| Streaming | :white_check_mark: | :white_check_mark: | :x: |
| Attachments | :white_check_mark: | :white_check_mark: | :x: |

---

# HTTP Middleware

Automatic tracing for HTTP handlers:

```go
import "github.com/agentplexus/go-opik/middleware"

// Wrap your handler
handler := middleware.Trace(client, myHandler,
    middleware.WithSpanName("api-request"),
    middleware.WithSpanType(opik.SpanTypeTool),
)

http.Handle("/api/", handler)
```

Captures:
- Request method, path, headers
- Response status and duration
- Automatic parent span detection

---

# Installation

```bash
go get github.com/agentplexus/go-opik
```

### Configuration

```go
// From environment variables
// OPIK_API_KEY, OPIK_WORKSPACE, OPIK_URL
client, _ := opik.NewClient()

// Or explicit configuration
client, _ := opik.NewClient(
    opik.WithAPIKey("your-api-key"),
    opik.WithWorkspace("your-workspace"),
    opik.WithProjectName("my-project"),
)
```

---

# Resources

### Links

- **Repository**: github.com/agentplexus/go-opik
- **Opik**: github.com/comet-ml/opik
- **OmniObserve**: github.com/agentplexus/omniobserve
- **Documentation**: agentplexus.github.io/go-opik

### License

MIT License - Open source and free to use

---

<!-- _paginate: false -->

# go-opik

**Full-featured Go SDK for Opik LLM Observability**

```bash
go get github.com/agentplexus/go-opik
```

- Traces, spans, feedback scores
- Datasets, experiments, prompts
- Evaluation metrics (heuristic + LLM)
- OmniObserve integration
- Production ready
