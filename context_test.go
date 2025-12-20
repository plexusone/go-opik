package opik

import (
	"context"
	"testing"
	"time"
)

func TestContextWithTrace(t *testing.T) {
	ctx := context.Background()
	trace := &Trace{
		id:        "test-trace-id",
		name:      "test-trace",
		startTime: time.Now(),
	}

	newCtx := ContextWithTrace(ctx, trace)

	// Verify trace is in new context
	retrievedTrace := TraceFromContext(newCtx)
	if retrievedTrace == nil {
		t.Fatal("TraceFromContext returned nil")
	}
	if retrievedTrace.ID() != trace.ID() {
		t.Errorf("TraceFromContext ID = %q, want %q", retrievedTrace.ID(), trace.ID())
	}
	if retrievedTrace.Name() != trace.Name() {
		t.Errorf("TraceFromContext Name = %q, want %q", retrievedTrace.Name(), trace.Name())
	}
}

func TestTraceFromContextEmpty(t *testing.T) {
	ctx := context.Background()
	trace := TraceFromContext(ctx)
	if trace != nil {
		t.Errorf("TraceFromContext on empty context = %v, want nil", trace)
	}
}

func TestContextWithSpan(t *testing.T) {
	ctx := context.Background()
	span := &Span{
		id:        "test-span-id",
		traceID:   "test-trace-id",
		name:      "test-span",
		startTime: time.Now(),
	}

	newCtx := ContextWithSpan(ctx, span)

	// Verify span is in new context
	retrievedSpan := SpanFromContext(newCtx)
	if retrievedSpan == nil {
		t.Fatal("SpanFromContext returned nil")
	}
	if retrievedSpan.ID() != span.ID() {
		t.Errorf("SpanFromContext ID = %q, want %q", retrievedSpan.ID(), span.ID())
	}
	if retrievedSpan.TraceID() != span.TraceID() {
		t.Errorf("SpanFromContext TraceID = %q, want %q", retrievedSpan.TraceID(), span.TraceID())
	}
	if retrievedSpan.Name() != span.Name() {
		t.Errorf("SpanFromContext Name = %q, want %q", retrievedSpan.Name(), span.Name())
	}
}

func TestSpanFromContextEmpty(t *testing.T) {
	ctx := context.Background()
	span := SpanFromContext(ctx)
	if span != nil {
		t.Errorf("SpanFromContext on empty context = %v, want nil", span)
	}
}

func TestContextWithClient(t *testing.T) {
	ctx := context.Background()
	client := &Client{
		config: &Config{
			URL:       "http://localhost:5173/api",
			Workspace: "test",
		},
	}

	newCtx := ContextWithClient(ctx, client)

	// Verify client is in new context
	retrievedClient := ClientFromContext(newCtx)
	if retrievedClient == nil {
		t.Fatal("ClientFromContext returned nil")
	}
	if retrievedClient.config.URL != client.config.URL {
		t.Errorf("ClientFromContext URL = %q, want %q", retrievedClient.config.URL, client.config.URL)
	}
}

func TestClientFromContextEmpty(t *testing.T) {
	ctx := context.Background()
	client := ClientFromContext(ctx)
	if client != nil {
		t.Errorf("ClientFromContext on empty context = %v, want nil", client)
	}
}

func TestContextChaining(t *testing.T) {
	ctx := context.Background()

	trace := &Trace{
		id:        "trace-1",
		name:      "my-trace",
		startTime: time.Now(),
	}

	span := &Span{
		id:        "span-1",
		traceID:   "trace-1",
		name:      "my-span",
		startTime: time.Now(),
	}

	client := &Client{
		config: &Config{URL: "http://localhost:5173/api"},
	}

	// Chain all context values
	ctx = ContextWithTrace(ctx, trace)
	ctx = ContextWithSpan(ctx, span)
	ctx = ContextWithClient(ctx, client)

	// Verify all values are accessible
	if TraceFromContext(ctx) == nil {
		t.Error("TraceFromContext returned nil after chaining")
	}
	if SpanFromContext(ctx) == nil {
		t.Error("SpanFromContext returned nil after chaining")
	}
	if ClientFromContext(ctx) == nil {
		t.Error("ClientFromContext returned nil after chaining")
	}
}

func TestContextOverwrite(t *testing.T) {
	ctx := context.Background()

	trace1 := &Trace{id: "trace-1", startTime: time.Now()}
	trace2 := &Trace{id: "trace-2", startTime: time.Now()}

	ctx = ContextWithTrace(ctx, trace1)
	ctx = ContextWithTrace(ctx, trace2)

	retrieved := TraceFromContext(ctx)
	if retrieved.ID() != "trace-2" {
		t.Errorf("TraceFromContext after overwrite = %q, want %q", retrieved.ID(), "trace-2")
	}
}

func TestCurrentTraceID(t *testing.T) {
	t.Run("from trace", func(t *testing.T) {
		ctx := context.Background()
		trace := &Trace{id: "trace-123", startTime: time.Now()}
		ctx = ContextWithTrace(ctx, trace)

		traceID := CurrentTraceID(ctx)
		if traceID != "trace-123" {
			t.Errorf("CurrentTraceID = %q, want %q", traceID, "trace-123")
		}
	})

	t.Run("from span", func(t *testing.T) {
		ctx := context.Background()
		span := &Span{id: "span-123", traceID: "trace-456", startTime: time.Now()}
		ctx = ContextWithSpan(ctx, span)

		traceID := CurrentTraceID(ctx)
		if traceID != "trace-456" {
			t.Errorf("CurrentTraceID = %q, want %q", traceID, "trace-456")
		}
	})

	t.Run("trace takes precedence", func(t *testing.T) {
		ctx := context.Background()
		trace := &Trace{id: "trace-primary", startTime: time.Now()}
		span := &Span{id: "span-123", traceID: "trace-secondary", startTime: time.Now()}
		ctx = ContextWithTrace(ctx, trace)
		ctx = ContextWithSpan(ctx, span)

		traceID := CurrentTraceID(ctx)
		if traceID != "trace-primary" {
			t.Errorf("CurrentTraceID = %q, want %q", traceID, "trace-primary")
		}
	})

	t.Run("empty context", func(t *testing.T) {
		ctx := context.Background()
		traceID := CurrentTraceID(ctx)
		if traceID != "" {
			t.Errorf("CurrentTraceID on empty context = %q, want empty", traceID)
		}
	})
}

func TestCurrentSpanID(t *testing.T) {
	t.Run("from span", func(t *testing.T) {
		ctx := context.Background()
		span := &Span{id: "span-123", startTime: time.Now()}
		ctx = ContextWithSpan(ctx, span)

		spanID := CurrentSpanID(ctx)
		if spanID != "span-123" {
			t.Errorf("CurrentSpanID = %q, want %q", spanID, "span-123")
		}
	})

	t.Run("empty context", func(t *testing.T) {
		ctx := context.Background()
		spanID := CurrentSpanID(ctx)
		if spanID != "" {
			t.Errorf("CurrentSpanID on empty context = %q, want empty", spanID)
		}
	})
}

func TestEndTraceNoActiveTrace(t *testing.T) {
	ctx := context.Background()
	err := EndTrace(ctx)
	if err != ErrNoActiveTrace {
		t.Errorf("EndTrace on empty context = %v, want ErrNoActiveTrace", err)
	}
}

func TestEndSpanNoActiveSpan(t *testing.T) {
	ctx := context.Background()
	err := EndSpan(ctx)
	if err != ErrNoActiveSpan {
		t.Errorf("EndSpan on empty context = %v, want ErrNoActiveSpan", err)
	}
}

func TestStartSpanNoActiveTrace(t *testing.T) {
	ctx := context.Background()
	_, span, err := StartSpan(ctx, "test-span")
	if err != ErrNoActiveTrace {
		t.Errorf("StartSpan without trace = %v, want ErrNoActiveTrace", err)
	}
	if span != nil {
		t.Errorf("StartSpan without trace returned span = %v, want nil", span)
	}
}

func TestContextKeyType(t *testing.T) {
	// Ensure context keys are unique and don't collide with other packages
	ctx := context.Background()

	// Add values with our keys
	ctx = ContextWithTrace(ctx, &Trace{id: "test", startTime: time.Now()})
	ctx = ContextWithSpan(ctx, &Span{id: "test", startTime: time.Now()})
	ctx = ContextWithClient(ctx, &Client{config: &Config{}})

	// Add a value with a different int key (simulating another package)
	type otherKey int
	ctx = context.WithValue(ctx, otherKey(0), "other-value")

	// Our values should still be accessible
	if TraceFromContext(ctx) == nil {
		t.Error("TraceFromContext returned nil with other key present")
	}
	if SpanFromContext(ctx) == nil {
		t.Error("SpanFromContext returned nil with other key present")
	}
	if ClientFromContext(ctx) == nil {
		t.Error("ClientFromContext returned nil with other key present")
	}
}

func TestNilValues(t *testing.T) {
	t.Run("nil trace", func(t *testing.T) {
		ctx := context.Background()
		ctx = ContextWithTrace(ctx, nil)
		trace := TraceFromContext(ctx)
		if trace != nil {
			t.Error("TraceFromContext should return nil for nil trace")
		}
	})

	t.Run("nil span", func(t *testing.T) {
		ctx := context.Background()
		ctx = ContextWithSpan(ctx, nil)
		span := SpanFromContext(ctx)
		if span != nil {
			t.Error("SpanFromContext should return nil for nil span")
		}
	})

	t.Run("nil client", func(t *testing.T) {
		ctx := context.Background()
		ctx = ContextWithClient(ctx, nil)
		client := ClientFromContext(ctx)
		if client != nil {
			t.Error("ClientFromContext should return nil for nil client")
		}
	})
}

func TestContextWithCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	trace := &Trace{id: "test-trace", startTime: time.Now()}
	ctx = ContextWithTrace(ctx, trace)

	// Trace should be accessible before cancellation
	if TraceFromContext(ctx) == nil {
		t.Error("TraceFromContext returned nil before cancellation")
	}

	cancel()

	// Trace should still be accessible after cancellation
	// (context values are not affected by cancellation)
	if TraceFromContext(ctx) == nil {
		t.Error("TraceFromContext returned nil after cancellation")
	}
}
