package llm

import (
	"context"
	"errors"
	"testing"

	"github.com/grokify/go-comet-ml-opik/evaluation"
)

func TestMessage(t *testing.T) {
	m := Message{
		Role:    "user",
		Content: "Hello",
	}

	if m.Role != "user" {
		t.Errorf("Role = %q, want %q", m.Role, "user")
	}
	if m.Content != "Hello" {
		t.Errorf("Content = %q, want %q", m.Content, "Hello")
	}
}

func TestCompletionRequest(t *testing.T) {
	req := CompletionRequest{
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
		Model:       "gpt-4",
		Temperature: 0.7,
		MaxTokens:   100,
	}

	if len(req.Messages) != 1 {
		t.Errorf("Messages length = %d, want 1", len(req.Messages))
	}
	if req.Model != "gpt-4" {
		t.Errorf("Model = %q, want %q", req.Model, "gpt-4")
	}
	if req.Temperature != 0.7 {
		t.Errorf("Temperature = %v, want 0.7", req.Temperature)
	}
	if req.MaxTokens != 100 {
		t.Errorf("MaxTokens = %d, want 100", req.MaxTokens)
	}
}

func TestCompletionResponse(t *testing.T) {
	resp := CompletionResponse{
		Content:      "Hello!",
		Model:        "gpt-4",
		PromptTokens: 10,
		OutputTokens: 5,
	}

	if resp.Content != "Hello!" {
		t.Errorf("Content = %q, want %q", resp.Content, "Hello!")
	}
	if resp.Model != "gpt-4" {
		t.Errorf("Model = %q, want %q", resp.Model, "gpt-4")
	}
	if resp.PromptTokens != 10 {
		t.Errorf("PromptTokens = %d, want 10", resp.PromptTokens)
	}
	if resp.OutputTokens != 5 {
		t.Errorf("OutputTokens = %d, want 5", resp.OutputTokens)
	}
}

func TestProviderOptions(t *testing.T) {
	cfg := &providerConfig{}

	WithModel("gpt-4")(cfg)
	WithTemperature(0.7)(cfg)
	WithMaxTokens(100)(cfg)
	WithAPIKey("test-key")(cfg)
	WithBaseURL("https://custom.api")(cfg)

	if cfg.model != "gpt-4" {
		t.Errorf("model = %q, want %q", cfg.model, "gpt-4")
	}
	if cfg.temperature != 0.7 {
		t.Errorf("temperature = %v, want 0.7", cfg.temperature)
	}
	if cfg.maxTokens != 100 {
		t.Errorf("maxTokens = %d, want 100", cfg.maxTokens)
	}
	if cfg.apiKey != "test-key" {
		t.Errorf("apiKey = %q, want %q", cfg.apiKey, "test-key")
	}
	if cfg.baseURL != "https://custom.api" {
		t.Errorf("baseURL = %q, want %q", cfg.baseURL, "https://custom.api")
	}
}

func TestSimpleProvider(t *testing.T) {
	p := NewSimpleProvider("test", "test-model", func(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
		return &CompletionResponse{Content: "Response"}, nil
	})

	if p.Name() != "test" {
		t.Errorf("Name() = %q, want %q", p.Name(), "test")
	}
	if p.DefaultModel() != "test-model" {
		t.Errorf("DefaultModel() = %q, want %q", p.DefaultModel(), "test-model")
	}

	resp, err := p.Complete(context.Background(), CompletionRequest{})
	if err != nil {
		t.Fatalf("Complete error: %v", err)
	}
	if resp.Content != "Response" {
		t.Errorf("Content = %q, want %q", resp.Content, "Response")
	}
}

func TestSimpleProviderError(t *testing.T) {
	expectedErr := errors.New("provider error")
	p := NewSimpleProvider("test", "model", func(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
		return nil, expectedErr
	})

	_, err := p.Complete(context.Background(), CompletionRequest{})
	if err != expectedErr {
		t.Errorf("error = %v, want %v", err, expectedErr)
	}
}

func TestMockProvider(t *testing.T) {
	responses := map[string]string{
		"Hello": "Hi there!",
		"Bye":   "Goodbye!",
	}
	p := NewMockProvider(responses, "default response")

	if p.Name() != "mock" {
		t.Errorf("Name() = %q, want %q", p.Name(), "mock")
	}
	if p.DefaultModel() != "mock-model" {
		t.Errorf("DefaultModel() = %q, want %q", p.DefaultModel(), "mock-model")
	}

	t.Run("matches response", func(t *testing.T) {
		resp, err := p.Complete(context.Background(), CompletionRequest{
			Messages: []Message{{Role: "user", Content: "Hello"}},
		})
		if err != nil {
			t.Fatalf("Complete error: %v", err)
		}
		if resp.Content != "Hi there!" {
			t.Errorf("Content = %q, want %q", resp.Content, "Hi there!")
		}
	})

	t.Run("default response", func(t *testing.T) {
		resp, _ := p.Complete(context.Background(), CompletionRequest{
			Messages: []Message{{Role: "user", Content: "Unknown"}},
		})
		if resp.Content != "default response" {
			t.Errorf("Content = %q, want %q", resp.Content, "default response")
		}
	})
}

func TestCachingProvider(t *testing.T) {
	callCount := 0
	inner := NewSimpleProvider("test", "model", func(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
		callCount++
		return &CompletionResponse{Content: "Cached"}, nil
	})

	p := NewCachingProvider(inner)

	if p.Name() != "test" {
		t.Errorf("Name() = %q, want %q", p.Name(), "test")
	}
	if p.DefaultModel() != "model" {
		t.Errorf("DefaultModel() = %q, want %q", p.DefaultModel(), "model")
	}

	// First call
	req := CompletionRequest{Model: "gpt-4", Messages: []Message{{Role: "user", Content: "Hello"}}}
	resp1, _ := p.Complete(context.Background(), req)
	if callCount != 1 {
		t.Errorf("callCount = %d, want 1", callCount)
	}
	if resp1.Content != "Cached" {
		t.Errorf("Content = %q, want %q", resp1.Content, "Cached")
	}

	// Second call with same request - should use cache
	resp2, _ := p.Complete(context.Background(), req)
	if callCount != 1 {
		t.Errorf("callCount = %d, want 1 (cached)", callCount)
	}
	if resp2.Content != "Cached" {
		t.Errorf("Content = %q, want %q", resp2.Content, "Cached")
	}

	// Different request - should call provider
	req2 := CompletionRequest{Model: "gpt-4", Messages: []Message{{Role: "user", Content: "Different"}}}
	_, _ = p.Complete(context.Background(), req2)
	if callCount != 2 {
		t.Errorf("callCount = %d, want 2", callCount)
	}
}

func TestCacheKey(t *testing.T) {
	req1 := CompletionRequest{
		Model: "gpt-4",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
	}
	req2 := CompletionRequest{
		Model: "gpt-4",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
	}
	req3 := CompletionRequest{
		Model: "gpt-4",
		Messages: []Message{
			{Role: "user", Content: "Different"},
		},
	}

	key1 := cacheKey(req1)
	key2 := cacheKey(req2)
	key3 := cacheKey(req3)

	if key1 != key2 {
		t.Error("identical requests should have same cache key")
	}
	if key1 == key3 {
		t.Error("different requests should have different cache keys")
	}
}

func TestParseScoreFromResponse(t *testing.T) {
	tests := []struct {
		name     string
		response string
		want     float64
		wantErr  bool
	}{
		{"JSON with score", `{"score": 0.8, "reason": "good"}`, 0.8, false},
		{"standalone decimal", "0.75", 0.75, false},
		{"score out of 10", "8/10", 0.8, false},
		{"score out of 100", "75 out of 100", 0.75, false},
		{"yes", "Yes, it's good", 1.0, false},
		{"no", "No, it's bad", 0.0, false},
		{"true", "True", 1.0, false},
		{"false", "False", 0.0, false},
		{"unparseable", "random text without score", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseScoreFromResponse(tt.response)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("score = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseJSONResponse(t *testing.T) {
	t.Run("direct JSON", func(t *testing.T) {
		var result map[string]any
		err := ParseJSONResponse(`{"score": 0.8, "reason": "good"}`, &result)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if result["score"] != 0.8 {
			t.Errorf("score = %v, want 0.8", result["score"])
		}
	})

	t.Run("JSON in code block", func(t *testing.T) {
		var result map[string]any
		response := "Here's the result:\n```json\n{\"score\": 0.9}\n```"
		err := ParseJSONResponse(response, &result)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if result["score"] != 0.9 {
			t.Errorf("score = %v, want 0.9", result["score"])
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		var result map[string]any
		err := ParseJSONResponse("not json", &result)
		if err == nil {
			t.Error("expected error for invalid JSON")
		}
	})
}

func TestParseReasonFromResponse(t *testing.T) {
	tests := []struct {
		name     string
		response string
		want     string
	}{
		{"reason pattern", "Reason: The output is accurate.", "The output is accurate."},
		{"explanation pattern", "Explanation: Well structured response.", "Well structured response."},
		{"because pattern", "Because the answer addresses the question.", "the answer addresses the question."},
		{"first sentence", "Good output. More details here.", "Good output"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseReasonFromResponse(tt.response)
			if got != tt.want {
				t.Errorf("reason = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	short := "short"
	if truncate(short, 10) != short {
		t.Error("short string should not be truncated")
	}

	long := "this is a very long string that should be truncated"
	result := truncate(long, 20)
	if len(result) != 20 {
		t.Errorf("length = %d, want 20", len(result))
	}
	if result[len(result)-3:] != "..." {
		t.Error("truncated string should end with ...")
	}
}

func TestFormatPromptTemplate(t *testing.T) {
	template := "Input: {{input}}\nOutput: {{ output }}"
	vars := map[string]string{
		"input":  "question",
		"output": "answer",
	}

	result := FormatPromptTemplate(template, vars)

	if result != "Input: question\nOutput: answer" {
		t.Errorf("result = %q, want formatted template", result)
	}
}

func TestParseScoreResponse(t *testing.T) {
	t.Run("valid JSON", func(t *testing.T) {
		sr, err := ParseScoreResponse(`{"score": 0.8, "reason": "good"}`)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if sr.Score != 0.8 {
			t.Errorf("Score = %v, want 0.8", sr.Score)
		}
		if sr.Reason != "good" {
			t.Errorf("Reason = %q, want %q", sr.Reason, "good")
		}
	})

	t.Run("fallback to parsing", func(t *testing.T) {
		sr, err := ParseScoreResponse("Yes, it's correct")
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if sr.Score != 1.0 {
			t.Errorf("Score = %v, want 1.0", sr.Score)
		}
	})
}

func TestBaseJudge(t *testing.T) {
	provider := NewMockProvider(nil, `{"score": 0.8}`)

	j := NewBaseJudge("test_judge", provider)

	if j.Name() != "test_judge" {
		t.Errorf("Name() = %q, want %q", j.Name(), "test_judge")
	}
	if j.Provider() != provider {
		t.Error("Provider() should return the provider")
	}
	if j.Model() != "mock-model" {
		t.Errorf("Model() = %q, want %q", j.Model(), "mock-model")
	}
}

func TestBaseJudgeWithOptions(t *testing.T) {
	provider := NewMockProvider(nil, "")

	j := NewBaseJudge("test", provider,
		WithJudgeModel("custom-model"),
		WithJudgeTemperature(0.5),
	)

	if j.Model() != "custom-model" {
		t.Errorf("Model() = %q, want %q", j.Model(), "custom-model")
	}
	if j.temperature != 0.5 {
		t.Errorf("temperature = %v, want 0.5", j.temperature)
	}
}

func TestBaseJudgeComplete(t *testing.T) {
	provider := NewMockProvider(map[string]string{
		"test message": `{"score": 0.9}`,
	}, "")

	j := NewBaseJudge("test", provider)

	resp, err := j.Complete(context.Background(), []Message{
		{Role: "user", Content: "test message"},
	})

	if err != nil {
		t.Fatalf("Complete error: %v", err)
	}
	if resp.Content != `{"score": 0.9}` {
		t.Errorf("Content = %q, want JSON", resp.Content)
	}
}

// Test actual LLM metrics with mock provider

func TestGEval(t *testing.T) {
	provider := NewMockProvider(nil, `{"score": 0.85, "reason": "Good response"}`)

	g := NewGEval(provider, "Evaluate accuracy")

	if g.Name() != "g_eval" {
		t.Errorf("Name() = %q, want %q", g.Name(), "g_eval")
	}

	result := g.Score(context.Background(), evaluation.MetricInput{
		Input:  "What is 2+2?",
		Output: "4",
	})

	if result.Value != 0.85 {
		t.Errorf("Score = %v, want 0.85", result.Value)
	}
	if result.Reason != "Good response" {
		t.Errorf("Reason = %q, want %q", result.Reason, "Good response")
	}
}

func TestGEvalWithSteps(t *testing.T) {
	provider := NewMockProvider(nil, `{"score": 0.9}`)

	g := NewGEval(provider, "Evaluate").WithEvaluationSteps([]string{
		"Check accuracy",
		"Check completeness",
	})

	if len(g.steps) != 2 {
		t.Errorf("steps length = %d, want 2", len(g.steps))
	}
}

func TestAnswerRelevance(t *testing.T) {
	provider := NewMockProvider(nil, `{"score": 0.9}`)

	m := NewAnswerRelevance(provider)

	if m.Name() != "answer_relevance" {
		t.Errorf("Name() = %q, want %q", m.Name(), "answer_relevance")
	}

	result := m.Score(context.Background(), evaluation.MetricInput{
		Input:  "What is Python?",
		Output: "Python is a programming language",
	})

	if result.Value != 0.9 {
		t.Errorf("Score = %v, want 0.9", result.Value)
	}
}

func TestHallucination(t *testing.T) {
	provider := NewMockProvider(nil, `{"score": 0.0}`)

	m := NewHallucination(provider)

	if m.Name() != "hallucination" {
		t.Errorf("Name() = %q, want %q", m.Name(), "hallucination")
	}

	result := m.Score(context.Background(), evaluation.MetricInput{
		Output:  "The answer is correct",
		Context: "Reference information",
	})

	if result.Value != 0.0 {
		t.Errorf("Score = %v, want 0.0 (no hallucination)", result.Value)
	}
}

func TestContextRecall(t *testing.T) {
	provider := NewMockProvider(nil, `{"score": 0.8}`)

	m := NewContextRecall(provider)

	if m.Name() != "context_recall" {
		t.Errorf("Name() = %q, want %q", m.Name(), "context_recall")
	}
}

func TestContextPrecision(t *testing.T) {
	provider := NewMockProvider(nil, `{"score": 0.7}`)

	m := NewContextPrecision(provider)

	if m.Name() != "context_precision" {
		t.Errorf("Name() = %q, want %q", m.Name(), "context_precision")
	}
}

func TestModeration(t *testing.T) {
	provider := NewMockProvider(nil, `{"score": 0.0}`)

	m := NewModeration(provider)

	if m.Name() != "moderation" {
		t.Errorf("Name() = %q, want %q", m.Name(), "moderation")
	}

	// Test WithCategories
	m = m.WithCategories([]string{"custom1", "custom2"})
	if len(m.categories) != 2 {
		t.Errorf("categories length = %d, want 2", len(m.categories))
	}
}

func TestFactuality(t *testing.T) {
	provider := NewMockProvider(nil, `{"score": 1.0}`)

	m := NewFactuality(provider)

	if m.Name() != "factuality" {
		t.Errorf("Name() = %q, want %q", m.Name(), "factuality")
	}
}

func TestCoherence(t *testing.T) {
	provider := NewMockProvider(nil, `{"score": 0.95}`)

	m := NewCoherence(provider)

	if m.Name() != "coherence" {
		t.Errorf("Name() = %q, want %q", m.Name(), "coherence")
	}
}

func TestHelpfulness(t *testing.T) {
	provider := NewMockProvider(nil, `{"score": 0.9}`)

	m := NewHelpfulness(provider)

	if m.Name() != "helpfulness" {
		t.Errorf("Name() = %q, want %q", m.Name(), "helpfulness")
	}
}

func TestCustomJudge(t *testing.T) {
	provider := NewMockProvider(nil, `{"score": 0.75}`)

	template := "Evaluate input: {{input}}, output: {{output}}"
	m := NewCustomJudge("custom_metric", template, provider)

	if m.Name() != "custom_metric" {
		t.Errorf("Name() = %q, want %q", m.Name(), "custom_metric")
	}

	result := m.Score(context.Background(), evaluation.MetricInput{
		Input:  "question",
		Output: "answer",
	})

	if result.Value != 0.75 {
		t.Errorf("Score = %v, want 0.75", result.Value)
	}
}

func TestScoreWithRetry(t *testing.T) {
	t.Run("success on first try", func(t *testing.T) {
		provider := NewMockProvider(nil, `{"score": 0.8}`)
		j := NewBaseJudge("test", provider)

		sr, err := ScoreWithRetry(context.Background(), j, []Message{
			{Role: "user", Content: "test"},
		}, 3)

		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if sr.Score != 0.8 {
			t.Errorf("Score = %v, want 0.8", sr.Score)
		}
	})

	t.Run("fails after retries", func(t *testing.T) {
		provider := NewMockProvider(nil, "unparseable response")
		j := NewBaseJudge("test", provider)

		_, err := ScoreWithRetry(context.Background(), j, []Message{
			{Role: "user", Content: "test"},
		}, 2)

		if err == nil {
			t.Error("expected error after retries")
		}
	})
}
