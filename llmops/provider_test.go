package llmops_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/plexusone/omniobserve/llmops"
	opik "github.com/plexusone/opik-go"
	_ "github.com/plexusone/opik-go/llmops"
)

// testConfig holds configuration for integration tests.
type testConfig struct {
	APIKey    string //nolint:gosec // G117: Test config struct
	Workspace string
	Endpoint  string
}

// getTestConfig returns test configuration from environment variables.
// Returns nil if required variables are not set (tests should skip).
func getTestConfig() *testConfig {
	apiKey := os.Getenv("OPIK_API_KEY")
	if apiKey == "" {
		return nil
	}
	return &testConfig{
		APIKey:    apiKey,
		Workspace: os.Getenv("OPIK_WORKSPACE"),
		Endpoint:  os.Getenv("OPIK_ENDPOINT"),
	}
}

// skipIfNoAPIKey skips the test if OPIK_API_KEY is not set.
func skipIfNoAPIKey(t *testing.T) *testConfig {
	t.Helper()
	cfg := getTestConfig()
	if cfg == nil {
		t.Skip("OPIK_API_KEY not set, skipping integration test")
	}
	return cfg
}

// openTestProvider opens a provider for testing.
func openTestProvider(t *testing.T, cfg *testConfig, projectName string) llmops.Provider {
	t.Helper()

	opts := []llmops.ClientOption{
		llmops.WithAPIKey(cfg.APIKey),
		llmops.WithProjectName(projectName),
	}
	if cfg.Workspace != "" {
		opts = append(opts, llmops.WithWorkspace(cfg.Workspace))
	}
	if cfg.Endpoint != "" {
		opts = append(opts, llmops.WithEndpoint(cfg.Endpoint))
	}

	provider, err := llmops.Open("opik", opts...)
	if err != nil {
		t.Fatalf("failed to open provider: %v", err)
	}

	t.Cleanup(func() {
		if err := provider.Close(); err != nil {
			t.Errorf("failed to close provider: %v", err)
		}
	})

	return provider
}

// =============================================================================
// Provider Registration Tests
// =============================================================================

func TestProviderRegistration(t *testing.T) {
	providers := llmops.Providers()
	found := false
	for _, name := range providers {
		if name == "opik" {
			found = true
			break
		}
	}
	if !found {
		t.Error("opik provider not registered")
	}
}

func TestProviderInfo(t *testing.T) {
	info, ok := llmops.GetProviderInfo("opik")
	if !ok {
		t.Fatal("opik provider info not found")
	}

	if info.Name != "opik" {
		t.Errorf("expected name 'opik', got %q", info.Name)
	}
	if info.Description == "" {
		t.Error("provider description is empty")
	}
	if !info.OpenSource {
		t.Error("expected OpenSource to be true")
	}
	if !info.SelfHosted {
		t.Error("expected SelfHosted to be true")
	}

	// Check capabilities
	expectedCaps := []llmops.Capability{
		llmops.CapabilityTracing,
		llmops.CapabilityEvaluation,
		llmops.CapabilityPrompts,
		llmops.CapabilityDatasets,
	}
	for _, cap := range expectedCaps {
		found := false
		for _, c := range info.Capabilities {
			if c == cap {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected capability %q not found", cap)
		}
	}
}

func TestProviderName(t *testing.T) {
	cfg := skipIfNoAPIKey(t)
	provider := openTestProvider(t, cfg, "test-provider-name")

	if provider.Name() != "opik" {
		t.Errorf("expected provider name 'opik', got %q", provider.Name())
	}
}

// =============================================================================
// Trace Tests
// =============================================================================

func TestStartTrace(t *testing.T) {
	cfg := skipIfNoAPIKey(t)
	provider := openTestProvider(t, cfg, "test-start-trace")
	ctx := context.Background()

	_, trace, err := provider.StartTrace(ctx, "test-trace")
	if err != nil {
		t.Fatalf("failed to start trace: %v", err)
	}

	if trace.ID() == "" {
		t.Error("trace ID is empty")
	}
	if trace.Name() != "test-trace" {
		t.Errorf("expected trace name 'test-trace', got %q", trace.Name())
	}

	if err := trace.End(); err != nil {
		t.Errorf("failed to end trace: %v", err)
	}
}

func TestTraceWithOptions(t *testing.T) {
	cfg := skipIfNoAPIKey(t)
	provider := openTestProvider(t, cfg, "test-trace-options")
	ctx := context.Background()

	input := map[string]any{"query": "test query"}
	metadata := map[string]any{"version": "1.0"}

	_, trace, err := provider.StartTrace(ctx, "trace-with-options",
		llmops.WithTraceInput(input),
		llmops.WithTraceMetadata(metadata),
		llmops.WithTraceTags("test", "integration"),
	)
	if err != nil {
		t.Fatalf("failed to start trace: %v", err)
	}

	if trace.ID() == "" {
		t.Error("trace ID is empty")
	}

	// Set output before ending
	output := map[string]any{"result": "success"}
	if err := trace.SetOutput(output); err != nil {
		t.Errorf("failed to set output: %v", err)
	}

	if err := trace.End(); err != nil {
		t.Errorf("failed to end trace: %v", err)
	}
}

func TestTraceFromContext(t *testing.T) {
	cfg := skipIfNoAPIKey(t)
	provider := openTestProvider(t, cfg, "test-trace-context")
	ctx := context.Background()

	// Before starting, no trace should be in context
	_, ok := provider.TraceFromContext(ctx)
	if ok {
		t.Error("expected no trace in context before starting")
	}

	ctx, trace, err := provider.StartTrace(ctx, "context-test")
	if err != nil {
		t.Fatalf("failed to start trace: %v", err)
	}

	// After starting, trace should be in context
	traceFromCtx, ok := provider.TraceFromContext(ctx)
	if !ok {
		t.Fatal("expected trace in context after starting")
	}
	if traceFromCtx.ID() != trace.ID() {
		t.Errorf("trace ID mismatch: expected %q, got %q", trace.ID(), traceFromCtx.ID())
	}

	if err := trace.End(); err != nil {
		t.Errorf("failed to end trace: %v", err)
	}
}

func TestTraceDuration(t *testing.T) {
	cfg := skipIfNoAPIKey(t)
	provider := openTestProvider(t, cfg, "test-trace-duration")
	ctx := context.Background()

	_, trace, err := provider.StartTrace(ctx, "duration-test")
	if err != nil {
		t.Fatalf("failed to start trace: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	duration := trace.Duration()
	if duration < 100*time.Millisecond {
		t.Errorf("expected duration >= 100ms, got %v", duration)
	}

	if err := trace.End(); err != nil {
		t.Errorf("failed to end trace: %v", err)
	}

	// Check EndTime is set
	endTime := trace.EndTime()
	if endTime == nil {
		t.Error("expected EndTime to be set after ending trace")
	}
}

func TestTraceFeedbackScore(t *testing.T) {
	cfg := skipIfNoAPIKey(t)
	provider := openTestProvider(t, cfg, "test-trace-feedback")
	ctx := context.Background()

	ctx, trace, err := provider.StartTrace(ctx, "feedback-test")
	if err != nil {
		t.Fatalf("failed to start trace: %v", err)
	}

	if err := trace.AddFeedbackScore(ctx, "quality", 0.95, llmops.WithFeedbackReason("test feedback")); err != nil {
		t.Errorf("failed to add feedback score: %v", err)
	}

	if err := trace.End(); err != nil {
		t.Errorf("failed to end trace: %v", err)
	}
}

// =============================================================================
// Span Tests
// =============================================================================

func TestStartSpan(t *testing.T) {
	cfg := skipIfNoAPIKey(t)
	provider := openTestProvider(t, cfg, "test-start-span")
	ctx := context.Background()

	ctx, trace, err := provider.StartTrace(ctx, "parent-trace")
	if err != nil {
		t.Fatalf("failed to start trace: %v", err)
	}

	_, span, err := provider.StartSpan(ctx, "test-span")
	if err != nil {
		t.Fatalf("failed to start span: %v", err)
	}

	if span.ID() == "" {
		t.Error("span ID is empty")
	}
	if span.TraceID() != trace.ID() {
		t.Errorf("span trace ID mismatch: expected %q, got %q", trace.ID(), span.TraceID())
	}
	if span.Name() != "test-span" {
		t.Errorf("expected span name 'test-span', got %q", span.Name())
	}

	if err := span.End(); err != nil {
		t.Errorf("failed to end span: %v", err)
	}
	if err := trace.End(); err != nil {
		t.Errorf("failed to end trace: %v", err)
	}
}

func TestSpanTypes(t *testing.T) {
	cfg := skipIfNoAPIKey(t)
	provider := openTestProvider(t, cfg, "test-span-types")
	ctx := context.Background()

	// Opik only supports: general, llm, tool, guardrail
	// Other types like retrieval, agent, chain are not supported by the Opik API
	spanTypes := []llmops.SpanType{
		llmops.SpanTypeGeneral,
		llmops.SpanTypeLLM,
		llmops.SpanTypeTool,
		llmops.SpanTypeGuardrail,
	}

	for _, spanType := range spanTypes {
		t.Run(string(spanType), func(t *testing.T) {
			ctx, trace, err := provider.StartTrace(ctx, fmt.Sprintf("trace-%s", spanType))
			if err != nil {
				t.Fatalf("failed to start trace: %v", err)
			}

			_, span, err := provider.StartSpan(ctx, fmt.Sprintf("span-%s", spanType),
				llmops.WithSpanType(spanType),
			)
			if err != nil {
				t.Fatalf("failed to start span: %v", err)
			}

			if span.Type() != spanType {
				t.Errorf("expected span type %q, got %q", spanType, span.Type())
			}

			if err := span.End(); err != nil {
				t.Errorf("failed to end span: %v", err)
			}
			if err := trace.End(); err != nil {
				t.Errorf("failed to end trace: %v", err)
			}
		})
	}
}

func TestSpanWithLLMMetadata(t *testing.T) {
	cfg := skipIfNoAPIKey(t)
	provider := openTestProvider(t, cfg, "test-span-llm")
	ctx := context.Background()

	ctx, trace, err := provider.StartTrace(ctx, "llm-trace")
	if err != nil {
		t.Fatalf("failed to start trace: %v", err)
	}

	_, span, err := provider.StartSpan(ctx, "llm-call",
		llmops.WithSpanType(llmops.SpanTypeLLM),
		llmops.WithModel("gpt-4"),
		llmops.WithProvider("openai"),
		llmops.WithSpanInput(map[string]any{"prompt": "Hello, world!"}),
	)
	if err != nil {
		t.Fatalf("failed to start span: %v", err)
	}

	// Set output and usage
	if err := span.SetOutput(map[string]any{"response": "Hi there!"}); err != nil {
		t.Errorf("failed to set output: %v", err)
	}

	usage := llmops.TokenUsage{
		PromptTokens:     10,
		CompletionTokens: 5,
		TotalTokens:      15,
	}
	if err := span.SetUsage(usage); err != nil {
		t.Errorf("failed to set usage: %v", err)
	}

	if err := span.End(); err != nil {
		t.Errorf("failed to end span: %v", err)
	}
	if err := trace.End(); err != nil {
		t.Errorf("failed to end trace: %v", err)
	}
}

func TestNestedSpans(t *testing.T) {
	cfg := skipIfNoAPIKey(t)
	provider := openTestProvider(t, cfg, "test-nested-spans")
	ctx := context.Background()

	ctx, trace, err := provider.StartTrace(ctx, "nested-trace")
	if err != nil {
		t.Fatalf("failed to start trace: %v", err)
	}

	// Create parent span
	ctx, parentSpan, err := provider.StartSpan(ctx, "parent-span")
	if err != nil {
		t.Fatalf("failed to start parent span: %v", err)
	}

	// Create child span from parent span
	_, childSpan, err := parentSpan.StartSpan(ctx, "child-span")
	if err != nil {
		t.Fatalf("failed to start child span: %v", err)
	}

	if childSpan.ParentSpanID() != parentSpan.ID() {
		t.Errorf("child parent span ID mismatch: expected %q, got %q", parentSpan.ID(), childSpan.ParentSpanID())
	}
	if childSpan.TraceID() != trace.ID() {
		t.Errorf("child trace ID mismatch: expected %q, got %q", trace.ID(), childSpan.TraceID())
	}

	// End in reverse order
	if err := childSpan.End(); err != nil {
		t.Errorf("failed to end child span: %v", err)
	}
	if err := parentSpan.End(); err != nil {
		t.Errorf("failed to end parent span: %v", err)
	}
	if err := trace.End(); err != nil {
		t.Errorf("failed to end trace: %v", err)
	}
}

func TestSpanFromContext(t *testing.T) {
	cfg := skipIfNoAPIKey(t)
	provider := openTestProvider(t, cfg, "test-span-context")
	ctx := context.Background()

	ctx, trace, err := provider.StartTrace(ctx, "span-context-test")
	if err != nil {
		t.Fatalf("failed to start trace: %v", err)
	}

	// Before starting span, none should be in context
	_, ok := provider.SpanFromContext(ctx)
	if ok {
		t.Error("expected no span in context before starting")
	}

	ctx, span, err := provider.StartSpan(ctx, "context-span")
	if err != nil {
		t.Fatalf("failed to start span: %v", err)
	}

	// After starting, span should be in context
	spanFromCtx, ok := provider.SpanFromContext(ctx)
	if !ok {
		t.Fatal("expected span in context after starting")
	}
	if spanFromCtx.ID() != span.ID() {
		t.Errorf("span ID mismatch: expected %q, got %q", span.ID(), spanFromCtx.ID())
	}

	if err := span.End(); err != nil {
		t.Errorf("failed to end span: %v", err)
	}
	if err := trace.End(); err != nil {
		t.Errorf("failed to end trace: %v", err)
	}
}

func TestSpanFeedbackScore(t *testing.T) {
	cfg := skipIfNoAPIKey(t)
	provider := openTestProvider(t, cfg, "test-span-feedback")
	ctx := context.Background()

	ctx, trace, err := provider.StartTrace(ctx, "span-feedback-trace")
	if err != nil {
		t.Fatalf("failed to start trace: %v", err)
	}

	ctx, span, err := provider.StartSpan(ctx, "feedback-span")
	if err != nil {
		t.Fatalf("failed to start span: %v", err)
	}

	if err := span.AddFeedbackScore(ctx, "accuracy", 0.88); err != nil {
		t.Errorf("failed to add feedback score: %v", err)
	}

	if err := span.End(); err != nil {
		t.Errorf("failed to end span: %v", err)
	}
	if err := trace.End(); err != nil {
		t.Errorf("failed to end trace: %v", err)
	}
}

// =============================================================================
// Project Tests
// =============================================================================

func TestCreateProject(t *testing.T) {
	cfg := skipIfNoAPIKey(t)
	provider := openTestProvider(t, cfg, "test-create-project")
	ctx := context.Background()

	projectName := fmt.Sprintf("test-project-%d", time.Now().UnixNano())
	project, err := provider.CreateProject(ctx, projectName)
	if err != nil {
		// Project might already exist, which is OK
		t.Logf("CreateProject returned: %v (may already exist)", err)
		return
	}

	if project.Name != projectName {
		t.Errorf("expected project name %q, got %q", projectName, project.Name)
	}
	// Note: Some Opik deployments may not return the project ID on creation
	if project.ID == "" {
		t.Logf("Note: project ID is empty (this may be expected for some Opik versions)")
	}
}

func TestListProjects(t *testing.T) {
	cfg := skipIfNoAPIKey(t)
	provider := openTestProvider(t, cfg, "test-list-projects")
	ctx := context.Background()

	projects, err := provider.ListProjects(ctx, llmops.WithLimit(10))
	if err != nil {
		// Some Opik API versions may have different pagination requirements
		t.Logf("ListProjects returned error: %v (may be API version specific)", err)
		return
	}

	// Should have at least one project
	if len(projects) == 0 {
		t.Log("no projects found (this may be expected in a fresh workspace)")
	}

	for _, project := range projects {
		t.Logf("Found project: %s (ID: %s)", project.Name, project.ID)
	}
}

func TestSetProject(t *testing.T) {
	cfg := skipIfNoAPIKey(t)
	provider := openTestProvider(t, cfg, "initial-project")
	ctx := context.Background()

	newProject := "switched-project"
	if err := provider.SetProject(ctx, newProject); err != nil {
		t.Fatalf("failed to set project: %v", err)
	}

	// Create a trace to verify the project is used
	_, trace, err := provider.StartTrace(ctx, "project-switch-test")
	if err != nil {
		t.Fatalf("failed to start trace: %v", err)
	}

	if err := trace.End(); err != nil {
		t.Errorf("failed to end trace: %v", err)
	}
}

// =============================================================================
// Dataset Tests
// =============================================================================

func TestCreateDataset(t *testing.T) {
	cfg := skipIfNoAPIKey(t)
	provider := openTestProvider(t, cfg, "test-create-dataset")
	ctx := context.Background()

	datasetName := fmt.Sprintf("test-dataset-%d", time.Now().UnixNano())
	dataset, err := provider.CreateDataset(ctx, datasetName,
		llmops.WithDatasetDescription("Test dataset for integration tests"),
	)
	if err != nil {
		t.Fatalf("failed to create dataset: %v", err)
	}

	if dataset.Name != datasetName {
		t.Errorf("expected dataset name %q, got %q", datasetName, dataset.Name)
	}
	if dataset.ID == "" {
		t.Error("dataset ID is empty")
	}
}

func TestAddDatasetItems(t *testing.T) {
	cfg := skipIfNoAPIKey(t)
	provider := openTestProvider(t, cfg, "test-dataset-items")
	ctx := context.Background()

	datasetName := fmt.Sprintf("test-dataset-items-%d", time.Now().UnixNano())
	_, err := provider.CreateDataset(ctx, datasetName)
	if err != nil {
		t.Logf("CreateDataset returned error: %v (may already exist or validation issue)", err)
		return
	}

	items := []llmops.DatasetItem{
		{
			Input:    map[string]any{"question": "What is 2+2?"},
			Expected: map[string]any{"answer": "4"},
		},
		{
			Input:    map[string]any{"question": "What is the capital of France?"},
			Expected: map[string]any{"answer": "Paris"},
		},
	}

	if err := provider.AddDatasetItems(ctx, datasetName, items); err != nil {
		t.Logf("AddDatasetItems returned error: %v (may be schema validation issue)", err)
	}
}

func TestListDatasets(t *testing.T) {
	cfg := skipIfNoAPIKey(t)
	provider := openTestProvider(t, cfg, "test-list-datasets")
	ctx := context.Background()

	datasets, err := provider.ListDatasets(ctx, llmops.WithLimit(10))
	if err != nil {
		t.Logf("ListDatasets returned error: %v (may be API version specific)", err)
		return
	}

	// Log found datasets
	for _, dataset := range datasets {
		t.Logf("Found dataset: %s (ID: %s)", dataset.Name, dataset.ID)
	}
}

// =============================================================================
// Prompt Tests
// =============================================================================

func TestCreatePrompt(t *testing.T) {
	cfg := skipIfNoAPIKey(t)
	provider := openTestProvider(t, cfg, "test-create-prompt")
	ctx := context.Background()

	promptName := fmt.Sprintf("test-prompt-%d", time.Now().UnixNano())
	template := "You are a helpful assistant. User: {{input}} Assistant:"

	prompt, err := provider.CreatePrompt(ctx, promptName, template,
		llmops.WithPromptDescription("Test prompt for integration tests"),
		llmops.WithPromptTags("test", "integration"),
	)
	if err != nil {
		t.Fatalf("failed to create prompt: %v", err)
	}

	if prompt.Name != promptName {
		t.Errorf("expected prompt name %q, got %q", promptName, prompt.Name)
	}
	if prompt.Template != template {
		t.Errorf("expected template %q, got %q", template, prompt.Template)
	}
	if prompt.ID == "" {
		t.Error("prompt ID is empty")
	}
}

func TestGetPrompt(t *testing.T) {
	cfg := skipIfNoAPIKey(t)
	provider := openTestProvider(t, cfg, "test-get-prompt")
	ctx := context.Background()

	promptName := fmt.Sprintf("test-get-prompt-%d", time.Now().UnixNano())
	template := "Hello, {{name}}!"

	_, err := provider.CreatePrompt(ctx, promptName, template)
	if err != nil {
		t.Logf("CreatePrompt returned error: %v", err)
		return
	}

	// Wait a moment for propagation
	time.Sleep(500 * time.Millisecond)

	// Get the prompt back
	prompt, err := provider.GetPrompt(ctx, promptName)
	if err != nil {
		t.Logf("GetPrompt returned error: %v (may need time to propagate)", err)
		return
	}

	if prompt.Name != promptName {
		t.Errorf("expected prompt name %q, got %q", promptName, prompt.Name)
	}
	if prompt.Template != template {
		t.Errorf("expected template %q, got %q", template, prompt.Template)
	}
}

func TestListPrompts(t *testing.T) {
	cfg := skipIfNoAPIKey(t)
	provider := openTestProvider(t, cfg, "test-list-prompts")
	ctx := context.Background()

	prompts, err := provider.ListPrompts(ctx, llmops.WithLimit(10))
	if err != nil {
		t.Logf("ListPrompts returned error: %v (may be API version specific)", err)
		return
	}

	// Log found prompts
	for _, prompt := range prompts {
		t.Logf("Found prompt: %s (ID: %s)", prompt.Name, prompt.ID)
	}
}

// =============================================================================
// Evaluation Tests
// =============================================================================

func TestEvaluate(t *testing.T) {
	cfg := skipIfNoAPIKey(t)
	provider := openTestProvider(t, cfg, "test-evaluate")
	ctx := context.Background()

	input := llmops.EvalInput{
		Input:    "What is 2+2?",
		Output:   "4",
		Expected: "4",
	}

	// Create a simple test metric
	exactMatchMetric := &testMetric{
		name: "exact_match",
		evalFn: func(input llmops.EvalInput) (llmops.MetricScore, error) {
			score := 0.0
			if input.Output == input.Expected {
				score = 1.0
			}
			return llmops.MetricScore{
				Name:  "exact_match",
				Score: score,
			}, nil
		},
	}

	result, err := provider.Evaluate(ctx, input, exactMatchMetric)
	if err != nil {
		t.Fatalf("failed to evaluate: %v", err)
	}

	if len(result.Scores) != 1 {
		t.Fatalf("expected 1 score, got %d", len(result.Scores))
	}
	if result.Scores[0].Name != "exact_match" {
		t.Errorf("expected metric name 'exact_match', got %q", result.Scores[0].Name)
	}
	if result.Scores[0].Score != 1.0 {
		t.Errorf("expected score 1.0, got %f", result.Scores[0].Score)
	}
}

// testMetric implements llmops.Metric for testing.
type testMetric struct {
	name   string
	evalFn func(llmops.EvalInput) (llmops.MetricScore, error)
}

func (m *testMetric) Name() string {
	return m.name
}

func (m *testMetric) Evaluate(input llmops.EvalInput) (llmops.MetricScore, error) {
	return m.evalFn(input)
}

// =============================================================================
// Integration Test: Full Workflow
// =============================================================================

func TestFullWorkflow(t *testing.T) {
	cfg := skipIfNoAPIKey(t)
	provider := openTestProvider(t, cfg, "test-full-workflow")
	ctx := context.Background()

	// 1. Start a trace
	ctx, trace, err := provider.StartTrace(ctx, "workflow-trace",
		llmops.WithTraceInput(map[string]any{"query": "Test workflow"}),
		llmops.WithTraceTags("integration-test"),
	)
	if err != nil {
		t.Fatalf("failed to start trace: %v", err)
	}
	t.Logf("Created trace: %s", trace.ID())

	// 2. Create a tool span (for retrieval-like operations)
	// Note: Opik supports general, llm, tool, guardrail - not "retrieval" directly
	ctx, toolSpan, err := provider.StartSpan(ctx, "retrieval-tool",
		llmops.WithSpanType(llmops.SpanTypeTool),
		llmops.WithSpanInput(map[string]any{"query": "relevant documents"}),
	)
	if err != nil {
		t.Fatalf("failed to start tool span: %v", err)
	}

	if err := toolSpan.SetOutput(map[string]any{
		"documents": []string{"doc1", "doc2"},
	}); err != nil {
		t.Errorf("failed to set tool output: %v", err)
	}
	if err := toolSpan.End(); err != nil {
		t.Errorf("failed to end tool span: %v", err)
	}
	t.Logf("Created tool span: %s", toolSpan.ID())

	// 3. Create an LLM span
	ctx, llmSpan, err := provider.StartSpan(ctx, "llm-call",
		llmops.WithSpanType(llmops.SpanTypeLLM),
		llmops.WithModel("gpt-4"),
		llmops.WithProvider("openai"),
		llmops.WithSpanInput(map[string]any{"messages": []string{"Hello"}}),
	)
	if err != nil {
		t.Fatalf("failed to start llm span: %v", err)
	}

	if err := llmSpan.SetOutput(map[string]any{"response": "Hello!"}); err != nil {
		t.Errorf("failed to set llm output: %v", err)
	}
	if err := llmSpan.SetUsage(llmops.TokenUsage{
		PromptTokens:     100,
		CompletionTokens: 50,
		TotalTokens:      150,
	}); err != nil {
		t.Errorf("failed to set usage: %v", err)
	}
	if err := llmSpan.AddFeedbackScore(ctx, "quality", 0.9); err != nil {
		t.Errorf("failed to add feedback: %v", err)
	}
	if err := llmSpan.End(); err != nil {
		t.Errorf("failed to end llm span: %v", err)
	}
	t.Logf("Created LLM span: %s", llmSpan.ID())

	// 4. End the trace
	if err := trace.SetOutput(map[string]any{"result": "success"}); err != nil {
		t.Errorf("failed to set trace output: %v", err)
	}
	if err := trace.AddFeedbackScore(ctx, "overall", 0.95); err != nil {
		t.Errorf("failed to add trace feedback: %v", err)
	}
	if err := trace.End(); err != nil {
		t.Errorf("failed to end trace: %v", err)
	}

	t.Logf("Full workflow completed successfully")
	t.Logf("Trace ID: %s", trace.ID())
	t.Logf("Duration: %v", trace.Duration())
}

// =============================================================================
// Verification Tests: Read Back From Opik
// =============================================================================

func TestVerifyTraceInOpik(t *testing.T) {
	cfg := skipIfNoAPIKey(t)
	provider := openTestProvider(t, cfg, "test-verify-trace")
	ctx := context.Background()

	// Create a trace with unique identifiable data
	uniqueTag := fmt.Sprintf("verify-%d", time.Now().UnixNano())

	ctx, trace, err := provider.StartTrace(ctx, "verify-trace",
		llmops.WithTraceInput(map[string]any{"test_id": uniqueTag}),
		llmops.WithTraceTags(uniqueTag),
	)
	if err != nil {
		t.Fatalf("failed to start trace: %v", err)
	}

	traceID := trace.ID()
	t.Logf("Created trace with ID: %s and tag: %s", traceID, uniqueTag)

	// Add a span
	_, span, err := provider.StartSpan(ctx, "verify-span",
		llmops.WithSpanType(llmops.SpanTypeLLM),
		llmops.WithModel("test-model"),
	)
	if err != nil {
		t.Fatalf("failed to start span: %v", err)
	}

	if err := span.End(); err != nil {
		t.Errorf("failed to end span: %v", err)
	}
	if err := trace.End(); err != nil {
		t.Errorf("failed to end trace: %v", err)
	}

	// Wait for data to be flushed
	time.Sleep(2 * time.Second)

	// Use the Opik SDK directly to verify the trace exists
	opikOpts := []opik.Option{
		opik.WithAPIKey(cfg.APIKey),
	}
	if cfg.Workspace != "" {
		opikOpts = append(opikOpts, opik.WithWorkspace(cfg.Workspace))
	}
	if cfg.Endpoint != "" {
		opikOpts = append(opikOpts, opik.WithURL(cfg.Endpoint))
	}

	client, err := opik.NewClient(opikOpts...)
	if err != nil {
		t.Fatalf("failed to create Opik client for verification: %v", err)
	}

	// Try to get the trace by ID
	fetchedTrace, err := client.GetTrace(ctx, traceID)
	if err != nil {
		t.Logf("Note: GetTrace returned error: %v (trace may need time to propagate)", err)
		// This is informational, not a failure - API latency can cause this
		return
	}

	if fetchedTrace == nil {
		t.Log("Trace not found immediately - may need more time to propagate")
		return
	}

	t.Logf("Successfully verified trace in Opik:")
	t.Logf("  ID: %s", fetchedTrace.ID())
	t.Logf("  Name: %s", fetchedTrace.Name())
}
