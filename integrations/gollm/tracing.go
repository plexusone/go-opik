package gollm

import (
	"context"
	"io"
	"strings"
	"time"

	opik "github.com/grokify/go-comet-ml-opik"
	gollmlib "github.com/grokify/gollm"
	"github.com/grokify/gollm/provider"
)

// TracingClient wraps a gollm.ChatClient with automatic Opik tracing.
type TracingClient struct {
	client      *gollmlib.ChatClient
	opikClient  *opik.Client
	spanOptions []opik.SpanOption
}

// NewTracingClient creates a new tracing client wrapper.
func NewTracingClient(client *gollmlib.ChatClient, opikClient *opik.Client, opts ...opik.SpanOption) *TracingClient {
	return &TracingClient{
		client:      client,
		opikClient:  opikClient,
		spanOptions: opts,
	}
}

// CreateChatCompletion creates a chat completion with automatic tracing.
func (t *TracingClient) CreateChatCompletion(ctx context.Context, req *provider.ChatCompletionRequest) (*provider.ChatCompletionResponse, error) {
	// Prepare span options
	opts := append([]opik.SpanOption{
		opik.WithSpanType(opik.SpanTypeLLM),
		opik.WithSpanProvider("gollm"),
		opik.WithSpanInput(requestToMap(req)),
	}, t.spanOptions...)

	if req.Model != "" {
		opts = append(opts, opik.WithSpanModel(req.Model))
	}

	// Try to create span from context
	var span *opik.Span
	var err error

	if parentSpan := opik.SpanFromContext(ctx); parentSpan != nil {
		span, err = parentSpan.Span(ctx, "gollm.chat", opts...)
	} else if trace := opik.TraceFromContext(ctx); trace != nil {
		span, err = trace.Span(ctx, "gollm.chat", opts...)
	}

	// Execute request
	startTime := time.Now()
	resp, respErr := t.client.CreateChatCompletion(ctx, req)
	duration := time.Since(startTime)

	// End span with results
	if span != nil && err == nil {
		endOpts := []opik.SpanOption{}

		if resp != nil {
			endOpts = append(endOpts, opik.WithSpanOutput(responseToMap(resp)))

			// Add usage metadata
			metadata := map[string]any{
				"duration_ms":       duration.Milliseconds(),
				"prompt_tokens":     resp.Usage.PromptTokens,
				"completion_tokens": resp.Usage.CompletionTokens,
				"total_tokens":      resp.Usage.TotalTokens,
			}
			if resp.Model != "" {
				metadata["model"] = resp.Model
			}
			endOpts = append(endOpts, opik.WithSpanMetadata(metadata))
		}

		if respErr != nil {
			endOpts = append(endOpts, opik.WithSpanMetadata(map[string]any{
				"error": respErr.Error(),
			}))
		}

		_ = span.End(ctx, endOpts...)
	}

	return resp, respErr
}

// CreateChatCompletionStream creates a streaming chat completion with automatic tracing.
func (t *TracingClient) CreateChatCompletionStream(ctx context.Context, req *provider.ChatCompletionRequest) (provider.ChatCompletionStream, error) {
	// Prepare span options
	opts := append([]opik.SpanOption{
		opik.WithSpanType(opik.SpanTypeLLM),
		opik.WithSpanProvider("gollm"),
		opik.WithSpanInput(requestToMap(req)),
	}, t.spanOptions...)

	if req.Model != "" {
		opts = append(opts, opik.WithSpanModel(req.Model))
	}

	// Try to create span from context
	var span *opik.Span
	var err error

	if parentSpan := opik.SpanFromContext(ctx); parentSpan != nil {
		span, err = parentSpan.Span(ctx, "gollm.chat.stream", opts...)
	} else if trace := opik.TraceFromContext(ctx); trace != nil {
		span, err = trace.Span(ctx, "gollm.chat.stream", opts...)
	}

	// Create the stream
	stream, streamErr := t.client.CreateChatCompletionStream(ctx, req)
	if streamErr != nil {
		// End span with error if stream creation failed
		if span != nil && err == nil {
			_ = span.End(ctx, opik.WithSpanMetadata(map[string]any{
				"error": streamErr.Error(),
			}))
		}
		return nil, streamErr
	}

	// Wrap stream to capture output when complete
	return &tracingStream{
		stream:    stream,
		span:      span,
		ctx:       ctx,
		startTime: time.Now(),
	}, nil
}

// CreateChatCompletionWithMemory creates a chat completion using conversation memory with tracing.
func (t *TracingClient) CreateChatCompletionWithMemory(ctx context.Context, sessionID string, req *provider.ChatCompletionRequest) (*provider.ChatCompletionResponse, error) {
	// Prepare span options
	opts := append([]opik.SpanOption{
		opik.WithSpanType(opik.SpanTypeLLM),
		opik.WithSpanProvider("gollm"),
		opik.WithSpanInput(requestToMap(req)),
		opik.WithSpanMetadata(map[string]any{"session_id": sessionID}),
	}, t.spanOptions...)

	if req.Model != "" {
		opts = append(opts, opik.WithSpanModel(req.Model))
	}

	// Try to create span from context
	var span *opik.Span
	var err error

	if parentSpan := opik.SpanFromContext(ctx); parentSpan != nil {
		span, err = parentSpan.Span(ctx, "gollm.chat.memory", opts...)
	} else if trace := opik.TraceFromContext(ctx); trace != nil {
		span, err = trace.Span(ctx, "gollm.chat.memory", opts...)
	}

	// Execute request
	startTime := time.Now()
	resp, respErr := t.client.CreateChatCompletionWithMemory(ctx, sessionID, req)
	duration := time.Since(startTime)

	// End span with results
	if span != nil && err == nil {
		endOpts := []opik.SpanOption{}

		if resp != nil {
			endOpts = append(endOpts, opik.WithSpanOutput(responseToMap(resp)))

			metadata := map[string]any{
				"duration_ms":       duration.Milliseconds(),
				"prompt_tokens":     resp.Usage.PromptTokens,
				"completion_tokens": resp.Usage.CompletionTokens,
				"total_tokens":      resp.Usage.TotalTokens,
			}
			endOpts = append(endOpts, opik.WithSpanMetadata(metadata))
		}

		if respErr != nil {
			endOpts = append(endOpts, opik.WithSpanMetadata(map[string]any{
				"error": respErr.Error(),
			}))
		}

		_ = span.End(ctx, endOpts...)
	}

	return resp, respErr
}

// Close closes the underlying client.
func (t *TracingClient) Close() error {
	return t.client.Close()
}

// Client returns the underlying gollm.ChatClient.
func (t *TracingClient) Client() *gollmlib.ChatClient {
	return t.client
}

// tracingStream wraps a ChatCompletionStream to capture the complete response.
type tracingStream struct {
	stream    provider.ChatCompletionStream
	span      *opik.Span
	ctx       context.Context
	startTime time.Time

	// Buffer to collect the complete response
	responseBuffer strings.Builder
	model          string
	usage          *provider.Usage
	closed         bool
}

// Recv receives the next chunk from the stream.
func (s *tracingStream) Recv() (*provider.ChatCompletionChunk, error) {
	chunk, err := s.stream.Recv()
	if err != nil {
		if err == io.EOF && !s.closed {
			s.endSpan(nil)
			s.closed = true
		} else if err.Error() == "EOF" && !s.closed {
			s.endSpan(nil)
			s.closed = true
		}
		return chunk, err
	}

	// Buffer the response content
	if len(chunk.Choices) > 0 && chunk.Choices[0].Delta != nil {
		s.responseBuffer.WriteString(chunk.Choices[0].Delta.Content)
	}

	// Capture model name
	if chunk.Model != "" && s.model == "" {
		s.model = chunk.Model
	}

	// Capture usage if provided
	if chunk.Usage != nil {
		s.usage = chunk.Usage
	}

	return chunk, nil
}

// Close closes the stream and ends the span.
func (s *tracingStream) Close() error {
	if !s.closed {
		s.endSpan(nil)
		s.closed = true
	}
	return s.stream.Close()
}

// endSpan ends the span with the collected response data.
func (s *tracingStream) endSpan(err error) {
	if s.span == nil {
		return
	}

	duration := time.Since(s.startTime)
	endOpts := []opik.SpanOption{}

	// Add the accumulated response as output
	if s.responseBuffer.Len() > 0 {
		endOpts = append(endOpts, opik.WithSpanOutput(map[string]any{
			"content": s.responseBuffer.String(),
			"model":   s.model,
		}))
	}

	// Add metadata
	metadata := map[string]any{
		"duration_ms": duration.Milliseconds(),
		"streaming":   true,
	}
	if s.model != "" {
		metadata["model"] = s.model
	}
	if s.usage != nil {
		metadata["prompt_tokens"] = s.usage.PromptTokens
		metadata["completion_tokens"] = s.usage.CompletionTokens
		metadata["total_tokens"] = s.usage.TotalTokens
	}
	endOpts = append(endOpts, opik.WithSpanMetadata(metadata))

	if err != nil {
		endOpts = append(endOpts, opik.WithSpanMetadata(map[string]any{
			"error": err.Error(),
		}))
	}

	_ = s.span.End(s.ctx, endOpts...)
}

// requestToMap converts a ChatCompletionRequest to a map for span input.
func requestToMap(req *provider.ChatCompletionRequest) map[string]any {
	m := map[string]any{
		"model": req.Model,
	}

	// Convert messages
	messages := make([]map[string]any, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = map[string]any{
			"role":    string(msg.Role),
			"content": msg.Content,
		}
	}
	m["messages"] = messages

	if req.MaxTokens != nil {
		m["max_tokens"] = *req.MaxTokens
	}
	if req.Temperature != nil {
		m["temperature"] = *req.Temperature
	}
	if req.TopP != nil {
		m["top_p"] = *req.TopP
	}
	if len(req.Stop) > 0 {
		m["stop"] = req.Stop
	}

	return m
}

// responseToMap converts a ChatCompletionResponse to a map for span output.
func responseToMap(resp *provider.ChatCompletionResponse) map[string]any {
	m := map[string]any{
		"id":      resp.ID,
		"model":   resp.Model,
		"created": resp.Created,
	}

	// Extract content from first choice
	if len(resp.Choices) > 0 {
		choice := resp.Choices[0]
		m["content"] = choice.Message.Content
		if choice.FinishReason != nil {
			m["finish_reason"] = *choice.FinishReason
		}
	}

	// Add usage
	m["usage"] = map[string]any{
		"prompt_tokens":     resp.Usage.PromptTokens,
		"completion_tokens": resp.Usage.CompletionTokens,
		"total_tokens":      resp.Usage.TotalTokens,
	}

	return m
}
