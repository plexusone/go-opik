// Package llmops provides an omniobserve/llmops adapter for go-opik.
//
// Import this package to register the Opik provider:
//
//	import _ "github.com/agentplexus/go-opik/llmops"
//
// Then open it:
//
//	provider, err := llmops.Open("opik", llmops.WithAPIKey("..."))
//
// The standalone go-opik SDK can still be used directly without this package:
//
//	import opik "github.com/agentplexus/go-opik"
//	client, err := opik.NewClient(opik.WithAPIKey("..."))
package llmops

import (
	"context"
	"time"

	opik "github.com/agentplexus/go-opik"
	"github.com/agentplexus/omniobserve/llmops"
)

// ProviderName is the name used to register and open this provider.
const ProviderName = "opik"

func init() {
	llmops.Register(ProviderName, New)
	llmops.RegisterInfo(llmops.ProviderInfo{
		Name:        ProviderName,
		Description: "Comet Opik - Open-source LLM observability platform",
		Website:     "https://github.com/comet-ml/opik",
		OpenSource:  true,
		SelfHosted:  true,
		Capabilities: []llmops.Capability{
			llmops.CapabilityTracing,
			llmops.CapabilityEvaluation,
			llmops.CapabilityPrompts,
			llmops.CapabilityDatasets,
			llmops.CapabilityExperiments,
			llmops.CapabilityStreaming,
			llmops.CapabilityDistributed,
			llmops.CapabilityCostTracking,
		},
	})
}

// Provider implements llmops.Provider for Opik.
type Provider struct {
	client      *opik.Client
	projectName string
}

// New creates a new Opik provider.
func New(opts ...llmops.ClientOption) (llmops.Provider, error) {
	cfg := llmops.ApplyClientOptions(opts...)

	// Map llmops options to opik options
	opikOpts := []opik.Option{}
	if cfg.APIKey != "" {
		opikOpts = append(opikOpts, opik.WithAPIKey(cfg.APIKey))
	}
	if cfg.Endpoint != "" {
		opikOpts = append(opikOpts, opik.WithURL(cfg.Endpoint))
	}
	if cfg.Workspace != "" {
		opikOpts = append(opikOpts, opik.WithWorkspace(cfg.Workspace))
	}
	if cfg.ProjectName != "" {
		opikOpts = append(opikOpts, opik.WithProjectName(cfg.ProjectName))
	}
	if cfg.HTTPClient != nil {
		opikOpts = append(opikOpts, opik.WithHTTPClient(cfg.HTTPClient))
	}
	if cfg.Timeout > 0 {
		opikOpts = append(opikOpts, opik.WithTimeout(cfg.Timeout))
	}
	if cfg.Disabled {
		opikOpts = append(opikOpts, opik.WithTracingDisabled(true))
	}

	client, err := opik.NewClient(opikOpts...)
	if err != nil {
		return nil, err
	}

	return &Provider{
		client:      client,
		projectName: cfg.ProjectName,
	}, nil
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return ProviderName
}

// Close closes the provider.
func (p *Provider) Close() error {
	// Opik client doesn't have a Close method currently
	return nil
}

// StartTrace starts a new trace.
func (p *Provider) StartTrace(ctx context.Context, name string, opts ...llmops.TraceOption) (context.Context, llmops.Trace, error) {
	cfg := llmops.ApplyTraceOptions(opts...)

	opikOpts := []opik.TraceOption{}
	if cfg.ProjectName != "" {
		opikOpts = append(opikOpts, opik.WithTraceProject(cfg.ProjectName))
	}
	if cfg.Input != nil {
		opikOpts = append(opikOpts, opik.WithTraceInput(cfg.Input))
	}
	if cfg.Metadata != nil {
		opikOpts = append(opikOpts, opik.WithTraceMetadata(cfg.Metadata))
	}
	if len(cfg.Tags) > 0 {
		opikOpts = append(opikOpts, opik.WithTraceTags(cfg.Tags...))
	}
	if cfg.ThreadID != "" {
		opikOpts = append(opikOpts, opik.WithTraceThreadID(cfg.ThreadID))
	}

	newCtx, trace, err := opik.StartTrace(ctx, p.client, name, opikOpts...)
	if err != nil {
		return ctx, nil, err
	}

	return newCtx, &traceAdapter{trace: trace}, nil
}

// StartSpan starts a new span.
func (p *Provider) StartSpan(ctx context.Context, name string, opts ...llmops.SpanOption) (context.Context, llmops.Span, error) {
	cfg := llmops.ApplySpanOptions(opts...)

	opikOpts := mapSpanOptions(cfg)

	newCtx, span, err := opik.StartSpan(ctx, name, opikOpts...)
	if err != nil {
		return ctx, nil, err
	}

	return newCtx, &spanAdapter{span: span}, nil
}

// TraceFromContext gets the current trace from context.
func (p *Provider) TraceFromContext(ctx context.Context) (llmops.Trace, bool) {
	trace := opik.TraceFromContext(ctx)
	if trace == nil {
		return nil, false
	}
	return &traceAdapter{trace: trace}, true
}

// SpanFromContext gets the current span from context.
func (p *Provider) SpanFromContext(ctx context.Context) (llmops.Span, bool) {
	span := opik.SpanFromContext(ctx)
	if span == nil {
		return nil, false
	}
	return &spanAdapter{span: span}, true
}

// Evaluate runs evaluation metrics.
func (p *Provider) Evaluate(ctx context.Context, input llmops.EvalInput, metrics ...llmops.Metric) (*llmops.EvalResult, error) {
	startTime := time.Now()

	scores := make([]llmops.MetricScore, 0, len(metrics))
	for _, metric := range metrics {
		score, err := metric.Evaluate(input)
		if err != nil {
			scores = append(scores, llmops.MetricScore{
				Name:  metric.Name(),
				Error: err.Error(),
			})
		} else {
			scores = append(scores, score)
		}
	}

	return &llmops.EvalResult{
		Scores:   scores,
		Duration: time.Since(startTime),
	}, nil
}

// AddFeedbackScore adds a feedback score.
func (p *Provider) AddFeedbackScore(ctx context.Context, opts llmops.FeedbackScoreOpts) error {
	reason := opts.Reason
	if reason == "" {
		reason = "feedback"
	}

	// Try span first, then trace
	if span := opik.SpanFromContext(ctx); span != nil {
		return span.AddFeedbackScore(ctx, opts.Name, opts.Score, reason)
	}
	if trace := opik.TraceFromContext(ctx); trace != nil {
		return trace.AddFeedbackScore(ctx, opts.Name, opts.Score, reason)
	}
	return llmops.ErrNoActiveTrace
}

// CreatePrompt creates a new prompt.
func (p *Provider) CreatePrompt(ctx context.Context, name string, template string, opts ...llmops.PromptOption) (*llmops.Prompt, error) {
	cfg := &llmops.PromptOptions{}
	for _, opt := range opts {
		opt(cfg)
	}

	opikOpts := []opik.PromptOption{
		opik.WithPromptTemplate(template),
	}
	if cfg.Description != "" {
		opikOpts = append(opikOpts, opik.WithPromptDescription(cfg.Description))
	}
	if len(cfg.Tags) > 0 {
		opikOpts = append(opikOpts, opik.WithPromptTags(cfg.Tags...))
	}

	prompt, err := p.client.CreatePrompt(ctx, name, opikOpts...)
	if err != nil {
		return nil, err
	}

	return &llmops.Prompt{
		ID:          prompt.ID(),
		Name:        prompt.Name(),
		Template:    prompt.Template(),
		Description: prompt.Description(),
		Tags:        prompt.Tags(),
	}, nil
}

// GetPrompt gets a prompt by name and optional commit/version.
func (p *Provider) GetPrompt(ctx context.Context, name string, version ...string) (*llmops.Prompt, error) {
	// Opik requires a commit hash; use "latest" as default
	commit := "latest"
	if len(version) > 0 && version[0] != "" {
		commit = version[0]
	}

	promptVersion, err := p.client.GetPromptByName(ctx, name, commit)
	if err != nil {
		return nil, err
	}

	return &llmops.Prompt{
		ID:       promptVersion.ID(),
		Name:     name,
		Template: promptVersion.Template(),
		Version:  promptVersion.Commit(),
		Tags:     promptVersion.Tags(),
	}, nil
}

// ListPrompts lists prompts.
func (p *Provider) ListPrompts(ctx context.Context, opts ...llmops.ListOption) ([]*llmops.Prompt, error) {
	cfg := llmops.ApplyListOptions(opts...)

	prompts, err := p.client.ListPrompts(ctx, cfg.Limit, cfg.Offset)
	if err != nil {
		return nil, err
	}

	result := make([]*llmops.Prompt, len(prompts))
	for i, prompt := range prompts {
		result[i] = &llmops.Prompt{
			ID:          prompt.ID(),
			Name:        prompt.Name(),
			Template:    prompt.Template(),
			Description: prompt.Description(),
			Tags:        prompt.Tags(),
		}
	}
	return result, nil
}

// CreateDataset creates a new dataset.
func (p *Provider) CreateDataset(ctx context.Context, name string, opts ...llmops.DatasetOption) (*llmops.Dataset, error) {
	cfg := &llmops.DatasetOptions{}
	for _, opt := range opts {
		opt(cfg)
	}

	opikOpts := []opik.DatasetOption{}
	if cfg.Description != "" {
		opikOpts = append(opikOpts, opik.WithDatasetDescription(cfg.Description))
	}

	dataset, err := p.client.CreateDataset(ctx, name, opikOpts...)
	if err != nil {
		return nil, err
	}

	return &llmops.Dataset{
		ID:          dataset.ID(),
		Name:        dataset.Name(),
		Description: dataset.Description(),
	}, nil
}

// GetDataset gets a dataset by name.
func (p *Provider) GetDataset(ctx context.Context, name string) (*llmops.Dataset, error) {
	dataset, err := p.client.GetDatasetByName(ctx, name)
	if err != nil {
		return nil, err
	}
	return &llmops.Dataset{
		ID:          dataset.ID(),
		Name:        dataset.Name(),
		Description: dataset.Description(),
	}, nil
}

// AddDatasetItems adds items to a dataset.
func (p *Provider) AddDatasetItems(ctx context.Context, datasetName string, items []llmops.DatasetItem) error {
	dataset, err := p.client.GetDatasetByName(ctx, datasetName)
	if err != nil {
		return err
	}

	// Convert to map format expected by Opik
	opikItems := make([]map[string]any, len(items))
	for i, item := range items {
		data := make(map[string]any)
		if item.Input != nil {
			data["input"] = item.Input
		}
		if item.Expected != nil {
			data["expected"] = item.Expected
		}
		if item.Metadata != nil {
			for k, v := range item.Metadata {
				data[k] = v
			}
		}
		opikItems[i] = data
	}

	return dataset.InsertItems(ctx, opikItems)
}

// ListDatasets lists datasets.
func (p *Provider) ListDatasets(ctx context.Context, opts ...llmops.ListOption) ([]*llmops.Dataset, error) {
	cfg := llmops.ApplyListOptions(opts...)

	datasets, err := p.client.ListDatasets(ctx, cfg.Limit, cfg.Offset)
	if err != nil {
		return nil, err
	}

	result := make([]*llmops.Dataset, len(datasets))
	for i, dataset := range datasets {
		result[i] = &llmops.Dataset{
			ID:          dataset.ID(),
			Name:        dataset.Name(),
			Description: dataset.Description(),
		}
	}
	return result, nil
}

// CreateProject creates a new project.
func (p *Provider) CreateProject(ctx context.Context, name string, opts ...llmops.ProjectOption) (*llmops.Project, error) {
	project, err := p.client.CreateProject(ctx, name)
	if err != nil {
		return nil, err
	}

	return &llmops.Project{
		ID:   project.ID,
		Name: project.Name,
	}, nil
}

// GetProject gets a project by name.
func (p *Provider) GetProject(ctx context.Context, name string) (*llmops.Project, error) {
	projects, err := p.client.ListProjects(ctx, 100, 0)
	if err != nil {
		return nil, err
	}

	for _, project := range projects {
		if project.Name == name {
			return &llmops.Project{
				ID:   project.ID,
				Name: project.Name,
			}, nil
		}
	}
	return nil, llmops.ErrProjectNotFound
}

// ListProjects lists projects.
func (p *Provider) ListProjects(ctx context.Context, opts ...llmops.ListOption) ([]*llmops.Project, error) {
	cfg := llmops.ApplyListOptions(opts...)

	projects, err := p.client.ListProjects(ctx, cfg.Limit, cfg.Offset)
	if err != nil {
		return nil, err
	}

	result := make([]*llmops.Project, len(projects))
	for i, project := range projects {
		result[i] = &llmops.Project{
			ID:   project.ID,
			Name: project.Name,
		}
	}
	return result, nil
}

// SetProject sets the current project.
func (p *Provider) SetProject(ctx context.Context, name string) error {
	p.projectName = name
	p.client.SetProjectName(name)
	return nil
}

// GetDatasetByID gets a dataset by ID.
func (p *Provider) GetDatasetByID(ctx context.Context, id string) (*llmops.Dataset, error) {
	dataset, err := p.client.GetDataset(ctx, id)
	if err != nil {
		return nil, err
	}
	return &llmops.Dataset{
		ID:          dataset.ID(),
		Name:        dataset.Name(),
		Description: dataset.Description(),
	}, nil
}

// DeleteDataset deletes a dataset by ID.
func (p *Provider) DeleteDataset(ctx context.Context, datasetID string) error {
	return p.client.DeleteDataset(ctx, datasetID)
}

// CreateAnnotation creates an annotation on a span or trace.
// Note: Opik uses feedback scores rather than a separate annotation system.
// This method adds feedback scores to approximate annotation functionality.
func (p *Provider) CreateAnnotation(ctx context.Context, annotation llmops.Annotation) error {
	// Opik doesn't have a dedicated annotation API.
	// Feedback scores are the closest equivalent, but they require an active span/trace context.
	// For now, return not implemented until Opik adds annotation support.
	return llmops.WrapNotImplemented(ProviderName, "CreateAnnotation")
}

// ListAnnotations lists annotations for spans or traces.
// Note: Opik uses feedback scores rather than a separate annotation system.
func (p *Provider) ListAnnotations(ctx context.Context, opts llmops.ListAnnotationsOptions) ([]*llmops.Annotation, error) {
	// Opik doesn't have a dedicated annotation API.
	return nil, llmops.WrapNotImplemented(ProviderName, "ListAnnotations")
}

// mapSpanOptions converts llmops span options to Opik span options.
func mapSpanOptions(cfg *llmops.SpanOptions) []opik.SpanOption {
	opikOpts := []opik.SpanOption{}

	if cfg.Type != "" {
		opikOpts = append(opikOpts, opik.WithSpanType(string(cfg.Type)))
	}
	if cfg.Input != nil {
		opikOpts = append(opikOpts, opik.WithSpanInput(cfg.Input))
	}
	if cfg.Metadata != nil {
		opikOpts = append(opikOpts, opik.WithSpanMetadata(cfg.Metadata))
	}
	if len(cfg.Tags) > 0 {
		opikOpts = append(opikOpts, opik.WithSpanTags(cfg.Tags...))
	}
	if cfg.Model != "" {
		opikOpts = append(opikOpts, opik.WithSpanModel(cfg.Model))
	}
	if cfg.Provider != "" {
		opikOpts = append(opikOpts, opik.WithSpanProvider(cfg.Provider))
	}
	// Note: Usage is set via span.SetUsage() after creation in Opik

	return opikOpts
}
