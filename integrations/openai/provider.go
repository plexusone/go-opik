package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/grokify/go-comet-ml-opik/evaluation/llm"
)

const (
	defaultBaseURL = "https://api.openai.com/v1"
	defaultModel   = "gpt-4o"
)

// Provider implements llm.Provider for OpenAI.
type Provider struct {
	apiKey      string
	baseURL     string
	model       string
	client      *http.Client
	temperature float64
	maxTokens   int
}

// NewProvider creates a new OpenAI provider.
func NewProvider(opts ...Option) *Provider {
	p := &Provider{
		apiKey:  os.Getenv("OPENAI_API_KEY"),
		baseURL: defaultBaseURL,
		model:   defaultModel,
		client:  http.DefaultClient,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// Option configures the provider.
type Option func(*Provider)

// WithAPIKey sets the API key.
func WithAPIKey(key string) Option {
	return func(p *Provider) {
		p.apiKey = key
	}
}

// WithBaseURL sets the base URL (for Azure OpenAI or proxies).
func WithBaseURL(url string) Option {
	return func(p *Provider) {
		p.baseURL = url
	}
}

// WithModel sets the default model.
func WithModel(model string) Option {
	return func(p *Provider) {
		p.model = model
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(p *Provider) {
		p.client = client
	}
}

// WithTemperature sets the default temperature.
func WithTemperature(temp float64) Option {
	return func(p *Provider) {
		p.temperature = temp
	}
}

// WithMaxTokens sets the default max tokens.
func WithMaxTokens(max int) Option {
	return func(p *Provider) {
		p.maxTokens = max
	}
}

// chatRequest represents an OpenAI chat completion request.
type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature *float64      `json:"temperature,omitempty"`
	MaxTokens   *int          `json:"max_tokens,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatResponse represents an OpenAI chat completion response.
type chatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// Complete sends a chat completion request.
func (p *Provider) Complete(ctx context.Context, req llm.CompletionRequest) (*llm.CompletionResponse, error) {
	model := req.Model
	if model == "" {
		model = p.model
	}

	messages := make([]chatMessage, len(req.Messages))
	for i, m := range req.Messages {
		messages[i] = chatMessage{
			Role:    m.Role,
			Content: m.Content,
		}
	}

	chatReq := chatRequest{
		Model:    model,
		Messages: messages,
	}

	temp := req.Temperature
	if temp == 0 && p.temperature != 0 {
		temp = p.temperature
	}
	if temp != 0 {
		chatReq.Temperature = &temp
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 && p.maxTokens != 0 {
		maxTokens = p.maxTokens
	}
	if maxTokens != 0 {
		chatReq.MaxTokens = &maxTokens
	}

	body, err := json.Marshal(chatReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	return &llm.CompletionResponse{
		Content:      chatResp.Choices[0].Message.Content,
		Model:        chatResp.Model,
		PromptTokens: chatResp.Usage.PromptTokens,
		OutputTokens: chatResp.Usage.CompletionTokens,
	}, nil
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return "openai"
}

// DefaultModel returns the default model.
func (p *Provider) DefaultModel() string {
	return p.model
}

// Ensure Provider implements llm.Provider.
var _ llm.Provider = (*Provider)(nil)
