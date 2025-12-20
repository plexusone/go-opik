package opik

import (
	"testing"
	"time"
)

func TestStreamChunk(t *testing.T) {
	now := time.Now()
	chunk := StreamChunk{
		Content:      "Hello",
		Index:        0,
		IsFirst:      true,
		IsLast:       false,
		Timestamp:    now,
		TokenCount:   5,
		FinishReason: "",
		Metadata:     map[string]any{"key": "value"},
	}

	if chunk.Content != "Hello" {
		t.Errorf("Content = %q, want %q", chunk.Content, "Hello")
	}
	if chunk.Index != 0 {
		t.Errorf("Index = %d, want 0", chunk.Index)
	}
	if !chunk.IsFirst {
		t.Error("IsFirst should be true")
	}
	if chunk.IsLast {
		t.Error("IsLast should be false")
	}
	if chunk.TokenCount != 5 {
		t.Errorf("TokenCount = %d, want 5", chunk.TokenCount)
	}
}

func TestNewStreamAccumulator(t *testing.T) {
	acc := NewStreamAccumulator()

	if acc == nil {
		t.Fatal("NewStreamAccumulator returned nil")
	}
	if acc.Content() != "" {
		t.Errorf("Content = %q, want empty", acc.Content())
	}
	if acc.TotalTokens() != 0 {
		t.Errorf("TotalTokens = %d, want 0", acc.TotalTokens())
	}
	if acc.ChunkCount() != 0 {
		t.Errorf("ChunkCount = %d, want 0", acc.ChunkCount())
	}
	if acc.FinishReason() != "" {
		t.Errorf("FinishReason = %q, want empty", acc.FinishReason())
	}
}

func TestStreamAccumulatorAddChunk(t *testing.T) {
	acc := NewStreamAccumulator()

	chunk1 := StreamChunk{
		Content:    "Hello ",
		Timestamp:  time.Now(),
		TokenCount: 2,
	}
	chunk2 := StreamChunk{
		Content:      "world!",
		Timestamp:    time.Now().Add(100 * time.Millisecond),
		TokenCount:   1,
		FinishReason: "stop",
	}

	acc.AddChunk(chunk1)
	acc.AddChunk(chunk2)

	if acc.Content() != "Hello world!" {
		t.Errorf("Content = %q, want %q", acc.Content(), "Hello world!")
	}
	if acc.TotalTokens() != 3 {
		t.Errorf("TotalTokens = %d, want 3", acc.TotalTokens())
	}
	if acc.ChunkCount() != 2 {
		t.Errorf("ChunkCount = %d, want 2", acc.ChunkCount())
	}
	if acc.FinishReason() != "stop" {
		t.Errorf("FinishReason = %q, want %q", acc.FinishReason(), "stop")
	}
}

func TestStreamAccumulatorDuration(t *testing.T) {
	acc := NewStreamAccumulator()

	// No chunks - duration should be 0
	if acc.Duration() != 0 {
		t.Errorf("Duration with no chunks = %v, want 0", acc.Duration())
	}

	// Add chunks with time gap
	now := time.Now()
	chunk1 := StreamChunk{Content: "a", Timestamp: now}
	chunk2 := StreamChunk{Content: "b", Timestamp: now.Add(100 * time.Millisecond)}

	acc.AddChunk(chunk1)
	acc.AddChunk(chunk2)

	duration := acc.Duration()
	if duration < 90*time.Millisecond || duration > 110*time.Millisecond {
		t.Errorf("Duration = %v, want ~100ms", duration)
	}
}

func TestStreamAccumulatorTimeToFirstChunk(t *testing.T) {
	acc := NewStreamAccumulator()
	streamStart := time.Now()

	// No chunks - TTFC should be 0
	if acc.TimeToFirstChunk(streamStart) != 0 {
		t.Errorf("TTFC with no chunks = %v, want 0", acc.TimeToFirstChunk(streamStart))
	}

	// Add first chunk after some delay
	time.Sleep(50 * time.Millisecond)
	chunk := StreamChunk{Content: "a", Timestamp: time.Now()}
	acc.AddChunk(chunk)

	ttfc := acc.TimeToFirstChunk(streamStart)
	if ttfc < 40*time.Millisecond || ttfc > 100*time.Millisecond {
		t.Errorf("TTFC = %v, want ~50ms", ttfc)
	}
}

func TestStreamAccumulatorMetadata(t *testing.T) {
	acc := NewStreamAccumulator()

	chunk1 := StreamChunk{
		Content:   "a",
		Timestamp: time.Now(),
		Metadata:  map[string]any{"key1": "value1"},
	}
	chunk2 := StreamChunk{
		Content:   "b",
		Timestamp: time.Now(),
		Metadata:  map[string]any{"key2": "value2"},
	}

	acc.AddChunk(chunk1)
	acc.AddChunk(chunk2)

	metadata := acc.Metadata()
	if metadata["key1"] != "value1" {
		t.Errorf("Metadata[key1] = %v, want %q", metadata["key1"], "value1")
	}
	if metadata["key2"] != "value2" {
		t.Errorf("Metadata[key2] = %v, want %q", metadata["key2"], "value2")
	}
}

func TestStreamAccumulatorToOutput(t *testing.T) {
	acc := NewStreamAccumulator()

	chunk := StreamChunk{
		Content:      "Hello",
		Timestamp:    time.Now(),
		TokenCount:   5,
		FinishReason: "stop",
	}
	acc.AddChunk(chunk)

	output := acc.ToOutput()

	if output["content"] != "Hello" {
		t.Errorf("output[content] = %v, want %q", output["content"], "Hello")
	}
	if output["chunk_count"] != 1 {
		t.Errorf("output[chunk_count] = %v, want 1", output["chunk_count"])
	}
	if output["total_tokens"] != 5 {
		t.Errorf("output[total_tokens] = %v, want 5", output["total_tokens"])
	}
	if output["finish_reason"] != "stop" {
		t.Errorf("output[finish_reason] = %v, want %q", output["finish_reason"], "stop")
	}
}

func TestWithChunkTokenCount(t *testing.T) {
	chunk := StreamChunk{}
	WithChunkTokenCount(10)(&chunk)

	if chunk.TokenCount != 10 {
		t.Errorf("TokenCount = %d, want 10", chunk.TokenCount)
	}
}

func TestWithChunkFinishReason(t *testing.T) {
	chunk := StreamChunk{}
	WithChunkFinishReason("stop")(&chunk)

	if chunk.FinishReason != "stop" {
		t.Errorf("FinishReason = %q, want %q", chunk.FinishReason, "stop")
	}
	if !chunk.IsLast {
		t.Error("IsLast should be true when finish reason is set")
	}
}

func TestWithChunkMetadata(t *testing.T) {
	chunk := StreamChunk{Metadata: make(map[string]any)}
	WithChunkMetadata("model", "gpt-4")(&chunk)

	if chunk.Metadata["model"] != "gpt-4" {
		t.Errorf("Metadata[model] = %v, want %q", chunk.Metadata["model"], "gpt-4")
	}
}

func TestNewStreamingSpan(t *testing.T) {
	span := &Span{id: "span-123", traceID: "trace-456"}
	streamingSpan := NewStreamingSpan(span)

	if streamingSpan == nil {
		t.Fatal("NewStreamingSpan returned nil")
	}
	if streamingSpan.Span() != span {
		t.Error("Span() should return underlying span")
	}
	if streamingSpan.ID() != "span-123" {
		t.Errorf("ID() = %q, want %q", streamingSpan.ID(), "span-123")
	}
	if streamingSpan.TraceID() != "trace-456" {
		t.Errorf("TraceID() = %q, want %q", streamingSpan.TraceID(), "trace-456")
	}
	if streamingSpan.Accumulator() == nil {
		t.Error("Accumulator() should not be nil")
	}
}

func TestStreamingSpanAddChunk(t *testing.T) {
	span := &Span{id: "span-123"}
	streamingSpan := NewStreamingSpan(span)

	streamingSpan.AddChunk("Hello ")
	streamingSpan.AddChunk("world!")

	acc := streamingSpan.Accumulator()
	if acc.Content() != "Hello world!" {
		t.Errorf("Content = %q, want %q", acc.Content(), "Hello world!")
	}
	if acc.ChunkCount() != 2 {
		t.Errorf("ChunkCount = %d, want 2", acc.ChunkCount())
	}
}

func TestStreamingSpanAddChunkWithOptions(t *testing.T) {
	span := &Span{id: "span-123"}
	streamingSpan := NewStreamingSpan(span)

	streamingSpan.AddChunk("Hello", WithChunkTokenCount(5), WithChunkMetadata("test", true))

	acc := streamingSpan.Accumulator()
	if acc.TotalTokens() != 5 {
		t.Errorf("TotalTokens = %d, want 5", acc.TotalTokens())
	}
}

func TestStreamingSpanOnChunk(t *testing.T) {
	span := &Span{id: "span-123"}
	streamingSpan := NewStreamingSpan(span)

	var received []string
	streamingSpan.OnChunk(func(chunk StreamChunk) {
		received = append(received, chunk.Content)
	})

	streamingSpan.AddChunk("a")
	streamingSpan.AddChunk("b")
	streamingSpan.AddChunk("c")

	if len(received) != 3 {
		t.Errorf("received %d chunks, want 3", len(received))
	}
	if received[0] != "a" || received[1] != "b" || received[2] != "c" {
		t.Errorf("received = %v, want [a b c]", received)
	}
}

func TestNewBufferingStreamHandler(t *testing.T) {
	var result string
	handler := NewBufferingStreamHandler(func(content string) error {
		result = content
		return nil
	})

	if handler == nil {
		t.Fatal("NewBufferingStreamHandler returned nil")
	}

	// Add chunks
	_ = handler.HandleChunk(StreamChunk{Content: "Hello "})
	_ = handler.HandleChunk(StreamChunk{Content: "world!"})

	// Check content before finalize
	if handler.Content() != "Hello world!" {
		t.Errorf("Content = %q, want %q", handler.Content(), "Hello world!")
	}

	// Finalize
	if err := handler.Finalize(); err != nil {
		t.Errorf("Finalize error = %v", err)
	}
	if result != "Hello world!" {
		t.Errorf("result = %q, want %q", result, "Hello world!")
	}
}

func TestBufferingStreamHandlerOnChunk(t *testing.T) {
	handler := NewBufferingStreamHandler(nil)

	var chunks []string
	handler.OnChunk(func(chunk StreamChunk) error {
		chunks = append(chunks, chunk.Content)
		return nil
	})

	_ = handler.HandleChunk(StreamChunk{Content: "a"})
	_ = handler.HandleChunk(StreamChunk{Content: "b"})

	if len(chunks) != 2 {
		t.Errorf("received %d chunks, want 2", len(chunks))
	}
}

func TestBufferingStreamHandlerAccumulator(t *testing.T) {
	handler := NewBufferingStreamHandler(nil)
	_ = handler.HandleChunk(StreamChunk{Content: "test", TokenCount: 1})

	acc := handler.Accumulator()
	if acc == nil {
		t.Fatal("Accumulator() returned nil")
	}
	if acc.TotalTokens() != 1 {
		t.Errorf("TotalTokens = %d, want 1", acc.TotalTokens())
	}
}

func TestNewTracingStreamHandler(t *testing.T) {
	span := &Span{id: "span-123"}

	// Create a simple inner handler
	inner := NewBufferingStreamHandler(nil)

	handler := NewTracingStreamHandler(inner, span)
	if handler == nil {
		t.Fatal("NewTracingStreamHandler returned nil")
	}

	if handler.StreamingSpan() == nil {
		t.Error("StreamingSpan() should not be nil")
	}
}

func TestTracingStreamHandlerHandleChunk(t *testing.T) {
	span := &Span{id: "span-123"}
	inner := NewBufferingStreamHandler(nil)
	handler := NewTracingStreamHandler(inner, span)

	chunk := StreamChunk{Content: "test", TokenCount: 5}
	if err := handler.HandleChunk(chunk); err != nil {
		t.Errorf("HandleChunk error = %v", err)
	}

	// Check that chunk was added to accumulator
	if handler.StreamingSpan().Accumulator().TotalTokens() != 5 {
		t.Error("chunk should be tracked in streaming span")
	}

	// Check that inner handler also received it
	if inner.Content() != "test" {
		t.Error("inner handler should receive chunk")
	}
}
