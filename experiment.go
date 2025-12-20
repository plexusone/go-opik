package opik

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/grokify/go-comet-ml-opik/internal/api"
)

// Experiment represents an experiment in Opik for evaluating models.
type Experiment struct {
	client      *Client
	id          string
	name        string
	datasetName string
	metadata    map[string]any
}

// ExperimentItem represents an item result in an experiment.
type ExperimentItem struct {
	ID            string
	ExperimentID  string
	DatasetItemID string
	TraceID       string
	Input         any
	Output        any
}

// ExperimentStatus represents the status of an experiment.
type ExperimentStatus string

const (
	ExperimentStatusRunning   ExperimentStatus = "running"
	ExperimentStatusCompleted ExperimentStatus = "completed"
	ExperimentStatusCancelled ExperimentStatus = "cancelled"
)

// ExperimentType represents the type of an experiment.
type ExperimentType string

const (
	ExperimentTypeRegular   ExperimentType = "regular"
	ExperimentTypeTrial     ExperimentType = "trial"
	ExperimentTypeMiniBatch ExperimentType = "mini-batch"
)

// ID returns the experiment ID.
func (e *Experiment) ID() string {
	return e.id
}

// Name returns the experiment name.
func (e *Experiment) Name() string {
	return e.name
}

// DatasetName returns the dataset name.
func (e *Experiment) DatasetName() string {
	return e.datasetName
}

// Metadata returns the experiment metadata.
func (e *Experiment) Metadata() map[string]any {
	return e.metadata
}

// ExperimentOption is a functional option for configuring an Experiment.
type ExperimentOption func(*experimentOptions)

type experimentOptions struct {
	name           string
	metadata       map[string]any
	experimentType ExperimentType
	status         ExperimentStatus
}

// WithExperimentName sets the name for the experiment.
func WithExperimentName(name string) ExperimentOption {
	return func(o *experimentOptions) {
		o.name = name
	}
}

// WithExperimentMetadata sets the metadata for the experiment.
func WithExperimentMetadata(metadata map[string]any) ExperimentOption {
	return func(o *experimentOptions) {
		o.metadata = metadata
	}
}

// WithExperimentType sets the type for the experiment.
func WithExperimentType(experimentType ExperimentType) ExperimentOption {
	return func(o *experimentOptions) {
		o.experimentType = experimentType
	}
}

// WithExperimentStatus sets the initial status for the experiment.
func WithExperimentStatus(status ExperimentStatus) ExperimentOption {
	return func(o *experimentOptions) {
		o.status = status
	}
}

// CreateExperiment creates a new experiment for a dataset.
func (c *Client) CreateExperiment(ctx context.Context, datasetName string, opts ...ExperimentOption) (*Experiment, error) {
	options := &experimentOptions{
		metadata:       make(map[string]any),
		experimentType: ExperimentTypeRegular,
		status:         ExperimentStatusRunning,
	}
	for _, opt := range opts {
		opt(options)
	}

	experimentUUID := uuid.New()

	var metadataJSON api.JsonListStringWrite
	if len(options.metadata) > 0 {
		data, _ := json.Marshal(options.metadata)
		metadataJSON = api.JsonListStringWrite(data)
	}

	req := api.ExperimentWrite{
		ID:          api.NewOptUUID(experimentUUID),
		DatasetName: datasetName,
		Name:        api.NewOptString(options.name),
		Metadata:    metadataJSON,
		Type:        api.NewOptExperimentWriteType(api.ExperimentWriteType(options.experimentType)),
		Status:      api.NewOptExperimentWriteStatus(api.ExperimentWriteStatus(options.status)),
	}

	resp, err := c.apiClient.CreateExperiment(ctx, api.NewOptExperimentWrite(req))
	if err != nil {
		return nil, err
	}

	_ = resp // Location header contains the ID

	name := options.name
	if name == "" {
		name = experimentUUID.String()
	}

	return &Experiment{
		client:      c,
		id:          experimentUUID.String(),
		name:        name,
		datasetName: datasetName,
		metadata:    options.metadata,
	}, nil
}

// GetExperiment retrieves an experiment by ID.
func (c *Client) GetExperiment(ctx context.Context, experimentID string) (*Experiment, error) {
	experimentUUID, err := uuid.Parse(experimentID)
	if err != nil {
		return nil, err
	}

	resp, err := c.apiClient.GetExperimentById(ctx, api.GetExperimentByIdParams{ID: experimentUUID})
	if err != nil {
		return nil, err
	}

	// Handle the response union type
	switch v := resp.(type) {
	case *api.ExperimentPublic:
		var id, name, datasetName string
		if v.ID.Set {
			id = v.ID.Value.String()
		}
		if v.Name.Set {
			name = v.Name.Value
		}
		datasetName = v.DatasetName

		return &Experiment{
			client:      c,
			id:          id,
			name:        name,
			datasetName: datasetName,
		}, nil
	default:
		return nil, ErrExperimentNotFound
	}
}

// ListExperiments lists experiments for a dataset.
func (c *Client) ListExperiments(ctx context.Context, datasetID string, page, size int) ([]*Experiment, error) {
	datasetUUID, err := uuid.Parse(datasetID)
	if err != nil {
		return nil, err
	}

	params := api.FindExperimentsParams{
		DatasetId: api.NewOptUUID(datasetUUID),
		Page:      api.NewOptInt32(int32(page)), //nolint:gosec // G115: page values are bounded by API limits
		Size:      api.NewOptInt32(int32(size)), //nolint:gosec // G115: size values are bounded by API limits
	}

	resp, err := c.apiClient.FindExperiments(ctx, params)
	if err != nil {
		return nil, err
	}

	// Handle the response union type
	switch v := resp.(type) {
	case *api.ExperimentPagePublic:
		experiments := make([]*Experiment, 0, len(v.Content))
		for _, exp := range v.Content {
			var id, name string
			if exp.ID.Set {
				id = exp.ID.Value.String()
			}
			if exp.Name.Set {
				name = exp.Name.Value
			}

			experiments = append(experiments, &Experiment{
				client:      c,
				id:          id,
				name:        name,
				datasetName: exp.DatasetName,
			})
		}
		return experiments, nil
	default:
		return []*Experiment{}, nil
	}
}

// DeleteExperiment deletes an experiment by ID.
func (c *Client) DeleteExperiment(ctx context.Context, experimentID string) error {
	experimentUUID, err := uuid.Parse(experimentID)
	if err != nil {
		return err
	}

	req := api.DeleteIdsHolder{
		Ids: []uuid.UUID{experimentUUID},
	}

	return c.apiClient.DeleteExperimentsById(ctx, api.NewOptDeleteIdsHolder(req))
}

// ExperimentItemOption is a functional option for configuring an ExperimentItem.
type ExperimentItemOption func(*experimentItemOptions)

type experimentItemOptions struct {
	input  any
	output any
}

// WithExperimentItemInput sets the input for the experiment item.
func WithExperimentItemInput(input any) ExperimentItemOption {
	return func(o *experimentItemOptions) {
		o.input = input
	}
}

// WithExperimentItemOutput sets the output for the experiment item.
func WithExperimentItemOutput(output any) ExperimentItemOption {
	return func(o *experimentItemOptions) {
		o.output = output
	}
}

// LogItem logs a result for a dataset item in this experiment.
func (e *Experiment) LogItem(ctx context.Context, datasetItemID, traceID string, opts ...ExperimentItemOption) error {
	options := &experimentItemOptions{}
	for _, opt := range opts {
		opt(options)
	}

	experimentUUID, err := uuid.Parse(e.id)
	if err != nil {
		return err
	}

	datasetItemUUID, err := uuid.Parse(datasetItemID)
	if err != nil {
		return err
	}

	traceUUID, err := uuid.Parse(traceID)
	if err != nil {
		return err
	}

	itemUUID := uuid.New()

	var inputJSON, outputJSON api.JsonListString
	if options.input != nil {
		data, _ := json.Marshal(options.input)
		inputJSON = api.JsonListString(data)
	}
	if options.output != nil {
		data, _ := json.Marshal(options.output)
		outputJSON = api.JsonListString(data)
	}

	req := api.ExperimentItemsBatch{
		ExperimentItems: []api.ExperimentItem{{
			ID:            api.NewOptUUID(itemUUID),
			ExperimentID:  experimentUUID,
			DatasetItemID: datasetItemUUID,
			TraceID:       traceUUID,
			Input:         inputJSON,
			Output:        outputJSON,
		}},
	}

	return e.client.apiClient.CreateExperimentItems(ctx, api.NewOptExperimentItemsBatch(req))
}

// Complete marks the experiment as completed.
func (e *Experiment) Complete(ctx context.Context) error {
	return e.updateStatus(ctx, ExperimentStatusCompleted)
}

// Cancel marks the experiment as cancelled.
func (e *Experiment) Cancel(ctx context.Context) error {
	return e.updateStatus(ctx, ExperimentStatusCancelled)
}

func (e *Experiment) updateStatus(ctx context.Context, status ExperimentStatus) error {
	experimentUUID, err := uuid.Parse(e.id)
	if err != nil {
		return err
	}

	req := api.ExperimentUpdate{
		Status: api.NewOptExperimentUpdateStatus(api.ExperimentUpdateStatus(status)),
	}

	_, err = e.client.apiClient.UpdateExperiment(ctx, api.NewOptExperimentUpdate(req), api.UpdateExperimentParams{
		ID: experimentUUID,
	})
	return err
}

// Delete deletes this experiment.
func (e *Experiment) Delete(ctx context.Context) error {
	return e.client.DeleteExperiment(ctx, e.id)
}
