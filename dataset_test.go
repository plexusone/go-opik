package opik

import (
	"testing"
)

func TestDatasetGetters(t *testing.T) {
	d := &Dataset{
		id:          "ds-123",
		name:        "test-dataset",
		description: "A test dataset",
		tags:        []string{"tag1", "tag2"},
	}

	if d.ID() != "ds-123" {
		t.Errorf("ID() = %q, want %q", d.ID(), "ds-123")
	}
	if d.Name() != "test-dataset" {
		t.Errorf("Name() = %q, want %q", d.Name(), "test-dataset")
	}
	if d.Description() != "A test dataset" {
		t.Errorf("Description() = %q, want %q", d.Description(), "A test dataset")
	}
	if len(d.Tags()) != 2 || d.Tags()[0] != "tag1" {
		t.Errorf("Tags() = %v, want [tag1, tag2]", d.Tags())
	}
}

func TestDatasetItemFields(t *testing.T) {
	item := DatasetItem{
		ID:      "item-123",
		TraceID: "trace-456",
		SpanID:  "span-789",
		Data:    map[string]any{"key": "value"},
		Tags:    []string{"item-tag"},
	}

	if item.ID != "item-123" {
		t.Errorf("ID = %q, want %q", item.ID, "item-123")
	}
	if item.TraceID != "trace-456" {
		t.Errorf("TraceID = %q, want %q", item.TraceID, "trace-456")
	}
	if item.SpanID != "span-789" {
		t.Errorf("SpanID = %q, want %q", item.SpanID, "span-789")
	}
	if item.Data["key"] != "value" {
		t.Errorf("Data[key] = %v, want %q", item.Data["key"], "value")
	}
	if len(item.Tags) != 1 || item.Tags[0] != "item-tag" {
		t.Errorf("Tags = %v, want [item-tag]", item.Tags)
	}
}

func TestDatasetOptions(t *testing.T) {
	t.Run("WithDatasetDescription", func(t *testing.T) {
		opts := &datasetOptions{}
		WithDatasetDescription("my description")(opts)
		if opts.description != "my description" {
			t.Errorf("description = %q, want %q", opts.description, "my description")
		}
	})

	t.Run("WithDatasetTags", func(t *testing.T) {
		opts := &datasetOptions{}
		WithDatasetTags("a", "b", "c")(opts)
		if len(opts.tags) != 3 {
			t.Errorf("tags length = %d, want 3", len(opts.tags))
		}
		if opts.tags[0] != "a" || opts.tags[1] != "b" || opts.tags[2] != "c" {
			t.Errorf("tags = %v, want [a, b, c]", opts.tags)
		}
	})
}

func TestDatasetItemOptions(t *testing.T) {
	t.Run("WithDatasetItemTags", func(t *testing.T) {
		opts := &datasetItemOptions{}
		WithDatasetItemTags("x", "y")(opts)
		if len(opts.tags) != 2 {
			t.Errorf("tags length = %d, want 2", len(opts.tags))
		}
		if opts.tags[0] != "x" || opts.tags[1] != "y" {
			t.Errorf("tags = %v, want [x, y]", opts.tags)
		}
	})
}

func TestMapToJsonNode(t *testing.T) {
	t.Run("nil map", func(t *testing.T) {
		result := mapToJsonNode(nil)
		if result != nil {
			t.Error("expected nil for nil map")
		}
	})

	t.Run("empty map", func(t *testing.T) {
		result := mapToJsonNode(map[string]any{})
		if len(result) != 0 {
			t.Errorf("expected empty map, got %v", result)
		}
	})

	t.Run("with values", func(t *testing.T) {
		input := map[string]any{
			"string": "hello",
			"number": 42,
			"bool":   true,
		}
		result := mapToJsonNode(input)

		if len(result) != 3 {
			t.Errorf("expected 3 entries, got %d", len(result))
		}

		// Check that keys exist (values are serialized JSON)
		if _, ok := result["string"]; !ok {
			t.Error("missing 'string' key")
		}
		if _, ok := result["number"]; !ok {
			t.Error("missing 'number' key")
		}
		if _, ok := result["bool"]; !ok {
			t.Error("missing 'bool' key")
		}
	})
}

func TestJsonNodeToMap(t *testing.T) {
	t.Run("nil node", func(t *testing.T) {
		result := jsonNodeToMap(nil)
		if result != nil {
			t.Error("expected nil for nil node")
		}
	})

	t.Run("round trip", func(t *testing.T) {
		original := map[string]any{
			"key": "value",
		}
		node := mapToJsonNode(original)
		result := jsonNodeToMap(node)

		if result["key"] != "value" {
			t.Errorf("result[key] = %v, want %q", result["key"], "value")
		}
	})

	t.Run("complex types", func(t *testing.T) {
		original := map[string]any{
			"array":  []any{1, 2, 3},
			"nested": map[string]any{"inner": "data"},
		}
		node := mapToJsonNode(original)
		result := jsonNodeToMap(node)

		// Check array was preserved (as []any after JSON round-trip)
		if arr, ok := result["array"].([]any); !ok || len(arr) != 3 {
			t.Errorf("result[array] = %v, expected array of length 3", result["array"])
		}

		// Check nested object was preserved
		if nested, ok := result["nested"].(map[string]any); !ok {
			t.Errorf("result[nested] = %v, expected map", result["nested"])
		} else if nested["inner"] != "data" {
			t.Errorf("result[nested][inner] = %v, want %q", nested["inner"], "data")
		}
	})
}

func TestDatasetEmptyState(t *testing.T) {
	d := &Dataset{}

	if d.ID() != "" {
		t.Errorf("ID() = %q, want empty", d.ID())
	}
	if d.Name() != "" {
		t.Errorf("Name() = %q, want empty", d.Name())
	}
	if d.Description() != "" {
		t.Errorf("Description() = %q, want empty", d.Description())
	}
	if d.Tags() != nil {
		t.Errorf("Tags() = %v, want nil", d.Tags())
	}
}
