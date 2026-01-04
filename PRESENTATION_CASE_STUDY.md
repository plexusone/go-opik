---
marp: true
theme: vibeminds
paginate: true
style: |
  /* Mermaid diagram styling */
  .mermaid-container {
    display: flex;
    justify-content: center;
    align-items: center;
    width: 100%;
    margin: 0.5em 0;
  }

  .mermaid {
    text-align: center;
  }

  .mermaid svg {
    max-height: 280px;
    width: auto;
  }

  .mermaid .node rect,
  .mermaid .node polygon {
    rx: 5px;
    ry: 5px;
  }

  .mermaid .nodeLabel {
    padding: 0 10px;
  }

  /* Two-column layout */
  .columns {
    display: flex;
    gap: 40px;
    align-items: flex-start;
  }

  .column-left {
    flex: 1;
  }

  .column-right {
    flex: 1;
  }

  .column-left .mermaid svg {
    min-height: 400px;
    height: auto;
    max-height: 500px;
  }

  /* Section divider slides */
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

  section.section-divider p {
    font-size: 1.1em;
    color: #9575cd;
    margin-top: 1em;
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

<!--
Welcome to Building go-opik. <break time="500ms"/>
A Go SDK for LLM Observability. <break time="700ms"/>
This is an AI-Assisted Development Case Study, <break time="400ms"/>
where we built an entire SDK using Claude Opus 4.5 with Claude Code. <break time="800ms"/>
-->

# Building go-opik
## A Go SDK for LLM Observability

**An AI-Assisted Development Case Study**

Using Claude Opus 4.5 with Claude Code

---

<!-- _class: section-divider -->
<!-- _paginate: false -->

<!--
Section 1: Introduction and Overview. <break time="600ms"/>
Let's start by understanding what Opik is, <break time="300ms"/>
and how we approached building this SDK. <break time="800ms"/>
-->

# Section 1
## Introduction & Overview

What is Opik and how we approached the SDK

---

<!--
What is Opik? <break time="500ms"/>
Opik is an open-source LLM observability platform created by Comet ML. <break time="600ms"/>
It provides several key capabilities. <break time="400ms"/>
Traces and Spans for tracking LLM calls and application flow. <break time="500ms"/>
Datasets for storing test data for evaluation. <break time="400ms"/>
Experiments for running and comparing model evaluations. <break time="500ms"/>
Prompts for versioning and managing prompt templates. <break time="400ms"/>
And Feedback for capturing quality scores and user feedback. <break time="600ms"/>
Our goal was to build a comprehensive Go SDK <break time="300ms"/>
that matches the Python SDK's capabilities. <break time="800ms"/>
-->

# What is Opik? 🔭

**Opik** is an open-source LLM observability platform by Comet ML

- **Traces & Spans** - Track LLM calls and application flow
- **Datasets** - Store test data for evaluation
- **Experiments** - Run and compare model evaluations
- **Prompts** - Version and manage prompt templates
- **Feedback** - Capture quality scores and user feedback

**Goal**: Build a comprehensive Go SDK matching the Python SDK's capabilities

---

<!--
Let's look at the project scope. <break time="500ms"/>
The SDK includes seven major components. <break time="400ms"/>
A Core SDK with client, traces, spans, and context propagation. <break time="500ms"/>
Data Management for datasets, experiments, and prompts. <break time="500ms"/>
An Evaluation framework with heuristic metrics and LLM judges. <break time="500ms"/>
Integrations with OpenAI, Anthropic, and GoLLM. <break time="500ms"/>
Infrastructure including CLI, middleware, streaming, and batching. <break time="500ms"/>
A comprehensive Testing suite with mocks. <break time="400ms"/>
And full Documentation with an MkDocs site, tutorials, and README. <break time="600ms"/>
For source analysis, we read over 50 Python SDK files totaling around 20,000 lines of code. <break time="500ms"/>
The OpenAPI specification contains 201 API operations across nearly 15,000 lines. <break time="500ms"/>
In total, we created over 50 Go source files across 12 packages. <break time="800ms"/>
-->

# Project Scope 📋

| Component | Description |
|-----------|-------------|
| **Core SDK** | Client, traces, spans, context propagation |
| **Data Management** | Datasets, experiments, prompts |
| **Evaluation** | Heuristic metrics + LLM judges |
| **Integrations** | OpenAI, Anthropic, GoLLM |
| **Infrastructure** | CLI, middleware, streaming, batching |
| **Testing** | Comprehensive test suite with mocks |
| **Documentation** | MkDocs site, tutorials, README |

**Source Analysis**: 50+ Python files (~20K lines) | **OpenAPI Spec**: 201 operations (~15K lines)

**Output**: ~50+ Go source files (~15K lines) across 12 packages

---

<!--
Here's the architecture overview. <break time="500ms"/>
At the root level, we have the core SDK files for client, trace, span, and configuration. <break time="600ms"/>
The cmd/opik directory contains the CLI tool. <break time="500ms"/>
The evaluation package is split into heuristic metrics like BLEU, ROUGE, and Levenshtein, <break time="500ms"/>
and LLM-as-judge metrics. <break time="500ms"/>
The integrations package provides support for OpenAI, Anthropic, and GoLLM. <break time="500ms"/>
We also have HTTP tracing middleware, <break time="400ms"/>
an internal API client generated by ogen, <break time="400ms"/>
and test utilities including a mock server and matchers. <break time="800ms"/>
-->

# Architecture Overview 🏗️

```
go-opik/
├── *.go                    # Core SDK (client, trace, span, config, etc.)
├── cmd/opik/               # CLI tool
├── evaluation/
│   ├── heuristic/          # BLEU, ROUGE, Levenshtein, etc.
│   └── llm/                # LLM-as-judge metrics
├── integrations/
│   ├── openai/             # OpenAI provider + tracing
│   ├── anthropic/          # Anthropic provider + tracing
│   └── gollm/              # GoLLM adapter
├── middleware/             # HTTP tracing middleware
├── internal/api/           # ogen-generated API client
└── testutil/               # Mock server + matchers
```

---

<!--
Let me walk you through the key design decisions we made. <break time="600ms"/>
First, we chose ogen for API client generation. <break time="500ms"/>
ogen provides type-safe code with no reflection, <break time="400ms"/>
and correctly handles optional and nullable fields. <break time="500ms"/>
We generated the client from an OpenAPI spec with over 14,000 lines. <break time="600ms"/>
Second, we used the Functional Options pattern for configuration. <break time="500ms"/>
This is idiomatic Go and allows for clean, readable client initialization. <break time="600ms"/>
Third, we implemented context-based tracing, <break time="400ms"/>
using Go's standard context.Context for automatic parent-child span relationships. <break time="800ms"/>
-->

# Key Design Decisions 🎯

### 1. **ogen for API Client Generation**
- Type-safe, no reflection
- Handles optional/nullable fields correctly
- Generated from OpenAPI spec (14,846 lines)

### 2. **Functional Options Pattern**
```go
client, err := opik.NewClient(
    opik.WithAPIKey("key"),
    opik.WithWorkspace("workspace"),
    opik.WithProjectName("my-project"),
)
```

### 3. **Context-Based Tracing**
- Idiomatic Go using `context.Context`
- Automatic parent-child span relationships

---

<!-- _class: section-divider -->
<!-- _paginate: false -->

<!--
Section 2: Implementation Deep Dive. <break time="600ms"/>
Now let's explore the features, testing approach, documentation, and DevOps setup. <break time="800ms"/>
-->

# Section 2
## Implementation Deep Dive

Features, Testing, Documentation & DevOps

---

<!--
Here are the core features we implemented. <break time="500ms"/>
On the tracing side, we built support for traces and nested spans, <break time="400ms"/>
multiple span types including LLM, Tool, and General, <break time="500ms"/>
distributed trace propagation, streaming support, and automatic batching. <break time="600ms"/>
For data management, we implemented dataset CRUD operations with items, <break time="400ms"/>
experiment tracking, prompt versioning, and template rendering. <break time="600ms"/>
The evaluation framework includes over 10 heuristic metrics, <break time="400ms"/>
8 or more LLM judge metrics, a G-EVAL implementation, <break time="400ms"/>
custom judge support, and a concurrent evaluation engine. <break time="600ms"/>
For integrations, we built OpenAI and Anthropic tracing transports, <break time="400ms"/>
a GoLLM adapter for ADK, and HTTP middleware. <break time="800ms"/>
-->

# Core Features Implemented ✨

<div style="display: grid; grid-template-columns: 1fr 1fr; gap: 2rem;">
<div>

**Tracing**
- Traces and nested spans
- LLM, Tool, General span types
- Distributed trace propagation
- Streaming support
- Automatic batching

**Data**
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
- Concurrent evaluation engine

**Integrations**
- OpenAI tracing transport
- Anthropic tracing transport
- GoLLM adapter for ADK
- HTTP middleware

</div>
</div>

---

<!--
Our testing strategy mirrored the Python SDK's patterns. <break time="500ms"/>
We built a test utilities package with a Matcher interface, <break time="400ms"/>
providing flexible matchers like Any, AnyButNil, AnyString, AnyMap, and AnyFloat. <break time="600ms"/>
We also built a MockServer, similar to Python's respx library. <break time="600ms"/>
Looking at test coverage, the Core SDK has 15 test files covering client, trace, span, config, and context. <break time="600ms"/>
The Evaluation package has 6 test files for all metrics, scoring, and the engine. <break time="500ms"/>
Integrations have 6 test files for providers and tracing transports. <break time="500ms"/>
And Test Utils has 2 test files for matchers and the mock server. <break time="800ms"/>
-->

# Testing Strategy 🧪

### Mirrored Python SDK's Testing Patterns

**Test Utilities** (`testutil/`)
- `Matcher` interface: `Any()`, `AnyButNil()`, `AnyString()`, `AnyMap()`, `AnyFloat()`
- `MockServer` - HTTP mock server similar to Python's `respx`

**Test Coverage**
| Package | Test Files | Key Tests |
|---------|------------|-----------|
| Core SDK | 15 files | Client, trace, span, config, context |
| Evaluation | 6 files | All metrics, scoring, engine |
| Integrations | 6 files | Providers, tracing transports |
| Test Utils | 2 files | Matchers, mock server |

---

<!--
Here's an example of how the mock server works. <break time="500ms"/>
You create a new MockServer and defer its close. <break time="500ms"/>
Then you set up a route using OnPost with a path, <break time="400ms"/>
and define the JSON response with RespondJSON. <break time="600ms"/>
You can then create your provider pointing to the mock server URL, <break time="400ms"/>
make the request, and verify the response. <break time="500ms"/>
Finally, you can assert that the correct number of requests were made. <break time="500ms"/>
This pattern makes testing HTTP integrations clean and reliable. <break time="800ms"/>
-->

# Test Example: Mock Server 💻

```go
func TestProviderComplete(t *testing.T) {
    ms := testutil.NewMockServer()
    defer ms.Close()

    ms.OnPost("/v1/chat/completions").RespondJSON(200, map[string]any{
        "choices": []map[string]any{{
            "message": map[string]any{"content": "Hello!"},
        }},
    })

    provider := openai.NewProvider(openai.WithBaseURL(ms.URL()))
    resp, err := provider.Complete(ctx, request)

    // Verify response
    assert.Equal(t, "Hello!", resp.Content)

    // Verify request was made correctly
    assert.Equal(t, 1, ms.RequestCount())
}
```

---

<!--
We created comprehensive documentation for the SDK. <break time="500ms"/>
The MkDocs site includes Getting Started guides for installation, configuration, and testing. <break time="500ms"/>
Core Concepts covering traces, spans, context, and feedback. <break time="500ms"/>
Feature documentation for datasets, experiments, prompts, and streaming. <break time="500ms"/>
Evaluation guides for heuristic metrics and LLM judges. <break time="500ms"/>
Integration guides for OpenAI, Anthropic, GoLLM, and middleware. <break time="500ms"/>
And tutorials including agentic observability with Google ADK and Eino. <break time="600ms"/>
The README provides comprehensive feature documentation with code examples, <break time="400ms"/>
CI badges, license information, and related links. <break time="800ms"/>
-->

# Documentation Created 📚

### MkDocs Site Structure
- **Getting Started**: Installation, configuration, testing
- **Core Concepts**: Traces, spans, context, feedback
- **Features**: Datasets, experiments, prompts, streaming
- **Evaluation**: Heuristic metrics, LLM judges
- **Integrations**: OpenAI, Anthropic, GoLLM, middleware
- **Tutorials**: Agentic observability with Google ADK & Eino

### README.md
- Comprehensive feature documentation
- Code examples for all major features
- CI badges, license, related links

---

<!--
We also created a tutorial on Agentic Observability. <break time="500ms"/>
This is a case study on integrating Opik with multi-agent systems, <break time="400ms"/>
based on the real-world stats-agent-team project. <break time="600ms"/>
The diagram shows the agent workflow. <break time="400ms"/>
The Orchestrator, built with Eino, coordinates the pipeline. <break time="500ms"/>
It sends tasks to the Research agent for searching. <break time="400ms"/>
Results flow to the Synthesis agent, which uses LLM with ADK. <break time="500ms"/>
Then to the Verification agent, also using LLM with ADK, <break time="400ms"/>
which reports back to the Orchestrator. <break time="600ms"/>
The tutorial covers Eino workflow graph tracing, Google ADK agent tracing, <break time="400ms"/>
tool invocation spans, and quality feedback scores. <break time="800ms"/>
-->

# Tutorial: Agentic Observability 🤖

**Case study**: Integrating Opik with multi-agent systems

Based on real-world `stats-agent-team` project:

<div class="mermaid">
flowchart LR
    D["🎯 Orchestrator<br/>(Eino)"] --> A["🔍 Research<br/>(Search)"]
    A --> B["🧠 Synthesis<br/>(LLM + ADK)"]
    B --> C["✅ Verification<br/>(LLM + ADK)"]
    C --> D
    style A fill:#667eea,stroke:#764ba2,color:#fff
    style B fill:#667eea,stroke:#764ba2,color:#fff
    style C fill:#667eea,stroke:#764ba2,color:#fff
    style D fill:#764ba2,stroke:#667eea,color:#fff
</div>

**Covers**:
- Eino workflow graph tracing
- Google ADK agent tracing
- Tool invocation spans
- Quality feedback scores

---

<!--
For CI/CD, we set up a GitHub Actions workflow. <break time="500ms"/>
The test job runs across Go versions 1.21, 1.22, and 1.23, <break time="400ms"/>
executing tests with race detection and coverage reporting. <break time="600ms"/>
The lint job uses golangci-lint with over 25 linters enabled. <break time="500ms"/>
And the build-cli job cross-compiles the CLI for Linux, Darwin, and Windows, <break time="400ms"/>
on both AMD64 and ARM64 architectures. <break time="500ms"/>
This ensures the SDK is thoroughly tested and portable. <break time="800ms"/>
-->

# CI/CD Infrastructure ⚙️

### GitHub Actions Workflow

```yaml
jobs:
  test:
    strategy:
      matrix:
        go-version: ['1.21', '1.22', '1.23']
    steps:
      - run: go test -v -race -coverprofile=coverage.out ./...

  lint:
    steps:
      - uses: golangci/golangci-lint-action@v6

  build-cli:
    strategy:
      matrix:
        goos: [linux, darwin, windows]
        goarch: [amd64, arm64]
```

**golangci-lint**: 25+ linters enabled

---

<!-- _class: section-divider -->
<!-- _paginate: false -->

<!--
Section 3: AI-Assisted Development. <break time="600ms"/>
This is where it gets interesting. <break time="400ms"/>
Let's look at Claude Opus 4.5's performance, <break time="300ms"/>
and the insights and lessons we learned. <break time="800ms"/>
-->

# Section 3
## AI-Assisted Development

Claude Opus 4.5 performance, insights & lessons learned

---

<!--
Let's look at the Claude Opus 4.5 developer experience. <break time="500ms"/>
For the session configuration, we used Claude Opus 4.5 model, <break time="400ms"/>
with the High effort setting for deeper analysis. <break time="500ms"/>
We used Extended context with summarization to handle the large codebase, <break time="500ms"/>
and had access to the full Claude Code toolset. <break time="600ms"/>
Our development approach was iterative, <break time="400ms"/>
implementing features with immediate testing. <break time="400ms"/>
We leveraged parallel file reads and tool calls for efficiency, <break time="500ms"/>
and used todo tracking for complex multi-step tasks. <break time="800ms"/>
-->

# Claude Opus 4.5 DevEx 🧠

### Session Configuration

| Setting | Value |
|---------|-------|
| **Model** | Claude Opus 4.5 (`claude-opus-4-5-20250514`) |
| **Effort** | High |
| **Context** | Extended (with summarization) |
| **Tools** | Full Claude Code toolset |

### Development Approach
- Iterative implementation with immediate testing
- Parallel file reads and tool calls for efficiency
- Todo tracking for complex multi-step tasks

---

<!--
Here are the session statistics. <break time="500ms"/>
Looking at the development timeline, we started around 7 PM. <break time="400ms"/>
The core SDK was complete by about 8:15 PM. <break time="400ms"/>
The test suite was finished by 10 PM. <break time="400ms"/>
And documentation and CI were complete by midnight. <break time="500ms"/>
Total time was approximately 4 to 5 hours. <break time="600ms"/>
For token usage, we estimate 800,000 to 1 million input tokens, <break time="400ms"/>
and 150,000 to 200,000 output tokens. <break time="500ms"/>
We read and analyzed over 50 Python SDK files, <break time="400ms"/>
totaling more than 20,000 lines of code as reference. <break time="500ms"/>
Over 60 Go files were created or modified, <break time="400ms"/>
totaling more than 15,000 lines of code written. <break time="500ms"/>
The estimated cost was just 20 to 30 dollars. <break time="800ms"/>
-->

# Session Statistics 📊

<div style="display: grid; grid-template-columns: 1fr 1fr; gap: 2rem;">
<div>
<center>

### Development Timeline

</center>

| Milestone | Time |
|-----------|-----------|
| **Session Start** | ~19:00 |
| **Core SDK Complete** | ~20:15 |
| **Test Suite Complete +** | ~22:00 |
| **Docs & CI Complete +** | ~24:00 |
| **Total Time +** | **~4-5 hours** |

&nbsp;&nbsp;+ includes human multi-tasking

</div><div>
<center>

### Token Usage (Estimated)

</center>

| Category | Estimate |
|----------|----------|
| **Input Tokens** | ~800K - 1M |
| **Output Tokens** | ~150K - 200K |
| **Files Read (Python)** | 50+ |
| **Lines of Code Read** | ~20,000+ |
| **Files Created/Modified** | 60+ |
| **Lines of Code Written** | ~15,000+ |
| **Estimated Cost** | ~$20 - $30 |
</div></div>

---

<!--
How does this compare to industry benchmarks? <break time="500ms"/>
SDK generation companies report consistent timelines. <break time="400ms"/>
APIMatic says 4 weeks to build a single SDK, at around 52K dollars including maintenance. <break time="500ms"/>
Speakeasy estimates 90K per hand-written SDK. <break time="500ms"/>
liblab reports weeks or months for manual development, at 50K plus per language. <break time="600ms"/>
These companies sell SDK tools, so they emphasize manual costs, <break time="400ms"/>
but the ballpark is consistent across sources. <break time="500ms"/>
Building a production SDK manually takes weeks, not hours. <break time="800ms"/>
-->

# Industry Benchmarks ⏱️

<center>

### What SDK Companies Report for Manual Development

</center>

| Source | Time Estimate | Cost Estimate |
|--------|---------------|---------------|
| [APIMatic](https://www.apimatic.io/blog/2021/09/the-great-sdk-battle-build-vs-buy) | 4 weeks per SDK | ~$52K including maintenance |
| [Speakeasy](https://www.speakeasy.com/blog/how-to-build-sdks) | Months | ~$90K per SDK |
| [liblab](https://liblab.com/blog/automated-sdk-generation) | "Weeks or months" | $50K+ per language |

<center>

*These companies sell SDK tools, so they emphasize manual costs—but the ballpark is consistent.*

**Building a production SDK manually takes weeks, not hours.**

</center>

---

<!--
Now let's compare productivity. <break time="500ms"/>
Industry sources suggest 4 or more weeks for manual SDK development. <break time="500ms"/>
Claude Opus 4.5 completed this work in just 4 to 5 hours. <break time="600ms"/>
What accounts for this difference? <break time="500ms"/>
First, parallel processing. Claude reads multiple files simultaneously. <break time="500ms"/>
Second, no context switching. Continuous focus on a single project. <break time="500ms"/>
Third, pattern recognition. Claude instantly applies Go idioms. <break time="500ms"/>
Fourth, reference implementation. Quickly mapping Python to Go patterns. <break time="500ms"/>
Fifth, no typing delay. Code is generated at output speed. <break time="500ms"/>
And sixth, integrated testing. Tests are written alongside implementation. <break time="600ms"/>
Important caveats: Human review is still essential. <break time="400ms"/>
And your mileage may vary based on API complexity and coverage requirements. <break time="800ms"/>
-->

# Productivity Comparison 🚀

### Time Comparison

| Approach | Time | Source |
|----------|------|--------|
| **Industry Benchmark** | 4+ weeks per SDK | APIMatic, Speakeasy, liblab |
| **Claude Opus 4.5** | 4-5 hours | This project |

1. What Accounts for the Difference?
    1. **Parallel Processing** - Claude reads multiple files simultaneously
    2. **No Context Switching** - Continuous focus on single project
    3. **Pattern Recognition** - Instantly applies Go idioms
    4. **Reference Implementation** - Quickly maps Python → Go patterns
    5. **No Typing Delay** - Generates code at output speed
    6. **Integrated Testing** - Writes tests alongside implementation
2. Caveats
    1. Human review still essential for production deployment
    2. Your mileage may vary based on API complexity and coverage requirements

---

<!--
What did Claude Opus 4.5 handle particularly well? <break time="500ms"/>
First, large codebase navigation. <break time="400ms"/>
Reading and understanding over 50 files, <break time="300ms"/>
and cross-referencing Python SDK patterns. <break time="500ms"/>
Second, code generation quality. <break time="400ms"/>
The code is idiomatic Go with proper error handling, <break time="400ms"/>
using the functional options pattern consistently. <break time="500ms"/>
Third, test development. <break time="400ms"/>
Table-driven tests, mock server implementation, and edge case coverage. <break time="500ms"/>
Fourth, documentation. <break time="400ms"/>
Creating the MkDocs structure, tutorials with code examples, <break time="400ms"/>
and a README with badges. <break time="800ms"/>
-->

# What Claude Handled Well 💪

<center>

### Strengths Demonstrated

</center>

<div style="display: grid; grid-template-columns: 1fr 1fr; gap: 2rem;">
<div>

1. **Large Codebase Navigation**
   - Reading and understanding 50+ files
   - Cross-referencing Python SDK patterns
2. **Code Generation Quality**
   - Idiomatic Go code
   - Proper error handling
   - Functional options pattern

</div><div>

3. **Test Development**
   - Table-driven tests
   - Mock server implementation
   - Edge case coverage
4. **Documentation**
   - MkDocs structure
   - Tutorial with code examples
   - README with badges
</div>

---

<!--
Of course, there were challenges along the way. <break time="500ms"/>
Challenge 1: The ogen generated code. <break time="400ms"/>
Complex optional types in API responses required careful handling <break time="400ms"/>
of OptXxx types with .Set checks. <break time="600ms"/>
Challenge 2: golangci-lint version 2 configuration. <break time="400ms"/>
The config format changed significantly, <break time="400ms"/>
requiring iterative fixes with immediate validation. <break time="600ms"/>
Challenge 3: Matching Python SDK patterns. <break time="400ms"/>
Different language idioms meant adapting patterns, <break time="400ms"/>
like using context instead of decorators. <break time="600ms"/>
Challenge 4: Test utilities. <break time="400ms"/>
We needed something like Python's respx for mocking, <break time="400ms"/>
so we built a custom MockServer with route matching. <break time="800ms"/>
-->

# Challenges & Solutions 🔧

1. ### Challenge 1: ogen Generated Code
    - **Issue**: Complex optional types in API responses
    - **Solution**: Careful handling of `OptXxx` types with `.Set` checks
2. ### Challenge 2: golangci-lint v2 Config
    - **Issue**: Config format changed significantly
    - **Solution**: Iterative fixes with immediate validation
3. ### Challenge 3: Matching Python SDK Patterns
    - **Issue**: Different language idioms
    - **Solution**: Adapted patterns (e.g., context vs decorators)
4. ### Challenge 4: Test Utilities
    - **Issue**: Need Python's `respx`-like mocking
    - **Solution**: Built custom `MockServer` with route matching

---

<!--
Let's look at the code quality results. <break time="500ms"/>
Running golangci-lint produced only 2 warnings. <break time="500ms"/>
These were expected duplication warnings in similar API patterns <break time="400ms"/>
between dataset.go and prompt.go. <break time="600ms"/>
For testing, running go test showed all tests passing <break time="400ms"/>
across 11 packages. <break time="500ms"/>
The core SDK, evaluation packages, and integrations all pass. <break time="500ms"/>
This demonstrates the code quality achieved with AI assistance. <break time="800ms"/>
-->

# Code Quality Results ✅

### golangci-lint Output
```
$ golangci-lint run
dataset.go:174: duplicate of prompt.go:279-314 (dupl)
prompt.go:279: duplicate of dataset.go:174-209 (dupl)
```

Only **2 warnings** - expected duplication in similar API patterns

### Test Results
```
$ go test ./...
ok  github.com/agentplexus/go-opik          0.070s
ok  github.com/agentplexus/go-opik/evaluation
ok  github.com/agentplexus/go-opik/evaluation/heuristic
ok  github.com/agentplexus/go-opik/evaluation/llm
ok  github.com/agentplexus/go-opik/integrations/...
```

**All tests passing** across 11 packages

---

<!--
Let's summarize the key takeaways for AI-assisted development. <break time="500ms"/>
First, the high effort setting enables deeper code analysis and better solutions. <break time="500ms"/>
Second, parallel tool calls significantly speed up exploration and implementation. <break time="500ms"/>
Third, todo tracking helps maintain focus on complex multi-file changes. <break time="500ms"/>
Fourth, iterative validation by running tests after each change catches issues early. <break time="500ms"/>
And fifth, reference implementations like the Python SDK provide valuable patterns. <break time="600ms"/>
The result? <break time="400ms"/>
A production-ready Go SDK built efficiently with AI assistance. <break time="800ms"/>
-->

# Key Takeaways 💡

### AI-Assisted Development Insights

1. **High effort setting** enables deeper code analysis and better solutions
2. **Parallel tool calls** significantly speed up exploration and implementation
3. **Todo tracking** helps maintain focus on complex multi-file changes
4. **Iterative validation** (run tests after each change) catches issues early
5. **Reference implementations** (Python SDK) provide valuable patterns

### Result
A production-ready Go SDK built efficiently with AI assistance

---

<!-- _class: section-divider -->
<!-- _paginate: false -->

<!--
Section 4: Conclusion. <break time="600ms"/>
Let's wrap up with the deliverables, future work, and resources. <break time="800ms"/>
-->

# Section 4
## Conclusion

Deliverables, future work & resources

---

<!--
Here's a summary of the project deliverables. <break time="500ms"/>
Core SDK: Complete. <break time="300ms"/>
CLI Tool: Complete. <break time="300ms"/>
Evaluation Framework: Complete. <break time="300ms"/>
LLM Integrations: Complete. <break time="300ms"/>
Test Suite: Complete. <break time="300ms"/>
MkDocs Documentation: Complete. <break time="300ms"/>
CI/CD Pipeline: Complete. <break time="300ms"/>
Agentic Tutorial: Complete. <break time="500ms"/>
All deliverables are available in the repository at github.com/agentplexus/go-opik. <break time="800ms"/>
-->

# Project Deliverables 📦

| Deliverable | Status |
|-------------|--------|
| Core SDK | ✅ Complete |
| CLI Tool | ✅ Complete |
| Evaluation Framework | ✅ Complete |
| LLM Integrations | ✅ Complete |
| Test Suite | ✅ Complete |
| MkDocs Documentation | ✅ Complete |
| CI/CD Pipeline | ✅ Complete |
| Agentic Tutorial | ✅ Complete |

**Repository**: `github.com/agentplexus/go-opik`

---

<!--
What about future enhancements? <break time="500ms"/>
There are several potential additions we're considering. <break time="400ms"/>
More LLM providers like Gemini, Mistral, and Cohere. <break time="500ms"/>
gRPC support for high-performance tracing. <break time="500ms"/>
An OpenTelemetry bridge to export to OTel collectors. <break time="500ms"/>
More tutorials covering RAG applications and chatbots. <break time="500ms"/>
And a benchmarks suite for performance testing. <break time="600ms"/>
The project is open for contributions. <break time="400ms"/>
Issues and pull requests are welcome. <break time="400ms"/>
The SDK is released under the MIT License. <break time="800ms"/>
-->

# Future Enhancements 🔮

### Potential Additions

- **More LLM Providers**: Gemini, Mistral, Cohere
- **gRPC Support**: For high-performance tracing
- **OpenTelemetry Bridge**: Export to OTel collectors
- **More Tutorials**: RAG applications, chatbots
- **Benchmarks**: Performance testing suite

### Community

- Open for contributions
- Issues and PRs welcome
- MIT License

---

<!--
Here are the important links. <break time="500ms"/>
The repository is at github.com/agentplexus/go-opik. <break time="500ms"/>
The official Opik project is at github.com/comet-ml/opik. <break time="500ms"/>
And documentation is available at agentplexus.github.io/go-opik. <break time="600ms"/>
You can find me on GitHub at @agentplexus. <break time="800ms"/>
-->

# Resources 🔗

### Links

- **Repository**: github.com/agentplexus/go-opik
- **Opik**: github.com/comet-ml/opik
- **Documentation**: agentplexus.github.io/go-opik

### Contact

- GitHub: @agentplexus

---

<!--
Thank you for joining this presentation. <break time="500ms"/>
go-opik: A Go SDK for LLM Observability. <break time="600ms"/>
Built with Claude Opus 4.5 and Claude Code. <break time="800ms"/>
Thanks for watching! <break time="800ms"/>
-->

# Thank You 🙏

## go-opik

**A Go SDK for LLM Observability**

Built with Claude Opus 4.5 + Claude Code
