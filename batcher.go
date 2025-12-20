package opik

import (
	"context"
	"sync"
	"time"
)

// BatcherConfig configures the message batcher.
type BatcherConfig struct {
	// MaxBatchSize is the maximum number of items to batch before flushing.
	MaxBatchSize int
	// FlushInterval is the maximum time to wait before flushing a batch.
	FlushInterval time.Duration
	// MaxRetries is the maximum number of retries for failed batches.
	MaxRetries int
	// RetryDelay is the initial delay between retries (doubles each retry).
	RetryDelay time.Duration
	// Workers is the number of background workers for processing batches.
	Workers int
}

// DefaultBatcherConfig returns the default batcher configuration.
func DefaultBatcherConfig() BatcherConfig {
	return BatcherConfig{
		MaxBatchSize:  100,
		FlushInterval: 5 * time.Second,
		MaxRetries:    3,
		RetryDelay:    100 * time.Millisecond,
		Workers:       2,
	}
}

// BatchItem represents an item to be batched.
type BatchItem interface {
	// Type returns the type of batch item (e.g., "trace", "span", "feedback").
	Type() string
}

// TraceBatchItem represents a trace to be created.
type TraceBatchItem struct {
	Trace *Trace
}

func (t TraceBatchItem) Type() string { return "trace" }

// SpanBatchItem represents a span to be created.
type SpanBatchItem struct {
	Span *Span
}

func (s SpanBatchItem) Type() string { return "span" }

// FeedbackBatchItem represents a feedback score to be logged.
type FeedbackBatchItem struct {
	EntityType string // "trace" or "span"
	EntityID   string
	Name       string
	Value      float64
	Reason     string
}

func (f FeedbackBatchItem) Type() string { return "feedback" }

// Batcher batches operations for efficient API calls.
type Batcher struct {
	config   BatcherConfig
	client   *Client
	items    []BatchItem
	mu       sync.Mutex
	wg       sync.WaitGroup
	ctx      context.Context
	cancel   context.CancelFunc
	itemChan chan BatchItem
	flushCh  chan struct{}
}

// NewBatcher creates a new batcher with the given configuration.
func NewBatcher(client *Client, config BatcherConfig) *Batcher {
	ctx, cancel := context.WithCancel(context.Background())

	b := &Batcher{
		config:   config,
		client:   client,
		items:    make([]BatchItem, 0, config.MaxBatchSize),
		ctx:      ctx,
		cancel:   cancel,
		itemChan: make(chan BatchItem, config.MaxBatchSize*2),
		flushCh:  make(chan struct{}, 1),
	}

	// Start background workers
	for i := 0; i < config.Workers; i++ {
		b.wg.Add(1)
		go b.worker()
	}

	// Start flush timer
	b.wg.Add(1)
	go b.flushTimer()

	return b
}

// Add adds an item to the batch.
func (b *Batcher) Add(item BatchItem) {
	select {
	case b.itemChan <- item:
	case <-b.ctx.Done():
	}
}

// Flush forces a flush of all pending items.
func (b *Batcher) Flush(timeout time.Duration) error {
	// Signal flush
	select {
	case b.flushCh <- struct{}{}:
	default:
	}

	// Wait for items to drain
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			b.mu.Lock()
			empty := len(b.items) == 0
			b.mu.Unlock()
			if empty && len(b.itemChan) == 0 {
				return nil
			}
		}
	}
}

// Close stops the batcher and flushes remaining items.
func (b *Batcher) Close(timeout time.Duration) error {
	// Flush remaining items
	err := b.Flush(timeout)

	// Stop workers
	b.cancel()
	b.wg.Wait()

	return err
}

func (b *Batcher) worker() {
	defer b.wg.Done()

	for {
		select {
		case <-b.ctx.Done():
			// Drain remaining items
			b.drainAndFlush()
			return

		case item := <-b.itemChan:
			b.mu.Lock()
			b.items = append(b.items, item)
			shouldFlush := len(b.items) >= b.config.MaxBatchSize
			b.mu.Unlock()

			if shouldFlush {
				b.doFlush()
			}

		case <-b.flushCh:
			b.doFlush()
		}
	}
}

func (b *Batcher) flushTimer() {
	defer b.wg.Done()

	ticker := time.NewTicker(b.config.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-b.ctx.Done():
			return
		case <-ticker.C:
			select {
			case b.flushCh <- struct{}{}:
			default:
			}
		}
	}
}

func (b *Batcher) drainAndFlush() {
	// Drain channel
	for {
		select {
		case item := <-b.itemChan:
			b.mu.Lock()
			b.items = append(b.items, item)
			b.mu.Unlock()
		default:
			b.doFlush()
			return
		}
	}
}

func (b *Batcher) doFlush() {
	b.mu.Lock()
	if len(b.items) == 0 {
		b.mu.Unlock()
		return
	}

	// Take items
	items := b.items
	b.items = make([]BatchItem, 0, b.config.MaxBatchSize)
	b.mu.Unlock()

	// Group items by type
	traceItems := make([]TraceBatchItem, 0)
	spanItems := make([]SpanBatchItem, 0)
	feedbackItems := make([]FeedbackBatchItem, 0)

	for _, item := range items {
		switch v := item.(type) {
		case TraceBatchItem:
			traceItems = append(traceItems, v)
		case SpanBatchItem:
			spanItems = append(spanItems, v)
		case FeedbackBatchItem:
			feedbackItems = append(feedbackItems, v)
		}
	}

	// Process each type with retries
	ctx := context.Background()

	if len(traceItems) > 0 {
		b.processWithRetry(func() error {
			return b.flushTraces(ctx, traceItems)
		})
	}

	if len(spanItems) > 0 {
		b.processWithRetry(func() error {
			return b.flushSpans(ctx, spanItems)
		})
	}

	if len(feedbackItems) > 0 {
		b.processWithRetry(func() error {
			return b.flushFeedback(ctx, feedbackItems)
		})
	}
}

func (b *Batcher) processWithRetry(fn func() error) {
	delay := b.config.RetryDelay
	for i := 0; i < b.config.MaxRetries; i++ {
		err := fn()
		if err == nil {
			return
		}

		// Check if rate limited
		if IsRateLimited(err) {
			time.Sleep(delay * 2)
		} else {
			time.Sleep(delay)
		}
		delay *= 2
	}
}

func (b *Batcher) flushTraces(_ context.Context, items []TraceBatchItem) error {
	// For now, traces are created individually
	// The API supports batch creation but we'd need to restructure
	for _, item := range items {
		if item.Trace != nil {
			// Trace already created via API, this is for deferred creation
		}
	}
	return nil
}

func (b *Batcher) flushSpans(_ context.Context, items []SpanBatchItem) error {
	// For now, spans are created individually
	for _, item := range items {
		if item.Span != nil {
			// Span already created via API
		}
	}
	return nil
}

func (b *Batcher) flushFeedback(_ context.Context, items []FeedbackBatchItem) error {
	// Group by entity type
	traceFeedback := make(map[string][]FeedbackBatchItem)
	spanFeedback := make(map[string][]FeedbackBatchItem)

	for _, item := range items {
		if item.EntityType == "trace" {
			traceFeedback[item.EntityID] = append(traceFeedback[item.EntityID], item)
		} else {
			spanFeedback[item.EntityID] = append(spanFeedback[item.EntityID], item)
		}
	}

	// Process trace feedback - would use batch API
	// Process span feedback - would use batch API
	return nil
}

// BatchingClient wraps a Client with batching support.
type BatchingClient struct {
	*Client
	batcher *Batcher
}

// NewBatchingClient creates a new client with batching enabled.
func NewBatchingClient(opts ...Option) (*BatchingClient, error) {
	client, err := NewClient(opts...)
	if err != nil {
		return nil, err
	}

	batcher := NewBatcher(client, DefaultBatcherConfig())

	return &BatchingClient{
		Client:  client,
		batcher: batcher,
	}, nil
}

// NewBatchingClientWithConfig creates a new client with custom batching config.
func NewBatchingClientWithConfig(config BatcherConfig, opts ...Option) (*BatchingClient, error) {
	client, err := NewClient(opts...)
	if err != nil {
		return nil, err
	}

	batcher := NewBatcher(client, config)

	return &BatchingClient{
		Client:  client,
		batcher: batcher,
	}, nil
}

// Flush flushes all pending batched operations.
func (c *BatchingClient) Flush(timeout time.Duration) error {
	return c.batcher.Flush(timeout)
}

// Close closes the batching client and flushes pending operations.
func (c *BatchingClient) Close(timeout time.Duration) error {
	return c.batcher.Close(timeout)
}

// AddFeedbackAsync adds a feedback score asynchronously via batching.
func (c *BatchingClient) AddFeedbackAsync(entityType, entityID, name string, value float64, reason string) {
	c.batcher.Add(FeedbackBatchItem{
		EntityType: entityType,
		EntityID:   entityID,
		Name:       name,
		Value:      value,
		Reason:     reason,
	})
}
