package opik

import (
	"testing"
	"time"
)

func TestTraceID(t *testing.T) {
	trace := &Trace{
		id: "test-trace-id-123",
	}

	if trace.ID() != "test-trace-id-123" {
		t.Errorf("ID() = %q, want %q", trace.ID(), "test-trace-id-123")
	}
}

func TestTraceName(t *testing.T) {
	trace := &Trace{
		name: "my-trace",
	}

	if trace.Name() != "my-trace" {
		t.Errorf("Name() = %q, want %q", trace.Name(), "my-trace")
	}
}

func TestTraceProjectName(t *testing.T) {
	trace := &Trace{
		projectName: "my-project",
	}

	if trace.ProjectName() != "my-project" {
		t.Errorf("ProjectName() = %q, want %q", trace.ProjectName(), "my-project")
	}
}

func TestTraceStartTime(t *testing.T) {
	now := time.Now()
	trace := &Trace{
		startTime: now,
	}

	if !trace.StartTime().Equal(now) {
		t.Errorf("StartTime() = %v, want %v", trace.StartTime(), now)
	}
}

func TestTraceEndTime(t *testing.T) {
	t.Run("not ended", func(t *testing.T) {
		trace := &Trace{}
		if trace.EndTime() != nil {
			t.Error("EndTime() should be nil for unended trace")
		}
	})

	t.Run("ended", func(t *testing.T) {
		endTime := time.Now()
		trace := &Trace{
			endTime: &endTime,
		}
		if trace.EndTime() == nil {
			t.Error("EndTime() should not be nil for ended trace")
		}
		if !trace.EndTime().Equal(endTime) {
			t.Errorf("EndTime() = %v, want %v", trace.EndTime(), endTime)
		}
	})
}

func TestTraceFields(t *testing.T) {
	now := time.Now()
	trace := &Trace{
		id:          "trace-123",
		name:        "test-trace",
		projectName: "test-project",
		startTime:   now,
		input:       map[string]any{"prompt": "hello"},
		output:      map[string]any{"response": "world"},
		metadata:    map[string]any{"user": "test"},
		tags:        []string{"tag1", "tag2"},
		ended:       false,
	}

	if trace.ID() != "trace-123" {
		t.Errorf("ID() = %q, want %q", trace.ID(), "trace-123")
	}
	if trace.Name() != "test-trace" {
		t.Errorf("Name() = %q, want %q", trace.Name(), "test-trace")
	}
	if trace.ProjectName() != "test-project" {
		t.Errorf("ProjectName() = %q, want %q", trace.ProjectName(), "test-project")
	}
	if !trace.StartTime().Equal(now) {
		t.Errorf("StartTime() = %v, want %v", trace.StartTime(), now)
	}
	if trace.EndTime() != nil {
		t.Error("EndTime() should be nil")
	}
}

func TestTraceEnded(t *testing.T) {
	t.Run("initially not ended", func(t *testing.T) {
		trace := &Trace{}
		if trace.ended {
			t.Error("trace should not be ended initially")
		}
	})

	t.Run("ended flag set", func(t *testing.T) {
		trace := &Trace{ended: true}
		if !trace.ended {
			t.Error("trace should be ended")
		}
	})
}

func TestTraceWithAllFields(t *testing.T) {
	now := time.Now()
	endTime := now.Add(5 * time.Second)

	trace := &Trace{
		id:          "full-trace-id",
		name:        "full-trace",
		projectName: "full-project",
		startTime:   now,
		endTime:     &endTime,
		input:       map[string]any{"query": "test query"},
		output:      map[string]any{"result": "test result"},
		metadata: map[string]any{
			"model":       "gpt-4",
			"temperature": 0.7,
		},
		tags:  []string{"production", "v1.0"},
		ended: true,
	}

	// Test all getters
	if trace.ID() != "full-trace-id" {
		t.Errorf("ID() = %q, want %q", trace.ID(), "full-trace-id")
	}
	if trace.Name() != "full-trace" {
		t.Errorf("Name() = %q, want %q", trace.Name(), "full-trace")
	}
	if trace.ProjectName() != "full-project" {
		t.Errorf("ProjectName() = %q, want %q", trace.ProjectName(), "full-project")
	}
	if !trace.StartTime().Equal(now) {
		t.Errorf("StartTime() = %v, want %v", trace.StartTime(), now)
	}
	if trace.EndTime() == nil || !trace.EndTime().Equal(endTime) {
		t.Errorf("EndTime() = %v, want %v", trace.EndTime(), endTime)
	}
}

func TestTraceEmptyFields(t *testing.T) {
	trace := &Trace{}

	if trace.ID() != "" {
		t.Errorf("ID() = %q, want empty", trace.ID())
	}
	if trace.Name() != "" {
		t.Errorf("Name() = %q, want empty", trace.Name())
	}
	if trace.ProjectName() != "" {
		t.Errorf("ProjectName() = %q, want empty", trace.ProjectName())
	}
	if !trace.StartTime().IsZero() {
		t.Errorf("StartTime() = %v, want zero", trace.StartTime())
	}
	if trace.EndTime() != nil {
		t.Errorf("EndTime() = %v, want nil", trace.EndTime())
	}
}

func TestTraceInputOutput(t *testing.T) {
	trace := &Trace{
		input: map[string]any{
			"messages": []map[string]string{
				{"role": "user", "content": "Hello"},
			},
		},
		output: map[string]any{
			"message": map[string]string{
				"role":    "assistant",
				"content": "Hi there!",
			},
		},
	}

	// Test that input/output are stored correctly
	if trace.input == nil {
		t.Error("input should not be nil")
	}
	if trace.output == nil {
		t.Error("output should not be nil")
	}
}

func TestTraceMetadata(t *testing.T) {
	trace := &Trace{
		metadata: map[string]any{
			"user_id":    "user-123",
			"session_id": "session-456",
			"tokens":     150,
			"cost":       0.003,
			"is_cached":  false,
			"nested":     map[string]any{"key": "value"},
		},
	}

	if trace.metadata == nil {
		t.Error("metadata should not be nil")
	}
	if trace.metadata["user_id"] != "user-123" {
		t.Errorf("metadata[user_id] = %v, want %q", trace.metadata["user_id"], "user-123")
	}
	if trace.metadata["tokens"] != 150 {
		t.Errorf("metadata[tokens] = %v, want 150", trace.metadata["tokens"])
	}
}

func TestTraceTags(t *testing.T) {
	trace := &Trace{
		tags: []string{"production", "high-priority", "gpt-4"},
	}

	if len(trace.tags) != 3 {
		t.Errorf("tags length = %d, want 3", len(trace.tags))
	}
	if trace.tags[0] != "production" {
		t.Errorf("tags[0] = %q, want %q", trace.tags[0], "production")
	}
}

func TestTraceEmptyTags(t *testing.T) {
	trace := &Trace{
		tags: []string{},
	}

	if trace.tags == nil {
		t.Error("tags should not be nil")
	}
	if len(trace.tags) != 0 {
		t.Errorf("tags length = %d, want 0", len(trace.tags))
	}
}

func TestTraceNilClient(t *testing.T) {
	// Test that a trace with nil client doesn't panic on getters
	trace := &Trace{
		id:   "test-id",
		name: "test-name",
	}

	// These should not panic
	_ = trace.ID()
	_ = trace.Name()
	_ = trace.ProjectName()
	_ = trace.StartTime()
	_ = trace.EndTime()
}
