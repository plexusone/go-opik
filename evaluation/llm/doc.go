// Package llm provides LLM-based evaluation metrics.
//
// These metrics use an LLM to evaluate the quality of outputs. They are more
// expensive than heuristic metrics but can evaluate subjective qualities like
// relevance, helpfulness, and coherence.
//
// # Provider Interface
//
// All LLM metrics require a Provider implementation. The package provides:
//   - SimpleProvider: Wraps a completion function
//   - MockProvider: For testing
//   - CachingProvider: Wraps another provider with caching
//
// # Available Metrics
//
//   - GEval: G-EVAL framework with chain-of-thought evaluation
//   - AnswerRelevance: How relevant an answer is to the question
//   - Hallucination: Detects fabricated information
//   - ContextRecall: How well the response uses provided context
//   - ContextPrecision: Whether response sticks to context
//   - Moderation: Content policy violation detection
//   - Factuality: Factual accuracy evaluation
//   - Coherence: Logical coherence assessment
//   - Helpfulness: How helpful the response is
//   - CustomJudge: Create metrics with custom prompts
//
// # Usage Example
//
//	// Create a provider (see integrations package for OpenAI, Anthropic)
//	provider := llm.NewSimpleProvider("custom", "model-name", func(ctx context.Context, req llm.CompletionRequest) (*llm.CompletionResponse, error) {
//	    // Call your LLM API here
//	    return &llm.CompletionResponse{Content: "..."}, nil
//	})
//
//	// Create metrics
//	relevance := llm.NewAnswerRelevance(provider)
//	hallucination := llm.NewHallucination(provider)
//
//	// Evaluate
//	input := evaluation.NewMetricInput("What is 2+2?", "The answer is 4.").WithExpected("4")
//	score := relevance.Score(ctx, input)
//
// # Custom Judge Example
//
//	judge := llm.NewCustomJudge("tone_check", `
//	Evaluate whether the following response maintains a professional tone.
//
//	User message: {{input}}
//	AI response: {{output}}
//
//	Return your response in JSON format:
//	{"score": <0.0-1.0>, "reason": "<explanation>"}
//	`, provider)
package llm
