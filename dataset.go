package opik

import (
	"context"
	"encoding/json"

	"github.com/go-faster/jx"
	"github.com/google/uuid"
	"github.com/grokify/go-comet-ml-opik/internal/api"
)

// Dataset represents a dataset in Opik for storing test data.
type Dataset struct {
	client      *Client
	id          string
	name        string
	description string
	tags        []string
}

// DatasetItem represents an item in a dataset.
type DatasetItem struct {
	ID      string
	TraceID string
	SpanID  string
	Data    map[string]any
	Tags    []string
}

// ID returns the dataset ID.
func (d *Dataset) ID() string {
	return d.id
}

// Name returns the dataset name.
func (d *Dataset) Name() string {
	return d.name
}

// Description returns the dataset description.
func (d *Dataset) Description() string {
	return d.description
}

// Tags returns the dataset tags.
func (d *Dataset) Tags() []string {
	return d.tags
}

// DatasetOption is a functional option for configuring a Dataset.
type DatasetOption func(*datasetOptions)

type datasetOptions struct {
	description string
	tags        []string
}

// WithDatasetDescription sets the description for the dataset.
func WithDatasetDescription(description string) DatasetOption {
	return func(o *datasetOptions) {
		o.description = description
	}
}

// WithDatasetTags sets the tags for the dataset.
func WithDatasetTags(tags ...string) DatasetOption {
	return func(o *datasetOptions) {
		o.tags = tags
	}
}

// CreateDataset creates a new dataset.
func (c *Client) CreateDataset(ctx context.Context, name string, opts ...DatasetOption) (*Dataset, error) {
	options := &datasetOptions{
		tags: []string{},
	}
	for _, opt := range opts {
		opt(options)
	}

	datasetUUID := uuid.New()

	req := api.DatasetWrite{
		ID:          api.NewOptUUID(datasetUUID),
		Name:        name,
		Description: api.NewOptString(options.description),
		Tags:        options.tags,
	}

	resp, err := c.apiClient.CreateDataset(ctx, api.NewOptDatasetWrite(req))
	if err != nil {
		return nil, err
	}

	// The response contains the location header with the ID
	_ = resp // Location header contains the ID

	return &Dataset{
		client:      c,
		id:          datasetUUID.String(),
		name:        name,
		description: options.description,
		tags:        options.tags,
	}, nil
}

// GetDataset retrieves a dataset by ID.
func (c *Client) GetDataset(ctx context.Context, datasetID string) (*Dataset, error) {
	datasetUUID, err := uuid.Parse(datasetID)
	if err != nil {
		return nil, err
	}

	resp, err := c.apiClient.GetDatasetById(ctx, api.GetDatasetByIdParams{ID: datasetUUID})
	if err != nil {
		return nil, err
	}

	if resp == nil {
		return nil, ErrDatasetNotFound
	}

	var id, name, description string
	if resp.ID.Set {
		id = resp.ID.Value.String()
	}
	name = resp.Name
	if resp.Description.Set {
		description = resp.Description.Value
	}

	return &Dataset{
		client:      c,
		id:          id,
		name:        name,
		description: description,
		tags:        resp.Tags,
	}, nil
}

// GetDatasetByName retrieves a dataset by name.
func (c *Client) GetDatasetByName(ctx context.Context, name string) (*Dataset, error) {
	req := api.DatasetIdentifierPublic{
		DatasetName: name,
	}

	resp, err := c.apiClient.GetDatasetByIdentifier(ctx, api.NewOptDatasetIdentifierPublic(req))
	if err != nil {
		return nil, err
	}

	if resp == nil {
		return nil, ErrDatasetNotFound
	}

	var id, description string
	if resp.ID.Set {
		id = resp.ID.Value.String()
	}
	if resp.Description.Set {
		description = resp.Description.Value
	}

	return &Dataset{
		client:      c,
		id:          id,
		name:        resp.Name,
		description: description,
		tags:        resp.Tags,
	}, nil
}

// ListDatasets lists all datasets.
//
//nolint:dupl // Similar structure to ListPrompts is intentional for consistency
func (c *Client) ListDatasets(ctx context.Context, page, size int) ([]*Dataset, error) {
	params := api.FindDatasetsParams{
		Page: api.NewOptInt32(int32(page)), //nolint:gosec // G115: page values are bounded by API limits
		Size: api.NewOptInt32(int32(size)), //nolint:gosec // G115: size values are bounded by API limits
	}

	resp, err := c.apiClient.FindDatasets(ctx, params)
	if err != nil {
		return nil, err
	}

	if resp == nil {
		return []*Dataset{}, nil
	}

	datasets := make([]*Dataset, 0, len(resp.Content))
	for _, d := range resp.Content {
		var id, description string
		if d.ID.Set {
			id = d.ID.Value.String()
		}
		if d.Description.Set {
			description = d.Description.Value
		}

		datasets = append(datasets, &Dataset{
			client:      c,
			id:          id,
			name:        d.Name,
			description: description,
			tags:        d.Tags,
		})
	}

	return datasets, nil
}

// DeleteDataset deletes a dataset by ID.
func (c *Client) DeleteDataset(ctx context.Context, datasetID string) error {
	datasetUUID, err := uuid.Parse(datasetID)
	if err != nil {
		return err
	}

	return c.apiClient.DeleteDataset(ctx, api.DeleteDatasetParams{ID: datasetUUID})
}

// InsertItem inserts a single item into the dataset.
func (d *Dataset) InsertItem(ctx context.Context, data map[string]any, opts ...DatasetItemOption) error {
	return d.InsertItems(ctx, []map[string]any{data}, opts...)
}

// DatasetItemOption is a functional option for configuring a DatasetItem.
type DatasetItemOption func(*datasetItemOptions)

type datasetItemOptions struct {
	tags []string
}

// WithDatasetItemTags sets the tags for the dataset item.
func WithDatasetItemTags(tags ...string) DatasetItemOption {
	return func(o *datasetItemOptions) {
		o.tags = tags
	}
}

// mapToJsonNode converts a map[string]any to api.JsonNode (map[string]jx.Raw).
func mapToJsonNode(m map[string]any) api.JsonNode {
	if m == nil {
		return nil
	}
	result := make(api.JsonNode, len(m))
	for k, v := range m {
		data, _ := json.Marshal(v)
		result[k] = jx.Raw(data)
	}
	return result
}

// jsonNodeToMap converts api.JsonNode (map[string]jx.Raw) to map[string]any.
func jsonNodeToMap(node api.JsonNode) map[string]any {
	if node == nil {
		return nil
	}
	result := make(map[string]any, len(node))
	for k, v := range node {
		var val any
		_ = json.Unmarshal([]byte(v), &val)
		result[k] = val
	}
	return result
}

// InsertItems inserts multiple items into the dataset.
func (d *Dataset) InsertItems(ctx context.Context, items []map[string]any, opts ...DatasetItemOption) error {
	options := &datasetItemOptions{
		tags: []string{},
	}
	for _, opt := range opts {
		opt(options)
	}

	datasetUUID, err := uuid.Parse(d.id)
	if err != nil {
		return err
	}

	apiItems := make([]api.DatasetItemWrite, 0, len(items))
	for _, item := range items {
		itemUUID := uuid.New()
		apiItems = append(apiItems, api.DatasetItemWrite{
			ID:     api.NewOptUUID(itemUUID),
			Source: api.DatasetItemWriteSourceSdk,
			Data:   mapToJsonNode(item),
			Tags:   options.tags,
		})
	}

	req := api.DatasetItemBatchWrite{
		DatasetID: api.NewOptUUID(datasetUUID),
		Items:     apiItems,
	}

	return d.client.apiClient.CreateOrUpdateDatasetItems(ctx, api.NewOptDatasetItemBatchWrite(req))
}

// GetItems retrieves items from the dataset.
func (d *Dataset) GetItems(ctx context.Context, page, size int) ([]DatasetItem, error) {
	datasetUUID, err := uuid.Parse(d.id)
	if err != nil {
		return nil, err
	}

	params := api.GetDatasetItemsParams{
		ID:   datasetUUID,
		Page: api.NewOptInt32(int32(page)), //nolint:gosec // G115: page values are bounded by API limits
		Size: api.NewOptInt32(int32(size)), //nolint:gosec // G115: size values are bounded by API limits
	}

	resp, err := d.client.apiClient.GetDatasetItems(ctx, params)
	if err != nil {
		return nil, err
	}

	if resp == nil {
		return []DatasetItem{}, nil
	}

	items := make([]DatasetItem, 0, len(resp.Content))
	for _, item := range resp.Content {
		var id, traceID, spanID string
		if item.ID.Set {
			id = item.ID.Value.String()
		}
		if item.TraceID.Set {
			traceID = item.TraceID.Value.String()
		}
		if item.SpanID.Set {
			spanID = item.SpanID.Value.String()
		}

		items = append(items, DatasetItem{
			ID:      id,
			TraceID: traceID,
			SpanID:  spanID,
			Data:    jsonNodeToMap(item.Data),
			Tags:    item.Tags,
		})
	}

	return items, nil
}

// Delete deletes this dataset.
func (d *Dataset) Delete(ctx context.Context) error {
	return d.client.DeleteDataset(ctx, d.id)
}
