package opik

import (
	"testing"
)

func TestPromptGetters(t *testing.T) {
	p := &Prompt{
		id:          "prompt-123",
		name:        "test-prompt",
		description: "A test prompt",
		template:    "Hello, {{name}}!",
		tags:        []string{"greeting", "test"},
	}

	if p.ID() != "prompt-123" {
		t.Errorf("ID() = %q, want %q", p.ID(), "prompt-123")
	}
	if p.Name() != "test-prompt" {
		t.Errorf("Name() = %q, want %q", p.Name(), "test-prompt")
	}
	if p.Description() != "A test prompt" {
		t.Errorf("Description() = %q, want %q", p.Description(), "A test prompt")
	}
	if p.Template() != "Hello, {{name}}!" {
		t.Errorf("Template() = %q, want %q", p.Template(), "Hello, {{name}}!")
	}
	if len(p.Tags()) != 2 || p.Tags()[0] != "greeting" {
		t.Errorf("Tags() = %v, want [greeting, test]", p.Tags())
	}
}

func TestPromptVersionGetters(t *testing.T) {
	v := &PromptVersion{
		id:                "version-456",
		promptID:          "prompt-123",
		commit:            "abc123",
		template:          "Hi, {{user}}!",
		changeDescription: "Updated greeting",
		tags:              []string{"v2"},
	}

	if v.ID() != "version-456" {
		t.Errorf("ID() = %q, want %q", v.ID(), "version-456")
	}
	if v.PromptID() != "prompt-123" {
		t.Errorf("PromptID() = %q, want %q", v.PromptID(), "prompt-123")
	}
	if v.Commit() != "abc123" {
		t.Errorf("Commit() = %q, want %q", v.Commit(), "abc123")
	}
	if v.Template() != "Hi, {{user}}!" {
		t.Errorf("Template() = %q, want %q", v.Template(), "Hi, {{user}}!")
	}
	if v.ChangeDescription() != "Updated greeting" {
		t.Errorf("ChangeDescription() = %q, want %q", v.ChangeDescription(), "Updated greeting")
	}
	if len(v.Tags()) != 1 || v.Tags()[0] != "v2" {
		t.Errorf("Tags() = %v, want [v2]", v.Tags())
	}
}

func TestPromptTemplateStructureConstants(t *testing.T) {
	tests := []struct {
		structure PromptTemplateStructure
		want      string
	}{
		{PromptTemplateStructureText, "text"},
		{PromptTemplateStructureChat, "chat"},
	}

	for _, tt := range tests {
		if string(tt.structure) != tt.want {
			t.Errorf("PromptTemplateStructure = %q, want %q", tt.structure, tt.want)
		}
	}
}

func TestPromptTypeConstants(t *testing.T) {
	tests := []struct {
		promptType PromptType
		want       string
	}{
		{PromptTypeMustache, "mustache"},
		{PromptTypeFString, "fstring"},
		{PromptTypeJinja2, "jinja2"},
	}

	for _, tt := range tests {
		if string(tt.promptType) != tt.want {
			t.Errorf("PromptType = %q, want %q", tt.promptType, tt.want)
		}
	}
}

func TestPromptOptions(t *testing.T) {
	t.Run("WithPromptDescription", func(t *testing.T) {
		opts := &promptOptions{}
		WithPromptDescription("my description")(opts)
		if opts.description != "my description" {
			t.Errorf("description = %q, want %q", opts.description, "my description")
		}
	})

	t.Run("WithPromptTemplate", func(t *testing.T) {
		opts := &promptOptions{}
		WithPromptTemplate("Hello {{world}}")(opts)
		if opts.template != "Hello {{world}}" {
			t.Errorf("template = %q, want %q", opts.template, "Hello {{world}}")
		}
	})

	t.Run("WithPromptChangeDescription", func(t *testing.T) {
		opts := &promptOptions{}
		WithPromptChangeDescription("Initial version")(opts)
		if opts.changeDescription != "Initial version" {
			t.Errorf("changeDescription = %q, want %q", opts.changeDescription, "Initial version")
		}
	})

	t.Run("WithPromptType", func(t *testing.T) {
		opts := &promptOptions{}
		WithPromptType(PromptTypeJinja2)(opts)
		if opts.promptType != PromptTypeJinja2 {
			t.Errorf("promptType = %q, want %q", opts.promptType, PromptTypeJinja2)
		}
	})

	t.Run("WithPromptTemplateStructure", func(t *testing.T) {
		opts := &promptOptions{}
		WithPromptTemplateStructure(PromptTemplateStructureChat)(opts)
		if opts.templateStructure != PromptTemplateStructureChat {
			t.Errorf("templateStructure = %q, want %q", opts.templateStructure, PromptTemplateStructureChat)
		}
	})

	t.Run("WithPromptTags", func(t *testing.T) {
		opts := &promptOptions{}
		WithPromptTags("a", "b", "c")(opts)
		if len(opts.tags) != 3 {
			t.Errorf("tags length = %d, want 3", len(opts.tags))
		}
	})
}

func TestPromptVersionOptions(t *testing.T) {
	t.Run("WithVersionChangeDescription", func(t *testing.T) {
		opts := &promptVersionOptions{}
		WithVersionChangeDescription("Updated template")(opts)
		if opts.changeDescription != "Updated template" {
			t.Errorf("changeDescription = %q, want %q", opts.changeDescription, "Updated template")
		}
	})

	t.Run("WithVersionType", func(t *testing.T) {
		opts := &promptVersionOptions{}
		WithVersionType(PromptTypeFString)(opts)
		if opts.promptType != PromptTypeFString {
			t.Errorf("promptType = %q, want %q", opts.promptType, PromptTypeFString)
		}
	})

	t.Run("WithVersionTags", func(t *testing.T) {
		opts := &promptVersionOptions{}
		WithVersionTags("x", "y")(opts)
		if len(opts.tags) != 2 {
			t.Errorf("tags length = %d, want 2", len(opts.tags))
		}
	})
}

func TestPromptVersionRender(t *testing.T) {
	tests := []struct {
		name      string
		template  string
		variables map[string]string
		want      string
	}{
		{
			name:      "simple substitution",
			template:  "Hello, {{name}}!",
			variables: map[string]string{"name": "World"},
			want:      "Hello, World!",
		},
		{
			name:      "multiple variables",
			template:  "{{greeting}}, {{name}}! Welcome to {{place}}.",
			variables: map[string]string{"greeting": "Hi", "name": "John", "place": "Go"},
			want:      "Hi, John! Welcome to Go.",
		},
		{
			name:      "no variables",
			template:  "Just plain text",
			variables: map[string]string{},
			want:      "Just plain text",
		},
		{
			name:      "missing variable (kept as is)",
			template:  "Hello, {{name}}!",
			variables: map[string]string{},
			want:      "Hello, {{name}}!",
		},
		{
			name:      "repeated variable",
			template:  "{{word}} {{word}} {{word}}",
			variables: map[string]string{"word": "echo"},
			want:      "echo echo echo",
		},
		{
			name:      "nil variables",
			template:  "Hello {{name}}",
			variables: nil,
			want:      "Hello {{name}}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &PromptVersion{template: tt.template}
			got := v.Render(tt.variables)
			if got != tt.want {
				t.Errorf("Render() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPromptVersionRenderWithDefault(t *testing.T) {
	tests := []struct {
		name         string
		template     string
		variables    map[string]string
		defaultValue string
		want         string
	}{
		{
			name:         "all variables provided",
			template:     "Hello, {{name}}!",
			variables:    map[string]string{"name": "World"},
			defaultValue: "[MISSING]",
			want:         "Hello, World!",
		},
		{
			name:         "missing variable replaced with default",
			template:     "Hello, {{name}}!",
			variables:    map[string]string{},
			defaultValue: "[UNKNOWN]",
			want:         "Hello, [UNKNOWN]!",
		},
		{
			name:         "partial variables",
			template:     "{{greeting}}, {{name}}!",
			variables:    map[string]string{"greeting": "Hi"},
			defaultValue: "???",
			want:         "Hi, ???!",
		},
		{
			name:         "empty default",
			template:     "Hello, {{name}}!",
			variables:    map[string]string{},
			defaultValue: "",
			want:         "Hello, !",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &PromptVersion{template: tt.template}
			got := v.RenderWithDefault(tt.variables, tt.defaultValue)
			if got != tt.want {
				t.Errorf("RenderWithDefault() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPromptVersionExtractVariables(t *testing.T) {
	tests := []struct {
		name     string
		template string
		want     []string
	}{
		{
			name:     "single variable",
			template: "Hello, {{name}}!",
			want:     []string{"name"},
		},
		{
			name:     "multiple variables",
			template: "{{greeting}}, {{name}}! Welcome to {{place}}.",
			want:     []string{"greeting", "name", "place"},
		},
		{
			name:     "no variables",
			template: "Just plain text",
			want:     nil,
		},
		{
			name:     "duplicate variables (deduped)",
			template: "{{word}} {{word}} {{other}}",
			want:     []string{"word", "other"},
		},
		{
			name:     "variables with spaces (trimmed)",
			template: "Hello, {{ name }}!",
			want:     []string{"name"},
		},
		{
			name:     "empty template",
			template: "",
			want:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &PromptVersion{template: tt.template}
			got := v.ExtractVariables()

			if len(got) != len(tt.want) {
				t.Errorf("ExtractVariables() length = %d, want %d", len(got), len(tt.want))
				return
			}

			for i, want := range tt.want {
				if got[i] != want {
					t.Errorf("ExtractVariables()[%d] = %q, want %q", i, got[i], want)
				}
			}
		})
	}
}

func TestPromptEmptyState(t *testing.T) {
	p := &Prompt{}

	if p.ID() != "" {
		t.Errorf("ID() = %q, want empty", p.ID())
	}
	if p.Name() != "" {
		t.Errorf("Name() = %q, want empty", p.Name())
	}
	if p.Description() != "" {
		t.Errorf("Description() = %q, want empty", p.Description())
	}
	if p.Template() != "" {
		t.Errorf("Template() = %q, want empty", p.Template())
	}
	if p.Tags() != nil {
		t.Errorf("Tags() = %v, want nil", p.Tags())
	}
}

func TestPromptVersionEmptyState(t *testing.T) {
	v := &PromptVersion{}

	if v.ID() != "" {
		t.Errorf("ID() = %q, want empty", v.ID())
	}
	if v.PromptID() != "" {
		t.Errorf("PromptID() = %q, want empty", v.PromptID())
	}
	if v.Commit() != "" {
		t.Errorf("Commit() = %q, want empty", v.Commit())
	}
	if v.Template() != "" {
		t.Errorf("Template() = %q, want empty", v.Template())
	}
	if v.ChangeDescription() != "" {
		t.Errorf("ChangeDescription() = %q, want empty", v.ChangeDescription())
	}
	if v.Tags() != nil {
		t.Errorf("Tags() = %v, want nil", v.Tags())
	}
}

func TestPromptOptionsChaining(t *testing.T) {
	opts := &promptOptions{}

	options := []PromptOption{
		WithPromptDescription("chained description"),
		WithPromptTemplate("{{chained}}"),
		WithPromptChangeDescription("chained change"),
		WithPromptType(PromptTypeFString),
		WithPromptTemplateStructure(PromptTemplateStructureChat),
		WithPromptTags("a", "b"),
	}

	for _, opt := range options {
		opt(opts)
	}

	if opts.description != "chained description" {
		t.Errorf("description = %q, want %q", opts.description, "chained description")
	}
	if opts.template != "{{chained}}" {
		t.Errorf("template = %q, want %q", opts.template, "{{chained}}")
	}
	if opts.changeDescription != "chained change" {
		t.Errorf("changeDescription = %q, want %q", opts.changeDescription, "chained change")
	}
	if opts.promptType != PromptTypeFString {
		t.Errorf("promptType = %q, want %q", opts.promptType, PromptTypeFString)
	}
	if opts.templateStructure != PromptTemplateStructureChat {
		t.Errorf("templateStructure = %q, want %q", opts.templateStructure, PromptTemplateStructureChat)
	}
	if len(opts.tags) != 2 {
		t.Errorf("tags length = %d, want 2", len(opts.tags))
	}
}
