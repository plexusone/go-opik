package opik

import (
	"testing"
)

func TestExperimentGetters(t *testing.T) {
	e := &Experiment{
		id:          "exp-123",
		name:        "test-experiment",
		datasetName: "my-dataset",
		metadata:    map[string]any{"key": "value"},
	}

	if e.ID() != "exp-123" {
		t.Errorf("ID() = %q, want %q", e.ID(), "exp-123")
	}
	if e.Name() != "test-experiment" {
		t.Errorf("Name() = %q, want %q", e.Name(), "test-experiment")
	}
	if e.DatasetName() != "my-dataset" {
		t.Errorf("DatasetName() = %q, want %q", e.DatasetName(), "my-dataset")
	}
	if e.Metadata()["key"] != "value" {
		t.Errorf("Metadata()[key] = %v, want %q", e.Metadata()["key"], "value")
	}
}

func TestExperimentItemFields(t *testing.T) {
	item := ExperimentItem{
		ID:            "item-123",
		ExperimentID:  "exp-456",
		DatasetItemID: "ds-item-789",
		TraceID:       "trace-abc",
		Input:         map[string]any{"prompt": "hello"},
		Output:        map[string]any{"response": "world"},
	}

	if item.ID != "item-123" {
		t.Errorf("ID = %q, want %q", item.ID, "item-123")
	}
	if item.ExperimentID != "exp-456" {
		t.Errorf("ExperimentID = %q, want %q", item.ExperimentID, "exp-456")
	}
	if item.DatasetItemID != "ds-item-789" {
		t.Errorf("DatasetItemID = %q, want %q", item.DatasetItemID, "ds-item-789")
	}
	if item.TraceID != "trace-abc" {
		t.Errorf("TraceID = %q, want %q", item.TraceID, "trace-abc")
	}
}

func TestExperimentStatusConstants(t *testing.T) {
	tests := []struct {
		status ExperimentStatus
		want   string
	}{
		{ExperimentStatusRunning, "running"},
		{ExperimentStatusCompleted, "completed"},
		{ExperimentStatusCancelled, "cancelled"},
	}

	for _, tt := range tests {
		if string(tt.status) != tt.want {
			t.Errorf("ExperimentStatus = %q, want %q", tt.status, tt.want)
		}
	}
}

func TestExperimentTypeConstants(t *testing.T) {
	tests := []struct {
		expType ExperimentType
		want    string
	}{
		{ExperimentTypeRegular, "regular"},
		{ExperimentTypeTrial, "trial"},
		{ExperimentTypeMiniBatch, "mini-batch"},
	}

	for _, tt := range tests {
		if string(tt.expType) != tt.want {
			t.Errorf("ExperimentType = %q, want %q", tt.expType, tt.want)
		}
	}
}

func TestExperimentOptions(t *testing.T) {
	t.Run("WithExperimentName", func(t *testing.T) {
		opts := &experimentOptions{}
		WithExperimentName("my-experiment")(opts)
		if opts.name != "my-experiment" {
			t.Errorf("name = %q, want %q", opts.name, "my-experiment")
		}
	})

	t.Run("WithExperimentMetadata", func(t *testing.T) {
		opts := &experimentOptions{}
		metadata := map[string]any{"model": "gpt-4", "version": 2}
		WithExperimentMetadata(metadata)(opts)
		if opts.metadata["model"] != "gpt-4" {
			t.Errorf("metadata[model] = %v, want %q", opts.metadata["model"], "gpt-4")
		}
		if opts.metadata["version"] != 2 {
			t.Errorf("metadata[version] = %v, want 2", opts.metadata["version"])
		}
	})

	t.Run("WithExperimentType", func(t *testing.T) {
		opts := &experimentOptions{}
		WithExperimentType(ExperimentTypeTrial)(opts)
		if opts.experimentType != ExperimentTypeTrial {
			t.Errorf("experimentType = %q, want %q", opts.experimentType, ExperimentTypeTrial)
		}
	})

	t.Run("WithExperimentStatus", func(t *testing.T) {
		opts := &experimentOptions{}
		WithExperimentStatus(ExperimentStatusCompleted)(opts)
		if opts.status != ExperimentStatusCompleted {
			t.Errorf("status = %q, want %q", opts.status, ExperimentStatusCompleted)
		}
	})
}

func TestExperimentItemOptions(t *testing.T) {
	t.Run("WithExperimentItemInput", func(t *testing.T) {
		opts := &experimentItemOptions{}
		input := map[string]any{"prompt": "test prompt"}
		WithExperimentItemInput(input)(opts)
		if inputMap, ok := opts.input.(map[string]any); !ok {
			t.Error("input should be map[string]any")
		} else if inputMap["prompt"] != "test prompt" {
			t.Errorf("input[prompt] = %v, want %q", inputMap["prompt"], "test prompt")
		}
	})

	t.Run("WithExperimentItemOutput", func(t *testing.T) {
		opts := &experimentItemOptions{}
		output := map[string]any{"response": "test response"}
		WithExperimentItemOutput(output)(opts)
		if outputMap, ok := opts.output.(map[string]any); !ok {
			t.Error("output should be map[string]any")
		} else if outputMap["response"] != "test response" {
			t.Errorf("output[response] = %v, want %q", outputMap["response"], "test response")
		}
	})
}

func TestExperimentEmptyState(t *testing.T) {
	e := &Experiment{}

	if e.ID() != "" {
		t.Errorf("ID() = %q, want empty", e.ID())
	}
	if e.Name() != "" {
		t.Errorf("Name() = %q, want empty", e.Name())
	}
	if e.DatasetName() != "" {
		t.Errorf("DatasetName() = %q, want empty", e.DatasetName())
	}
	if e.Metadata() != nil {
		t.Errorf("Metadata() = %v, want nil", e.Metadata())
	}
}

func TestExperimentOptionsChaining(t *testing.T) {
	// Test that multiple options can be applied together
	opts := &experimentOptions{}

	// Apply multiple options
	options := []ExperimentOption{
		WithExperimentName("chained-experiment"),
		WithExperimentMetadata(map[string]any{"chain": true}),
		WithExperimentType(ExperimentTypeMiniBatch),
		WithExperimentStatus(ExperimentStatusRunning),
	}

	for _, opt := range options {
		opt(opts)
	}

	if opts.name != "chained-experiment" {
		t.Errorf("name = %q, want %q", opts.name, "chained-experiment")
	}
	if opts.metadata["chain"] != true {
		t.Errorf("metadata[chain] = %v, want true", opts.metadata["chain"])
	}
	if opts.experimentType != ExperimentTypeMiniBatch {
		t.Errorf("experimentType = %q, want %q", opts.experimentType, ExperimentTypeMiniBatch)
	}
	if opts.status != ExperimentStatusRunning {
		t.Errorf("status = %q, want %q", opts.status, ExperimentStatusRunning)
	}
}
