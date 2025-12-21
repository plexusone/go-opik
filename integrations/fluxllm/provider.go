package fluxllm

import (
	"context"

	"github.com/grokify/go-comet-ml-opik/evaluation/llm"
	fluxllm "github.com/grokify/fluxllm"
	"github.com/grokify/fluxllm/provider"
)

// Provider implements llm.Provider using a gollm.ChatClient.
type Provider struct {
	client      *fluxllm.ChatClient
	model       string
	temperature float64
	maxTokens   int
}

// NewProvider creates a new evaluation provider using gollm.
func NewProvider(client *fluxllm.ChatClient, opts ...Option) *Provider {
	p := &Provider{
		client: client,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// Option configures the provider.
type Option func(*Provider)

// WithModel sets the model to use for completions.
func WithModel(model string) Option {
	return func(p *Provider) {
		p.model = model
	}
}

// WithTemperature sets the temperature for completions.
func WithTemperature(temp float64) Option {
	return func(p *Provider) {
		p.temperature = temp
	}
}

// WithMaxTokens sets the maximum tokens for completions.
func WithMaxTokens(max int) Option {
	return func(p *Provider) {
		p.maxTokens = max
	}
}

// Complete sends a chat completion request using gollm.
func (p *Provider) Complete(ctx context.Context, req llm.CompletionRequest) (*llm.CompletionResponse, error) {
	// Convert llm.Message to gollm provider.Message
	messages := make([]provider.Message, len(req.Messages))
	for i, m := range req.Messages {
		messages[i] = provider.Message{
			Role:    provider.Role(m.Role),
			Content: m.Content,
		}
	}

	// Build request
	gollmReq := &provider.ChatCompletionRequest{
		Model:    req.Model,
		Messages: messages,
	}

	// Use provider defaults if not specified in request
	if gollmReq.Model == "" && p.model != "" {
		gollmReq.Model = p.model
	}

	temp := req.Temperature
	if temp == 0 && p.temperature != 0 {
		temp = p.temperature
	}
	if temp != 0 {
		gollmReq.Temperature = &temp
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 && p.maxTokens != 0 {
		maxTokens = p.maxTokens
	}
	if maxTokens != 0 {
		gollmReq.MaxTokens = &maxTokens
	}

	// Make the call
	resp, err := p.client.CreateChatCompletion(ctx, gollmReq)
	if err != nil {
		return nil, err
	}

	// Extract response content
	content := ""
	if len(resp.Choices) > 0 {
		content = resp.Choices[0].Message.Content
	}

	return &llm.CompletionResponse{
		Content:      content,
		Model:        resp.Model,
		PromptTokens: resp.Usage.PromptTokens,
		OutputTokens: resp.Usage.CompletionTokens,
	}, nil
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return "gollm"
}

// DefaultModel returns the configured default model.
func (p *Provider) DefaultModel() string {
	return p.model
}

// Ensure Provider implements llm.Provider.
var _ llm.Provider = (*Provider)(nil)
