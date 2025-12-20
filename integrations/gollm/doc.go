// Package gollm provides Opik integration for the gollm LLM wrapper library.
//
// This package provides two main features:
//
//  1. Automatic tracing: Wrap gollm.ChatClient to automatically create spans
//     for all LLM calls with input/output/usage tracking.
//
//  2. Evaluation provider: Use gollm as an LLM provider for evaluation metrics
//     that require LLM-as-judge capabilities.
//
// # Automatic Tracing
//
// Wrap your gollm.ChatClient to automatically trace all LLM calls:
//
//	import (
//	    "github.com/grokify/gollm"
//	    opik "github.com/grokify/go-comet-ml-opik"
//	    opikgollm "github.com/grokify/go-comet-ml-opik/integrations/gollm"
//	)
//
//	// Create gollm client
//	client, _ := gollm.NewClient(gollm.ClientConfig{
//	    Provider: gollm.ProviderNameOpenAI,
//	    APIKey:   os.Getenv("OPENAI_API_KEY"),
//	})
//
//	// Create Opik client
//	opikClient, _ := opik.NewClient()
//
//	// Wrap for tracing
//	tracingClient := opikgollm.NewTracingClient(client, opikClient)
//
//	// Use within a trace context
//	ctx, trace, _ := opik.StartTrace(ctx, opikClient, "my-task")
//	defer trace.End(ctx)
//
//	// All calls are automatically traced
//	resp, _ := tracingClient.CreateChatCompletion(ctx, &gollm.ChatCompletionRequest{
//	    Model:    "gpt-4o",
//	    Messages: []gollm.Message{{Role: gollm.RoleUser, Content: "Hello!"}},
//	})
//
// # Evaluation Provider
//
// Use gollm as a provider for LLM-based evaluation metrics:
//
//	import (
//	    "github.com/grokify/gollm"
//	    opikgollm "github.com/grokify/go-comet-ml-opik/integrations/gollm"
//	    "github.com/grokify/go-comet-ml-opik/evaluation/llm"
//	)
//
//	// Create gollm client
//	client, _ := gollm.NewClient(gollm.ClientConfig{
//	    Provider: gollm.ProviderNameAnthropic,
//	    APIKey:   os.Getenv("ANTHROPIC_API_KEY"),
//	})
//
//	// Create evaluation provider
//	provider := opikgollm.NewProvider(client,
//	    opikgollm.WithModel("claude-3-opus-20240229"),
//	)
//
//	// Use with evaluation metrics
//	relevance := llm.NewAnswerRelevance(provider)
//	hallucination := llm.NewHallucination(provider)
//
// # Streaming Support
//
// The tracing client also supports streaming with automatic span capture:
//
//	stream, _ := tracingClient.CreateChatCompletionStream(ctx, req)
//	defer stream.Close()
//
//	for {
//	    chunk, err := stream.Recv()
//	    if err == io.EOF {
//	        break
//	    }
//	    // Process chunk...
//	}
//	// Span is automatically ended with complete response when stream closes
package gollm
