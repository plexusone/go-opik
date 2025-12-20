package llm

import (
	"context"
)

// Message represents a chat message.
type Message struct {
	Role    string `json:"role"` // "system", "user", "assistant"
	Content string `json:"content"`
}

// CompletionRequest represents a request for chat completion.
type CompletionRequest struct {
	Messages    []Message `json:"messages"`
	Model       string    `json:"model,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

// CompletionResponse represents a chat completion response.
type CompletionResponse struct {
	Content      string `json:"content"`
	Model        string `json:"model,omitempty"`
	PromptTokens int    `json:"prompt_tokens,omitempty"`
	OutputTokens int    `json:"output_tokens,omitempty"`
}

// Provider is an interface for LLM providers used in evaluation.
type Provider interface {
	// Complete sends a completion request and returns the response.
	Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)

	// Name returns the provider name (e.g., "openai", "anthropic").
	Name() string

	// DefaultModel returns the default model for this provider.
	DefaultModel() string
}

// ProviderOption configures a provider.
type ProviderOption func(*providerConfig)

type providerConfig struct {
	model       string
	temperature float64
	maxTokens   int
	apiKey      string
	baseURL     string
}

// WithModel sets the model to use.
func WithModel(model string) ProviderOption {
	return func(c *providerConfig) {
		c.model = model
	}
}

// WithTemperature sets the temperature for generation.
func WithTemperature(temp float64) ProviderOption {
	return func(c *providerConfig) {
		c.temperature = temp
	}
}

// WithMaxTokens sets the maximum tokens for generation.
func WithMaxTokens(max int) ProviderOption {
	return func(c *providerConfig) {
		c.maxTokens = max
	}
}

// WithAPIKey sets the API key.
func WithAPIKey(key string) ProviderOption {
	return func(c *providerConfig) {
		c.apiKey = key
	}
}

// WithBaseURL sets a custom base URL.
func WithBaseURL(url string) ProviderOption {
	return func(c *providerConfig) {
		c.baseURL = url
	}
}

// SimpleProvider is a basic provider implementation using a function.
type SimpleProvider struct {
	name         string
	defaultModel string
	fn           func(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
}

// NewSimpleProvider creates a provider from a completion function.
func NewSimpleProvider(name, defaultModel string, fn func(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)) *SimpleProvider {
	return &SimpleProvider{
		name:         name,
		defaultModel: defaultModel,
		fn:           fn,
	}
}

// Complete sends a completion request.
func (p *SimpleProvider) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	return p.fn(ctx, req)
}

// Name returns the provider name.
func (p *SimpleProvider) Name() string {
	return p.name
}

// DefaultModel returns the default model.
func (p *SimpleProvider) DefaultModel() string {
	return p.defaultModel
}

// MockProvider is a provider that returns predefined responses for testing.
type MockProvider struct {
	name         string
	defaultModel string
	responses    map[string]string // prompt -> response mapping
	defaultResp  string
}

// NewMockProvider creates a mock provider for testing.
func NewMockProvider(responses map[string]string, defaultResp string) *MockProvider {
	return &MockProvider{
		name:         "mock",
		defaultModel: "mock-model",
		responses:    responses,
		defaultResp:  defaultResp,
	}
}

// Complete returns a predefined response.
func (p *MockProvider) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	// Look for matching response
	for _, msg := range req.Messages {
		if resp, ok := p.responses[msg.Content]; ok {
			return &CompletionResponse{
				Content: resp,
				Model:   p.defaultModel,
			}, nil
		}
	}

	return &CompletionResponse{
		Content: p.defaultResp,
		Model:   p.defaultModel,
	}, nil
}

// Name returns the provider name.
func (p *MockProvider) Name() string {
	return p.name
}

// DefaultModel returns the default model.
func (p *MockProvider) DefaultModel() string {
	return p.defaultModel
}

// CachingProvider wraps a provider with response caching.
type CachingProvider struct {
	inner Provider
	cache map[string]*CompletionResponse
}

// NewCachingProvider creates a caching wrapper around a provider.
func NewCachingProvider(inner Provider) *CachingProvider {
	return &CachingProvider{
		inner: inner,
		cache: make(map[string]*CompletionResponse),
	}
}

// Complete returns cached response or calls inner provider.
func (p *CachingProvider) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	key := cacheKey(req)
	if resp, ok := p.cache[key]; ok {
		return resp, nil
	}

	resp, err := p.inner.Complete(ctx, req)
	if err == nil {
		p.cache[key] = resp
	}
	return resp, err
}

// Name returns the inner provider name.
func (p *CachingProvider) Name() string {
	return p.inner.Name()
}

// DefaultModel returns the inner provider's default model.
func (p *CachingProvider) DefaultModel() string {
	return p.inner.DefaultModel()
}

func cacheKey(req CompletionRequest) string {
	key := req.Model + ":"
	for _, msg := range req.Messages {
		key += msg.Role + ":" + msg.Content + "|"
	}
	return key
}
