package main

import (
	"context"
	"fmt"
	"log"
	"time"

	opik "github.com/grokify/go-comet-ml-opik"
)

func main() {
	// Create a new Opik client
	// For Opik Cloud, set OPIK_API_KEY and OPIK_WORKSPACE environment variables
	// For local Opik, just run without configuration (uses localhost:5173)
	client, err := opik.NewClient(
		opik.WithProjectName("My Go Project"),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Create a trace
	trace, err := client.Trace(ctx, "my-trace",
		opik.WithTraceInput(map[string]any{
			"prompt": "Hello, world!",
		}),
		opik.WithTraceTags("example", "go-sdk"),
	)
	if err != nil {
		log.Fatalf("Failed to create trace: %v", err)
	}

	fmt.Printf("Created trace: %s\n", trace.ID())

	// Create a span within the trace
	span, err := trace.Span(ctx, "llm-call",
		opik.WithSpanType(opik.SpanTypeLLM),
		opik.WithSpanModel("gpt-4"),
		opik.WithSpanProvider("openai"),
		opik.WithSpanInput(map[string]any{
			"messages": []map[string]string{
				{"role": "user", "content": "Hello!"},
			},
		}),
	)
	if err != nil {
		log.Fatalf("Failed to create span: %v", err)
	}

	fmt.Printf("Created span: %s\n", span.ID())

	// Simulate some work
	time.Sleep(100 * time.Millisecond)

	// End the span with output
	err = span.End(ctx,
		opik.WithSpanOutput(map[string]any{
			"response": "Hello! How can I help you today?",
		}),
	)
	if err != nil {
		log.Fatalf("Failed to end span: %v", err)
	}

	// Add feedback score
	err = span.AddFeedbackScore(ctx, "quality", 0.95, "Good response")
	if err != nil {
		log.Fatalf("Failed to add feedback score: %v", err)
	}

	// End the trace with output
	err = trace.End(ctx,
		opik.WithTraceOutput(map[string]any{
			"result": "Success",
		}),
	)
	if err != nil {
		log.Fatalf("Failed to end trace: %v", err)
	}

	fmt.Println("Trace completed successfully!")
}
