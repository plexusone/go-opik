package anthropic

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
	defaultBaseURL = "https://api.anthropic.com/v1"
	defaultModel   = "claude-sonnet-4-20250514"
	apiVersion     = "2023-06-01"
)

// Provider implements llm.Provider for Anthropic.
type Provider struct {
	apiKey      string
	baseURL     string
	model       string
	client      *http.Client
	temperature float64
	maxTokens   int
}

// NewProvider creates a new Anthropic provider.
func NewProvider(opts ...Option) *Provider {
	p := &Provider{
		apiKey:    os.Getenv("ANTHROPIC_API_KEY"),
		baseURL:   defaultBaseURL,
		model:     defaultModel,
		client:    http.DefaultClient,
		maxTokens: 4096, // Anthropic requires max_tokens
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

// WithBaseURL sets the base URL.
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

// messageRequest represents an Anthropic messages API request.
type messageRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens"`
	Temperature *float64      `json:"temperature,omitempty"`
	System      string        `json:"system,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// messageResponse represents an Anthropic messages API response.
type messageResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Model   string `json:"model"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	StopReason   string `json:"stop_reason"`
	StopSequence string `json:"stop_sequence"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// Complete sends a messages API request.
func (p *Provider) Complete(ctx context.Context, req llm.CompletionRequest) (*llm.CompletionResponse, error) {
	model := req.Model
	if model == "" {
		model = p.model
	}

	// Convert messages, extracting system message if present
	var system string
	messages := make([]chatMessage, 0, len(req.Messages))
	for _, m := range req.Messages {
		if m.Role == "system" {
			system = m.Content
		} else {
			messages = append(messages, chatMessage{
				Role:    m.Role,
				Content: m.Content,
			})
		}
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = p.maxTokens
	}

	msgReq := messageRequest{
		Model:     model,
		Messages:  messages,
		MaxTokens: maxTokens,
		System:    system,
	}

	temp := req.Temperature
	if temp == 0 && p.temperature != 0 {
		temp = p.temperature
	}
	if temp != 0 {
		msgReq.Temperature = &temp
	}

	body, err := json.Marshal(msgReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", apiVersion)

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

	var msgResp messageResponse
	if err := json.Unmarshal(respBody, &msgResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Extract text content
	var content string
	for _, c := range msgResp.Content {
		if c.Type == "text" {
			content = c.Text
			break
		}
	}

	return &llm.CompletionResponse{
		Content:      content,
		Model:        msgResp.Model,
		PromptTokens: msgResp.Usage.InputTokens,
		OutputTokens: msgResp.Usage.OutputTokens,
	}, nil
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return "anthropic"
}

// DefaultModel returns the default model.
func (p *Provider) DefaultModel() string {
	return p.model
}

// Ensure Provider implements llm.Provider.
var _ llm.Provider = (*Provider)(nil)
