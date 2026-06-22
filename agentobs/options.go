package agentobs

import "time"

// TraceManagerConfig configures the TraceManager behavior.
type TraceManagerConfig struct {
	// StaleTimeout is how long a trace can be inactive before being closed.
	// Default: 5 minutes.
	StaleTimeout time.Duration

	// SweepInterval is how often to check for stale traces.
	// Default: 1 minute.
	SweepInterval time.Duration

	// DefaultTags are applied to all traces.
	DefaultTags []string

	// ProjectName is the default Opik project name.
	ProjectName string

	// SanitizeContent enables content sanitization.
	// Default: true.
	SanitizeContent bool

	// SanitizeConfig is the sanitization configuration.
	// Used when SanitizeContent is true.
	SanitizeConfig SanitizeConfig

	// ExtractMedia enables media extraction from content.
	// Default: true.
	ExtractMedia bool

	// TraceNamePrefix is prepended to trace names.
	// Default: "agent.session."
	TraceNamePrefix string
}

// DefaultTraceManagerConfig returns a TraceManagerConfig with sensible defaults.
func DefaultTraceManagerConfig() TraceManagerConfig {
	return TraceManagerConfig{
		StaleTimeout:    5 * time.Minute,
		SweepInterval:   1 * time.Minute,
		SanitizeContent: true,
		SanitizeConfig:  DefaultSanitizeConfig(),
		ExtractMedia:    true,
		TraceNamePrefix: "agent.session.",
	}
}

// TraceManagerOption is a functional option for TraceManagerConfig.
type TraceManagerOption func(*TraceManagerConfig)

// WithStaleTimeout sets the stale trace timeout.
func WithStaleTimeout(d time.Duration) TraceManagerOption {
	return func(cfg *TraceManagerConfig) {
		cfg.StaleTimeout = d
	}
}

// WithSweepInterval sets the sweep interval for stale trace cleanup.
func WithSweepInterval(d time.Duration) TraceManagerOption {
	return func(cfg *TraceManagerConfig) {
		cfg.SweepInterval = d
	}
}

// WithDefaultTags sets the default tags for all traces.
func WithDefaultTags(tags ...string) TraceManagerOption {
	return func(cfg *TraceManagerConfig) {
		cfg.DefaultTags = tags
	}
}

// WithProjectName sets the default project name.
func WithProjectName(name string) TraceManagerOption {
	return func(cfg *TraceManagerConfig) {
		cfg.ProjectName = name
	}
}

// WithSanitization enables or disables content sanitization.
func WithSanitization(enabled bool) TraceManagerOption {
	return func(cfg *TraceManagerConfig) {
		cfg.SanitizeContent = enabled
	}
}

// WithSanitizeConfig sets the sanitization configuration.
func WithSanitizeConfig(sanitizeCfg SanitizeConfig) TraceManagerOption {
	return func(cfg *TraceManagerConfig) {
		cfg.SanitizeConfig = sanitizeCfg
	}
}

// WithMediaExtraction enables or disables media extraction.
func WithMediaExtraction(enabled bool) TraceManagerOption {
	return func(cfg *TraceManagerConfig) {
		cfg.ExtractMedia = enabled
	}
}

// WithTraceNamePrefix sets the prefix for trace names.
func WithTraceNamePrefix(prefix string) TraceManagerOption {
	return func(cfg *TraceManagerConfig) {
		cfg.TraceNamePrefix = prefix
	}
}

// ApplyOptions applies functional options to a config.
func (cfg *TraceManagerConfig) ApplyOptions(opts ...TraceManagerOption) {
	for _, opt := range opts {
		opt(cfg)
	}
}
