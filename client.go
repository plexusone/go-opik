package opik

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/grokify/go-comet-ml-opik/internal/api"
)

// Version is the SDK version.
const Version = "0.1.0"

// Client is the main Opik client for interacting with the Opik API.
type Client struct {
	config    *Config
	apiClient *api.Client

	// Default project name for new traces
	projectName string
}

// NewClient creates a new Opik client with the given options.
func NewClient(opts ...Option) (*Client, error) {
	options := defaultClientOptions()
	for _, opt := range opts {
		opt(options)
	}

	if err := options.config.Validate(); err != nil {
		return nil, err
	}

	// Create HTTP client with auth headers
	httpClient := options.httpClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: options.timeout,
		}
	}

	// Wrap with auth transport
	authClient := &authHTTPClient{
		client:    httpClient,
		apiKey:    options.config.APIKey,
		workspace: options.config.Workspace,
	}

	// Create the ogen client
	apiClient, err := api.NewClient(
		options.config.URL,
		api.WithClient(authClient),
	)
	if err != nil {
		return nil, err
	}

	return &Client{
		config:      options.config,
		apiClient:   apiClient,
		projectName: options.config.ProjectName,
	}, nil
}

// authHTTPClient wraps an http.Client to add authentication headers.
type authHTTPClient struct {
	client    *http.Client
	apiKey    string
	workspace string
}

// Do implements ht.Client interface.
func (c *authHTTPClient) Do(req *http.Request) (*http.Response, error) {
	// Add authentication headers
	if c.apiKey != "" {
		req.Header.Set("Authorization", c.apiKey) // No "Bearer " prefix for Opik
	}
	if c.workspace != "" {
		req.Header.Set("Comet-Workspace", c.workspace)
	}

	// Add SDK version headers
	req.Header.Set("X-OPIK-DEBUG-SDK-VERSION", Version)
	req.Header.Set("X-OPIK-DEBUG-SDK-LANG", "go")
	req.Header.Set("Accept-Encoding", "gzip")

	return c.client.Do(req)
}

// Config returns the client configuration.
func (c *Client) Config() *Config {
	return c.config
}

// ProjectName returns the default project name.
func (c *Client) ProjectName() string {
	return c.projectName
}

// SetProjectName sets the default project name.
func (c *Client) SetProjectName(name string) {
	c.projectName = name
}

// IsTracingEnabled returns true if tracing is enabled.
func (c *Client) IsTracingEnabled() bool {
	return !c.config.TracingDisabled
}

// Trace creates a new trace.
func (c *Client) Trace(ctx context.Context, name string, opts ...TraceOption) (*Trace, error) {
	if c.config.TracingDisabled {
		return nil, ErrTracingDisabled
	}

	options := defaultTraceOptions()
	for _, opt := range opts {
		opt(options)
	}

	// Use default project if not specified
	projectName := options.projectName
	if projectName == "" {
		projectName = c.projectName
	}

	// Generate trace ID
	traceUUID := uuid.New()
	traceID := traceUUID.String()

	// Prepare input/output/metadata as JSON
	var inputJSON, outputJSON, metadataJSON api.JsonListStringWrite
	if options.input != nil {
		data, _ := json.Marshal(options.input)
		inputJSON = api.JsonListStringWrite(data)
	}
	if options.output != nil {
		data, _ := json.Marshal(options.output)
		outputJSON = api.JsonListStringWrite(data)
	}
	if len(options.metadata) > 0 {
		data, _ := json.Marshal(options.metadata)
		metadataJSON = api.JsonListStringWrite(data)
	}

	startTime := time.Now()

	// Create trace request
	req := api.TraceBatchWrite{
		Traces: []api.TraceWrite{{
			ID:          api.NewOptUUID(traceUUID),
			ProjectName: api.NewOptString(projectName),
			Name:        api.NewOptString(name),
			StartTime:   startTime,
			Input:       inputJSON,
			Output:      outputJSON,
			Metadata:    metadataJSON,
			Tags:        options.tags,
		}},
	}

	// Send to API
	err := c.apiClient.CreateTraces(ctx, api.NewOptTraceBatchWrite(req))
	if err != nil {
		return nil, err
	}

	return &Trace{
		client:      c,
		id:          traceID,
		name:        name,
		projectName: projectName,
		startTime:   startTime,
		input:       options.input,
		output:      options.output,
		metadata:    options.metadata,
		tags:        options.tags,
	}, nil
}

// GetTrace retrieves a trace by ID.
func (c *Client) GetTrace(ctx context.Context, traceID string) (*Trace, error) {
	traceUUID, err := uuid.Parse(traceID)
	if err != nil {
		return nil, err
	}

	resp, err := c.apiClient.GetTraceById(ctx, api.GetTraceByIdParams{ID: traceUUID})
	if err != nil {
		return nil, err
	}

	if resp == nil {
		return nil, ErrTraceNotFound
	}

	var id, name string
	if resp.ID.Set {
		id = resp.ID.Value.String()
	}
	if resp.Name.Set {
		name = resp.Name.Value
	}

	return &Trace{
		client:    c,
		id:        id,
		name:      name,
		startTime: resp.StartTime,
	}, nil
}

// API returns the underlying ogen-generated API client for advanced usage.
func (c *Client) API() *api.Client {
	return c.apiClient
}

// Project represents an Opik project.
type Project struct {
	ID          string
	Name        string
	Description string
	CreatedAt   time.Time
	LastUpdated time.Time
}

// ListProjects lists all projects.
func (c *Client) ListProjects(ctx context.Context, page, size int) ([]*Project, error) {
	resp, err := c.apiClient.FindProjects(ctx, api.FindProjectsParams{
		Page: api.NewOptInt32(int32(page)), //nolint:gosec // G115: page values are bounded by API limits
		Size: api.NewOptInt32(int32(size)), //nolint:gosec // G115: size values are bounded by API limits
	})
	if err != nil {
		return nil, err
	}

	projects := make([]*Project, 0, len(resp.Content))
	for _, p := range resp.Content {
		project := &Project{
			Name: p.Name,
		}
		if p.ID.Set {
			project.ID = p.ID.Value.String()
		}
		if p.Description.Set {
			project.Description = p.Description.Value
		}
		if p.CreatedAt.Set {
			project.CreatedAt = p.CreatedAt.Value
		}
		if p.LastUpdatedAt.Set {
			project.LastUpdated = p.LastUpdatedAt.Value
		}
		projects = append(projects, project)
	}

	return projects, nil
}

// CreateProject creates a new project.
func (c *Client) CreateProject(ctx context.Context, name string, opts ...ProjectOption) (*Project, error) {
	options := &projectOptions{
		description: "",
	}
	for _, opt := range opts {
		opt(options)
	}

	req := api.ProjectWrite{
		Name: name,
	}
	if options.description != "" {
		req.Description = api.NewOptString(options.description)
	}

	_, err := c.apiClient.CreateProject(ctx, api.NewOptProjectWrite(req))
	if err != nil {
		return nil, err
	}

	return &Project{
		Name:        name,
		Description: options.description,
	}, nil
}

// projectOptions holds options for creating a project.
type projectOptions struct {
	description string
}

// ProjectOption configures a project creation.
type ProjectOption func(*projectOptions)

// WithProjectDescription sets the project description.
func WithProjectDescription(desc string) ProjectOption {
	return func(o *projectOptions) {
		o.description = desc
	}
}

// TraceInfo represents basic trace information.
type TraceInfo struct {
	ID        string
	Name      string
	StartTime time.Time
	EndTime   time.Time
}

// ListTraces lists recent traces.
func (c *Client) ListTraces(ctx context.Context, page, size int) ([]*TraceInfo, error) {
	resp, err := c.apiClient.GetTracesByProject(ctx, api.GetTracesByProjectParams{
		ProjectName: api.NewOptString(c.projectName),
		Page:        api.NewOptInt32(int32(page)), //nolint:gosec // G115: page values are bounded by API limits
		Size:        api.NewOptInt32(int32(size)), //nolint:gosec // G115: size values are bounded by API limits
	})
	if err != nil {
		return nil, err
	}

	traces := make([]*TraceInfo, 0, len(resp.Content))
	for _, t := range resp.Content {
		trace := &TraceInfo{
			StartTime: t.StartTime,
		}
		if t.ID.Set {
			trace.ID = t.ID.Value.String()
		}
		if t.Name.Set {
			trace.Name = t.Name.Value
		}
		if t.EndTime.Set {
			trace.EndTime = t.EndTime.Value
		}
		traces = append(traces, trace)
	}

	return traces, nil
}
