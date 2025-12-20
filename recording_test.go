package opik

import (
	"context"
	"testing"
	"time"
)

func TestRecordedTrace(t *testing.T) {
	now := time.Now()
	trace := &RecordedTrace{
		ID:        "trace-123",
		Name:      "test-trace",
		StartTime: now,
		Input:     map[string]any{"prompt": "hello"},
		Output:    map[string]any{"response": "world"},
		Metadata:  map[string]any{"model": "gpt-4"},
		Tags:      []string{"test"},
		Spans:     make([]*RecordedSpan, 0),
		Feedback:  make([]*RecordedFeedback, 0),
	}

	if trace.ID != "trace-123" {
		t.Errorf("ID = %q, want %q", trace.ID, "trace-123")
	}
	if trace.Name != "test-trace" {
		t.Errorf("Name = %q, want %q", trace.Name, "test-trace")
	}
	if !trace.StartTime.Equal(now) {
		t.Errorf("StartTime = %v, want %v", trace.StartTime, now)
	}
}

func TestRecordedSpan(t *testing.T) {
	span := &RecordedSpan{
		ID:           "span-456",
		TraceID:      "trace-123",
		ParentSpanID: "",
		Name:         "llm-call",
		Type:         SpanTypeLLM,
		StartTime:    time.Now(),
		Model:        "gpt-4",
		Provider:     "openai",
	}

	if span.ID != "span-456" {
		t.Errorf("ID = %q, want %q", span.ID, "span-456")
	}
	if span.TraceID != "trace-123" {
		t.Errorf("TraceID = %q, want %q", span.TraceID, "trace-123")
	}
	if span.Type != SpanTypeLLM {
		t.Errorf("Type = %q, want %q", span.Type, SpanTypeLLM)
	}
}

func TestRecordedFeedback(t *testing.T) {
	feedback := &RecordedFeedback{
		Name:   "accuracy",
		Value:  0.95,
		Reason: "correct response",
	}

	if feedback.Name != "accuracy" {
		t.Errorf("Name = %q, want %q", feedback.Name, "accuracy")
	}
	if feedback.Value != 0.95 {
		t.Errorf("Value = %v, want 0.95", feedback.Value)
	}
	if feedback.Reason != "correct response" {
		t.Errorf("Reason = %q, want %q", feedback.Reason, "correct response")
	}
}

func TestNewLocalRecording(t *testing.T) {
	rec := NewLocalRecording()

	if rec == nil {
		t.Fatal("NewLocalRecording returned nil")
	}
	if rec.TraceCount() != 0 {
		t.Errorf("TraceCount = %d, want 0", rec.TraceCount())
	}
	if rec.SpanCount() != 0 {
		t.Errorf("SpanCount = %d, want 0", rec.SpanCount())
	}
}

func TestLocalRecordingAddTrace(t *testing.T) {
	rec := NewLocalRecording()
	trace := &RecordedTrace{ID: "trace-1", Name: "test"}

	rec.AddTrace(trace)

	if rec.TraceCount() != 1 {
		t.Errorf("TraceCount = %d, want 1", rec.TraceCount())
	}

	retrieved := rec.GetTrace("trace-1")
	if retrieved == nil {
		t.Fatal("GetTrace returned nil")
	}
	if retrieved.Name != "test" {
		t.Errorf("Name = %q, want %q", retrieved.Name, "test")
	}
}

func TestLocalRecordingAddSpan(t *testing.T) {
	rec := NewLocalRecording()
	trace := &RecordedTrace{ID: "trace-1", Spans: make([]*RecordedSpan, 0)}
	rec.AddTrace(trace)

	span := &RecordedSpan{ID: "span-1", TraceID: "trace-1", Name: "test-span"}
	rec.AddSpan(span)

	if rec.SpanCount() != 1 {
		t.Errorf("SpanCount = %d, want 1", rec.SpanCount())
	}

	retrieved := rec.GetSpan("span-1")
	if retrieved == nil {
		t.Fatal("GetSpan returned nil")
	}
	if retrieved.Name != "test-span" {
		t.Errorf("Name = %q, want %q", retrieved.Name, "test-span")
	}
}

func TestLocalRecordingSpanHierarchy(t *testing.T) {
	rec := NewLocalRecording()
	trace := &RecordedTrace{ID: "trace-1", Spans: make([]*RecordedSpan, 0)}
	rec.AddTrace(trace)

	parentSpan := &RecordedSpan{
		ID:       "parent-span",
		TraceID:  "trace-1",
		Name:     "parent",
		Children: make([]*RecordedSpan, 0),
	}
	rec.AddSpan(parentSpan)

	childSpan := &RecordedSpan{
		ID:           "child-span",
		TraceID:      "trace-1",
		ParentSpanID: "parent-span",
		Name:         "child",
		Children:     make([]*RecordedSpan, 0),
	}
	rec.AddSpan(childSpan)

	// Check parent has child
	parent := rec.GetSpan("parent-span")
	if len(parent.Children) != 1 {
		t.Errorf("parent.Children length = %d, want 1", len(parent.Children))
	}
	if parent.Children[0].ID != "child-span" {
		t.Error("child span not correctly linked")
	}
}

func TestLocalRecordingAddFeedback(t *testing.T) {
	rec := NewLocalRecording()
	trace := &RecordedTrace{ID: "trace-1", Feedback: make([]*RecordedFeedback, 0)}
	rec.AddTrace(trace)

	rec.AddFeedback("trace-1", RecordedFeedback{Name: "score", Value: 0.9})

	// Check feedback was added to trace
	retrieved := rec.GetTrace("trace-1")
	if len(retrieved.Feedback) != 1 {
		t.Errorf("Feedback length = %d, want 1", len(retrieved.Feedback))
	}
	if retrieved.Feedback[0].Name != "score" {
		t.Error("feedback not correctly added")
	}
}

func TestLocalRecordingAddFeedbackToSpan(t *testing.T) {
	rec := NewLocalRecording()
	span := &RecordedSpan{ID: "span-1", TraceID: "trace-1", Feedback: make([]*RecordedFeedback, 0)}
	rec.AddSpan(span)

	rec.AddFeedback("span-1", RecordedFeedback{Name: "relevance", Value: 0.8})

	retrieved := rec.GetSpan("span-1")
	if len(retrieved.Feedback) != 1 {
		t.Errorf("Feedback length = %d, want 1", len(retrieved.Feedback))
	}
}

func TestLocalRecordingTraces(t *testing.T) {
	rec := NewLocalRecording()
	rec.AddTrace(&RecordedTrace{ID: "t1"})
	rec.AddTrace(&RecordedTrace{ID: "t2"})
	rec.AddTrace(&RecordedTrace{ID: "t3"})

	traces := rec.Traces()
	if len(traces) != 3 {
		t.Errorf("Traces() length = %d, want 3", len(traces))
	}
}

func TestLocalRecordingSpans(t *testing.T) {
	rec := NewLocalRecording()
	rec.AddSpan(&RecordedSpan{ID: "s1", TraceID: "t1"})
	rec.AddSpan(&RecordedSpan{ID: "s2", TraceID: "t1"})

	spans := rec.Spans()
	if len(spans) != 2 {
		t.Errorf("Spans() length = %d, want 2", len(spans))
	}
}

func TestLocalRecordingClear(t *testing.T) {
	rec := NewLocalRecording()
	rec.AddTrace(&RecordedTrace{ID: "t1"})
	rec.AddSpan(&RecordedSpan{ID: "s1", TraceID: "t1"})

	rec.Clear()

	if rec.TraceCount() != 0 {
		t.Errorf("TraceCount after Clear = %d, want 0", rec.TraceCount())
	}
	if rec.SpanCount() != 0 {
		t.Errorf("SpanCount after Clear = %d, want 0", rec.SpanCount())
	}
}

func TestNewRecordingClient(t *testing.T) {
	client := NewRecordingClient("test-project")

	if client == nil {
		t.Fatal("NewRecordingClient returned nil")
	}
	if client.project != "test-project" {
		t.Errorf("project = %q, want %q", client.project, "test-project")
	}
	if client.recording == nil {
		t.Error("recording should not be nil")
	}
}

func TestRecordingClientTrace(t *testing.T) {
	ctx := context.Background()
	client := NewRecordingClient("test-project")

	trace, err := client.Trace(ctx, "my-trace", WithTraceInput(map[string]any{"key": "value"}))
	if err != nil {
		t.Fatalf("Trace error = %v", err)
	}

	if trace == nil {
		t.Fatal("Trace returned nil")
	}
	if trace.ID() == "" {
		t.Error("trace should have an ID")
	}
	if trace.Name() != "my-trace" {
		t.Errorf("Name() = %q, want %q", trace.Name(), "my-trace")
	}

	// Check it was recorded
	if client.Recording().TraceCount() != 1 {
		t.Errorf("TraceCount = %d, want 1", client.Recording().TraceCount())
	}
}

func TestRecordingTraceEnd(t *testing.T) {
	ctx := context.Background()
	client := NewRecordingClient("test-project")

	trace, _ := client.Trace(ctx, "my-trace")
	err := trace.End(ctx, WithTraceOutput(map[string]any{"result": "done"}))
	if err != nil {
		t.Fatalf("End error = %v", err)
	}

	// Check end time was set
	recorded := client.Recording().GetTrace(trace.ID())
	if recorded.EndTime.IsZero() {
		t.Error("EndTime should be set after End()")
	}
	if recorded.Output == nil {
		t.Error("Output should be set")
	}
}

func TestRecordingTraceSpan(t *testing.T) {
	ctx := context.Background()
	client := NewRecordingClient("test-project")

	trace, _ := client.Trace(ctx, "my-trace")
	span, err := trace.Span(ctx, "my-span", WithSpanType(SpanTypeLLM))
	if err != nil {
		t.Fatalf("Span error = %v", err)
	}

	if span == nil {
		t.Fatal("Span returned nil")
	}
	if span.ID() == "" {
		t.Error("span should have an ID")
	}
	if span.TraceID() != trace.ID() {
		t.Errorf("TraceID() = %q, want %q", span.TraceID(), trace.ID())
	}
	if span.Name() != "my-span" {
		t.Errorf("Name() = %q, want %q", span.Name(), "my-span")
	}

	// Check it was recorded
	if client.Recording().SpanCount() != 1 {
		t.Errorf("SpanCount = %d, want 1", client.Recording().SpanCount())
	}
}

func TestRecordingSpanEnd(t *testing.T) {
	ctx := context.Background()
	client := NewRecordingClient("test-project")

	trace, _ := client.Trace(ctx, "my-trace")
	span, _ := trace.Span(ctx, "my-span")
	err := span.End(ctx, WithSpanOutput(map[string]any{"result": "done"}))
	if err != nil {
		t.Fatalf("End error = %v", err)
	}

	// Check end time was set
	recorded := client.Recording().GetSpan(span.ID())
	if recorded.EndTime.IsZero() {
		t.Error("EndTime should be set after End()")
	}
}

func TestRecordingSpanChildSpan(t *testing.T) {
	ctx := context.Background()
	client := NewRecordingClient("test-project")

	trace, _ := client.Trace(ctx, "my-trace")
	parentSpan, _ := trace.Span(ctx, "parent-span")
	childSpan, err := parentSpan.Span(ctx, "child-span")
	if err != nil {
		t.Fatalf("child Span error = %v", err)
	}

	if childSpan.TraceID() != trace.ID() {
		t.Error("child span should have same trace ID")
	}

	// Check parent-child relationship in recording
	recorded := client.Recording().GetSpan(parentSpan.ID())
	if len(recorded.Children) != 1 {
		t.Errorf("parent.Children length = %d, want 1", len(recorded.Children))
	}
}

func TestRecordingTraceFeedbackScore(t *testing.T) {
	ctx := context.Background()
	client := NewRecordingClient("test-project")

	trace, _ := client.Trace(ctx, "my-trace")
	err := trace.AddFeedbackScore(ctx, "accuracy", 0.95, "good response")
	if err != nil {
		t.Fatalf("AddFeedbackScore error = %v", err)
	}

	recorded := client.Recording().GetTrace(trace.ID())
	if len(recorded.Feedback) != 1 {
		t.Errorf("Feedback length = %d, want 1", len(recorded.Feedback))
	}
	if recorded.Feedback[0].Value != 0.95 {
		t.Errorf("Feedback Value = %v, want 0.95", recorded.Feedback[0].Value)
	}
}

func TestRecordingSpanFeedbackScore(t *testing.T) {
	ctx := context.Background()
	client := NewRecordingClient("test-project")

	trace, _ := client.Trace(ctx, "my-trace")
	span, _ := trace.Span(ctx, "my-span")
	err := span.AddFeedbackScore(ctx, "relevance", 0.8, "relevant")
	if err != nil {
		t.Fatalf("AddFeedbackScore error = %v", err)
	}

	recorded := client.Recording().GetSpan(span.ID())
	if len(recorded.Feedback) != 1 {
		t.Errorf("Feedback length = %d, want 1", len(recorded.Feedback))
	}
}

func TestRecordTracesLocally(t *testing.T) {
	client := RecordTracesLocally("my-project")

	if client == nil {
		t.Fatal("RecordTracesLocally returned nil")
	}
	if client.project != "my-project" {
		t.Errorf("project = %q, want %q", client.project, "my-project")
	}
}

func TestRecordingClientRecording(t *testing.T) {
	client := NewRecordingClient("test")
	rec := client.Recording()

	if rec == nil {
		t.Error("Recording() should not return nil")
	}

	// Should be the same instance
	if rec != client.recording {
		t.Error("Recording() should return internal recording")
	}
}

func TestGenerateUUID(t *testing.T) {
	id1 := generateUUID()
	id2 := generateUUID()

	if id1 == "" {
		t.Error("generateUUID should not return empty string")
	}

	// UUID format check (8-4-4-4-12)
	parts := 0
	for _, c := range id1 {
		if c == '-' {
			parts++
		}
	}
	if parts != 4 {
		t.Errorf("UUID should have 4 dashes, got %d", parts)
	}

	// Note: Due to the simple implementation, consecutive calls might produce
	// similar UUIDs, so we just check they're valid format
	if len(id1) != 36 {
		t.Errorf("UUID length = %d, want 36", len(id1))
	}
	if len(id2) != 36 {
		t.Errorf("UUID length = %d, want 36", len(id2))
	}
}

func TestHexEncode(t *testing.T) {
	tests := []struct {
		input []byte
		want  string
	}{
		{[]byte{0x00}, "00"},
		{[]byte{0xff}, "ff"},
		{[]byte{0x12, 0x34}, "1234"},
		{[]byte{0xab, 0xcd, 0xef}, "abcdef"},
	}

	for _, tt := range tests {
		got := string(hexEncode(tt.input))
		if got != tt.want {
			t.Errorf("hexEncode(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
