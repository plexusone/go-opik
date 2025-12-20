# Agentic Observability: Integrating Opik with Google ADK and Eino

This tutorial demonstrates how to add LLM observability to agentic Go applications using the go-comet-ml-opik SDK. We'll use a real-world case study based on the [stats-agent-team](https://github.com/grokify/stats-agent-team) project, which implements a multi-agent system for statistics research and verification.

## Overview

Modern AI applications increasingly use agentic architectures where multiple specialized agents collaborate to complete complex tasks. Observability is critical for:

- **Debugging**: Understanding why an agent made a particular decision
- **Performance**: Identifying bottlenecks in multi-agent workflows
- **Cost tracking**: Monitoring LLM token usage across agents
- **Quality assurance**: Evaluating agent outputs over time

This tutorial covers integration with two popular Go agent frameworks:

1. **Google Agent Development Kit (ADK)** - A framework for building LLM-powered agents with tools
2. **Eino** - CloudWeGo's framework for building deterministic agent workflows as graphs

## Case Study: Stats Agent Team

The stats-agent-team project implements a 4-agent system for researching and verifying statistics:

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  Research Agent │ ──▶ │ Synthesis Agent │ ──▶ │Verification Agent│ ──▶ │  Orchestrator   │
│  (Web Search)   │     │   (LLM + ADK)   │     │    (LLM + ADK)   │     │  (Eino Graph)   │
└─────────────────┘     └─────────────────┘     └─────────────────┘     └─────────────────┘
```

- **Research Agent**: Finds relevant sources using web search APIs
- **Synthesis Agent**: Extracts statistics from web pages using LLM (Google ADK)
- **Verification Agent**: Verifies extracted statistics against sources (Google ADK)
- **Orchestrator**: Coordinates the workflow using Eino's graph-based approach

## Part 1: Integrating with Google ADK Agents

Google ADK provides a structured way to build agents with tools. Here's how to add Opik observability.

### 1.1 Basic ADK Agent Structure

The Synthesis Agent uses ADK's `llmagent` and `functiontool` packages:

```go
import (
    "google.golang.org/adk/agent"
    "google.golang.org/adk/agent/llmagent"
    "google.golang.org/adk/model"
    "google.golang.org/adk/tool"
    "google.golang.org/adk/tool/functiontool"
)

type SynthesisAgent struct {
    model    model.LLM
    adkAgent agent.Agent
}

func NewSynthesisAgent(llmModel model.LLM) (*SynthesisAgent, error) {
    sa := &SynthesisAgent{model: llmModel}

    // Create a tool for statistics extraction
    synthesisTool, err := functiontool.New(functiontool.Config{
        Name:        "synthesize_statistics",
        Description: "Extracts numerical statistics from web page content",
    }, sa.synthesisToolHandler)
    if err != nil {
        return nil, err
    }

    // Create the ADK agent
    adkAgent, err := llmagent.New(llmagent.Config{
        Name:        "statistics_synthesis_agent",
        Model:       llmModel,
        Description: "Extracts statistics from web content",
        Instruction: "You are a statistics extraction expert...",
        Tools:       []tool.Tool{synthesisTool},
    })
    if err != nil {
        return nil, err
    }

    sa.adkAgent = adkAgent
    return sa, nil
}
```

### 1.2 Adding Opik Tracing to ADK Agents

Wrap your ADK agent's LLM calls with Opik traces:

```go
import (
    opik "github.com/grokify/go-comet-ml-opik"
)

type TracedSynthesisAgent struct {
    *SynthesisAgent
    opikClient *opik.Client
}

func NewTracedSynthesisAgent(llmModel model.LLM) (*TracedSynthesisAgent, error) {
    // Create base agent
    base, err := NewSynthesisAgent(llmModel)
    if err != nil {
        return nil, err
    }

    // Create Opik client
    opikClient, err := opik.NewClient(
        opik.WithProjectName("stats-agent-team"),
    )
    if err != nil {
        return nil, err
    }

    return &TracedSynthesisAgent{
        SynthesisAgent: base,
        opikClient:     opikClient,
    }, nil
}

func (tsa *TracedSynthesisAgent) ExtractStatistics(ctx context.Context, topic string, content string) ([]Statistic, error) {
    // Create a trace for this extraction operation
    trace, err := tsa.opikClient.Trace(ctx, "extract-statistics",
        opik.WithTraceInput(map[string]any{
            "topic":          topic,
            "content_length": len(content),
        }),
        opik.WithTraceTags("agent:synthesis", "operation:extraction"),
    )
    if err != nil {
        return nil, err
    }
    defer trace.End(ctx)

    // Create a span for the LLM call
    span, err := trace.Span(ctx, "llm-extraction",
        opik.WithSpanType(opik.SpanTypeLLM),
        opik.WithSpanModel("gemini-2.0-flash-exp"),
        opik.WithSpanProvider("google"),
        opik.WithSpanInput(map[string]any{
            "prompt_template": "statistics_extraction",
            "topic":           topic,
        }),
    )
    if err != nil {
        return nil, err
    }

    // Make the actual LLM call
    stats, err := tsa.doExtraction(ctx, topic, content)

    // End span with output and token usage
    span.End(ctx,
        opik.WithSpanOutput(map[string]any{
            "statistics_count": len(stats),
            "statistics":       stats,
        }),
        opik.WithSpanUsage(opik.Usage{
            PromptTokens:     estimateTokens(content),
            CompletionTokens: estimateTokens(formatStats(stats)),
        }),
    )

    if err != nil {
        // Add error feedback
        span.AddFeedbackScore(ctx, "error", 0, err.Error())
        return nil, err
    }

    // Add quality feedback
    span.AddFeedbackScore(ctx, "extraction_quality",
        float64(len(stats))/10.0, // Normalize to 0-1
        fmt.Sprintf("Extracted %d statistics", len(stats)),
    )

    // Update trace output
    trace.End(ctx, opik.WithTraceOutput(map[string]any{
        "total_statistics": len(stats),
        "success":          true,
    }))

    return stats, nil
}
```

### 1.3 Tracing Tool Invocations

When ADK agents invoke tools, trace each tool call as a separate span:

```go
func (tsa *TracedSynthesisAgent) synthesisToolHandler(ctx tool.Context, input SynthesisInput) (SynthesisOutput, error) {
    // Get parent trace from context (set by orchestrator)
    parentTrace := opik.TraceFromContext(ctx)

    // Create a span for this tool invocation
    span, err := parentTrace.Span(ctx, "tool:synthesize_statistics",
        opik.WithSpanType(opik.SpanTypeTool),
        opik.WithSpanInput(map[string]any{
            "topic":         input.Topic,
            "url_count":     len(input.SearchResults),
            "min_stats":     input.MinStatistics,
            "max_stats":     input.MaxStatistics,
        }),
    )
    if err != nil {
        return SynthesisOutput{}, err
    }
    defer span.End(ctx)

    // Process each URL with its own span
    var candidates []CandidateStatistic
    for i, result := range input.SearchResults {
        urlSpan, _ := span.Span(ctx, fmt.Sprintf("process-url-%d", i),
            opik.WithSpanType(opik.SpanTypeGeneral),
            opik.WithSpanInput(map[string]any{
                "url":    result.URL,
                "domain": result.Domain,
            }),
        )

        stats, err := tsa.processURL(ctx, input.Topic, result)

        urlSpan.End(ctx, opik.WithSpanOutput(map[string]any{
            "stats_extracted": len(stats),
            "error":           err != nil,
        }))

        if err == nil {
            candidates = append(candidates, stats...)
        }
    }

    span.End(ctx, opik.WithSpanOutput(map[string]any{
        "total_candidates": len(candidates),
    }))

    return SynthesisOutput{Candidates: candidates}, nil
}
```

## Part 2: Integrating with Eino Workflow Graphs

Eino provides a graph-based approach to building agent workflows. Each node in the graph can be traced.

### 2.1 Basic Eino Graph Structure

The Orchestrator uses Eino's `compose` package to build a workflow:

```go
import (
    "github.com/cloudwego/eino/compose"
)

type EinoOrchestrator struct {
    graph *compose.Graph[*OrchestrationRequest, *OrchestrationResponse]
}

func NewEinoOrchestrator() *EinoOrchestrator {
    eo := &EinoOrchestrator{}
    eo.graph = eo.buildWorkflowGraph()
    return eo
}

func (eo *EinoOrchestrator) buildWorkflowGraph() *compose.Graph[*OrchestrationRequest, *OrchestrationResponse] {
    g := compose.NewGraph[*OrchestrationRequest, *OrchestrationResponse]()

    // Define nodes
    const (
        nodeValidate     = "validate"
        nodeResearch     = "research"
        nodeSynthesis    = "synthesis"
        nodeVerification = "verification"
        nodeFormat       = "format"
    )

    // Add lambda nodes
    g.AddLambdaNode(nodeValidate, compose.InvokableLambda(eo.validateInput))
    g.AddLambdaNode(nodeResearch, compose.InvokableLambda(eo.callResearch))
    g.AddLambdaNode(nodeSynthesis, compose.InvokableLambda(eo.callSynthesis))
    g.AddLambdaNode(nodeVerification, compose.InvokableLambda(eo.callVerification))
    g.AddLambdaNode(nodeFormat, compose.InvokableLambda(eo.formatResponse))

    // Define edges
    g.AddEdge(compose.START, nodeValidate)
    g.AddEdge(nodeValidate, nodeResearch)
    g.AddEdge(nodeResearch, nodeSynthesis)
    g.AddEdge(nodeSynthesis, nodeVerification)
    g.AddEdge(nodeVerification, nodeFormat)
    g.AddEdge(nodeFormat, compose.END)

    return g
}
```

### 2.2 Adding Opik Tracing to Eino Workflows

Trace the entire workflow and each node:

```go
type TracedEinoOrchestrator struct {
    *EinoOrchestrator
    opikClient *opik.Client
}

func NewTracedEinoOrchestrator() (*TracedEinoOrchestrator, error) {
    opikClient, err := opik.NewClient(
        opik.WithProjectName("stats-agent-team"),
    )
    if err != nil {
        return nil, err
    }

    return &TracedEinoOrchestrator{
        EinoOrchestrator: NewEinoOrchestrator(),
        opikClient:       opikClient,
    }, nil
}

func (teo *TracedEinoOrchestrator) Orchestrate(ctx context.Context, req *OrchestrationRequest) (*OrchestrationResponse, error) {
    // Create a trace for the entire workflow
    trace, err := teo.opikClient.Trace(ctx, "eino-orchestration",
        opik.WithTraceInput(map[string]any{
            "topic":              req.Topic,
            "min_verified_stats": req.MinVerifiedStats,
            "max_candidates":     req.MaxCandidates,
        }),
        opik.WithTraceTags("framework:eino", "workflow:orchestration"),
    )
    if err != nil {
        return nil, err
    }

    // Store trace in context for child spans
    ctx = opik.ContextWithTrace(ctx, trace)

    // Compile and execute the graph
    compiled, err := teo.graph.Compile(ctx)
    if err != nil {
        trace.End(ctx, opik.WithTraceOutput(map[string]any{
            "error": err.Error(),
        }))
        return nil, err
    }

    result, err := compiled.Invoke(ctx, req)

    // End trace with final results
    if err != nil {
        trace.End(ctx, opik.WithTraceOutput(map[string]any{
            "error":   err.Error(),
            "success": false,
        }))
        return nil, err
    }

    trace.End(ctx, opik.WithTraceOutput(map[string]any{
        "success":          true,
        "verified_count":   result.VerifiedCount,
        "total_candidates": result.TotalCandidates,
        "failed_count":     result.FailedCount,
    }))

    // Add workflow quality score
    qualityScore := float64(result.VerifiedCount) / float64(req.MinVerifiedStats)
    if qualityScore > 1.0 {
        qualityScore = 1.0
    }
    trace.AddFeedbackScore(ctx, "workflow_quality", qualityScore,
        fmt.Sprintf("Verified %d/%d statistics", result.VerifiedCount, req.MinVerifiedStats))

    return result, nil
}
```

### 2.3 Tracing Individual Eino Nodes

Create spans for each node in the workflow:

```go
func (teo *TracedEinoOrchestrator) callResearch(ctx context.Context, req *OrchestrationRequest) (*ResearchState, error) {
    // Get trace from context
    trace := opik.TraceFromContext(ctx)

    // Create span for this node
    span, _ := trace.Span(ctx, "eino-node:research",
        opik.WithSpanType(opik.SpanTypeGeneral),
        opik.WithSpanInput(map[string]any{
            "topic":       req.Topic,
            "max_results": req.MaxCandidates,
        }),
        opik.WithSpanMetadata(map[string]any{
            "eino_node":    "research",
            "agent_url":    teo.researchAgentURL,
        }),
    )

    startTime := time.Now()

    // Call the research agent
    resp, err := teo.callResearchAgent(ctx, &ResearchRequest{
        Topic:         req.Topic,
        MaxStatistics: req.MaxCandidates,
    })

    duration := time.Since(startTime)

    if err != nil {
        span.End(ctx, opik.WithSpanOutput(map[string]any{
            "error":    err.Error(),
            "duration": duration.String(),
        }))
        return nil, err
    }

    span.End(ctx, opik.WithSpanOutput(map[string]any{
        "sources_found": len(resp.Candidates),
        "duration":      duration.String(),
    }))

    return &ResearchState{
        Request:       req,
        SearchResults: convertToSearchResults(resp.Candidates),
    }, nil
}

func (teo *TracedEinoOrchestrator) callSynthesis(ctx context.Context, state *ResearchState) (*SynthesisState, error) {
    trace := opik.TraceFromContext(ctx)

    span, _ := trace.Span(ctx, "eino-node:synthesis",
        opik.WithSpanType(opik.SpanTypeLLM), // LLM-heavy operation
        opik.WithSpanInput(map[string]any{
            "sources_count": len(state.SearchResults),
            "topic":         state.Request.Topic,
        }),
    )

    resp, err := teo.callSynthesisAgent(ctx, &SynthesisRequest{
        Topic:         state.Request.Topic,
        SearchResults: state.SearchResults,
    })

    if err != nil {
        span.End(ctx, opik.WithSpanOutput(map[string]any{"error": err.Error()}))
        return nil, err
    }

    span.End(ctx, opik.WithSpanOutput(map[string]any{
        "candidates_extracted": len(resp.Candidates),
    }))

    return &SynthesisState{
        Request:    state.Request,
        Candidates: resp.Candidates,
    }, nil
}
```

## Part 3: Complete Integration Example

Here's a complete example combining ADK agents with Eino orchestration and full Opik tracing:

```go
package main

import (
    "context"
    "log"

    opik "github.com/grokify/go-comet-ml-opik"
)

func main() {
    // Initialize Opik client
    opikClient, err := opik.NewClient(
        opik.WithProjectName("stats-agent-team"),
        opik.WithAPIKey(os.Getenv("OPIK_API_KEY")),
        opik.WithWorkspace(os.Getenv("OPIK_WORKSPACE")),
    )
    if err != nil {
        log.Fatalf("Failed to create Opik client: %v", err)
    }

    // Create the orchestrator with tracing
    orchestrator := NewTracedOrchestrator(opikClient)

    // Execute a research request
    ctx := context.Background()
    result, err := orchestrator.Research(ctx, &OrchestrationRequest{
        Topic:            "climate change statistics 2024",
        MinVerifiedStats: 10,
        MaxCandidates:    30,
    })

    if err != nil {
        log.Fatalf("Research failed: %v", err)
    }

    log.Printf("Research complete: %d verified statistics", result.VerifiedCount)
}

type TracedOrchestrator struct {
    opikClient  *opik.Client
    eino        *EinoOrchestrator
    synthesis   *SynthesisAgent
    verification *VerificationAgent
}

func (to *TracedOrchestrator) Research(ctx context.Context, req *OrchestrationRequest) (*OrchestrationResponse, error) {
    // Create main trace
    trace, _ := to.opikClient.Trace(ctx, "full-research-workflow",
        opik.WithTraceInput(map[string]any{
            "topic":     req.Topic,
            "min_stats": req.MinVerifiedStats,
        }),
        opik.WithTraceTags("workflow:research", "version:v1"),
    )
    defer trace.End(ctx)

    // Store in context
    ctx = opik.ContextWithTrace(ctx, trace)

    // Execute Eino workflow (which creates child spans)
    result, err := to.eino.Orchestrate(ctx, req)

    // Add evaluation scores
    if err == nil {
        // Calculate quality metrics
        accuracyScore := float64(result.VerifiedCount) / float64(result.TotalCandidates)
        coverageScore := float64(result.VerifiedCount) / float64(req.MinVerifiedStats)
        if coverageScore > 1.0 {
            coverageScore = 1.0
        }

        trace.AddFeedbackScore(ctx, "accuracy", accuracyScore, "Verification accuracy")
        trace.AddFeedbackScore(ctx, "coverage", coverageScore, "Target coverage")
    }

    return result, err
}
```

## Part 4: Viewing Traces in Opik Dashboard

After running your agents with tracing enabled, you can view the traces in the Opik dashboard:

### Trace Hierarchy

```
full-research-workflow (trace)
├── eino-node:validate (span)
├── eino-node:research (span)
│   └── http-call:research-agent (span)
├── eino-node:synthesis (span)
│   ├── llm-extraction (span, type=llm)
│   ├── process-url-0 (span)
│   │   └── llm-call (span, type=llm)
│   ├── process-url-1 (span)
│   │   └── llm-call (span, type=llm)
│   └── ... more URLs
├── eino-node:verification (span)
│   ├── verify-stat-0 (span, type=llm)
│   ├── verify-stat-1 (span, type=llm)
│   └── ... more verifications
└── eino-node:format (span)
```

### Key Metrics to Monitor

1. **Latency**: Total workflow time and per-node breakdown
2. **Token Usage**: LLM tokens per agent and total
3. **Success Rate**: Verification success rate over time
4. **Quality Scores**: Accuracy and coverage metrics

## Best Practices

### 1. Use Meaningful Span Names

```go
// Good: Descriptive and hierarchical
span, _ := trace.Span(ctx, "synthesis:extract-from-url",
    opik.WithSpanMetadata(map[string]any{
        "url_index": i,
        "domain":    result.Domain,
    }),
)

// Avoid: Generic names
span, _ := trace.Span(ctx, "process") // Too vague
```

### 2. Capture Relevant Inputs/Outputs

```go
// Capture enough context for debugging
span, _ := trace.Span(ctx, "llm-extraction",
    opik.WithSpanInput(map[string]any{
        "topic":          topic,
        "content_length": len(content),
        "content_hash":   hashContent(content), // For deduplication
    }),
)

span.End(ctx, opik.WithSpanOutput(map[string]any{
    "stats_count": len(stats),
    "stats":       stats, // Full output for analysis
}))
```

### 3. Add Structured Feedback Scores

```go
// Quantitative metrics
span.AddFeedbackScore(ctx, "extraction_count", float64(len(stats)), "")
span.AddFeedbackScore(ctx, "accuracy", verifiedCount/totalCount, "")

// Categorical with reason
span.AddFeedbackScore(ctx, "quality", 0.8, "Good extraction but missing units")
```

### 4. Use Tags for Filtering

```go
trace, _ := opikClient.Trace(ctx, "workflow",
    opik.WithTraceTags(
        "env:production",
        "agent:synthesis",
        "llm:gemini-2.0-flash",
        "topic:climate",
    ),
)
```

## Conclusion

Integrating Opik with Google ADK and Eino provides comprehensive observability for agentic applications:

- **ADK Integration**: Trace tool invocations and LLM calls within agents
- **Eino Integration**: Trace workflow nodes and graph execution
- **Combined**: Full visibility into multi-agent orchestration

This enables you to debug issues, optimize performance, track costs, and improve agent quality over time.

## Next Steps

- [Evaluation Metrics](../evaluation/metrics.md) - Use Opik's evaluation system to automatically score agent outputs
- [LLM Provider Integrations](../integrations/openai.md) - Direct integration with OpenAI, Anthropic, and other providers
- [Datasets and Experiments](../features/datasets.md) - Run experiments to compare agent configurations
