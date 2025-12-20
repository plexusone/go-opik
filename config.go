package opik

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

const (
	// DefaultCloudURL is the default URL for Opik Cloud
	DefaultCloudURL = "https://www.comet.com/opik/api"
	// DefaultLocalURL is the default URL for local Opik server
	DefaultLocalURL = "http://localhost:5173/api"
	// DefaultProjectName is the default project name
	DefaultProjectName = "Default Project"
	// DefaultWorkspace is the default workspace name
	DefaultWorkspace = "default"
	// DefaultConfigFile is the default config file path
	DefaultConfigFile = "~/.opik.config"
)

// Environment variable names
const (
	EnvURLOverride  = "OPIK_URL_OVERRIDE"
	EnvAPIKey       = "OPIK_API_KEY" //nolint:gosec // G101: This is an environment variable name, not a credential
	EnvWorkspace    = "OPIK_WORKSPACE"
	EnvProjectName  = "OPIK_PROJECT_NAME"
	EnvTraceDisable = "OPIK_TRACK_DISABLE"
)

// Config holds the configuration for the Opik client.
type Config struct {
	// URL is the Opik API endpoint URL.
	// Defaults to DefaultCloudURL if APIKey is set, otherwise DefaultLocalURL.
	URL string

	// APIKey is the API key for authentication with Opik Cloud.
	// Not required for local/self-hosted instances.
	APIKey string

	// Workspace is the workspace name for Opik Cloud.
	// Required for Opik Cloud, ignored for local instances.
	Workspace string

	// ProjectName is the default project name for traces.
	ProjectName string

	// TracingDisabled disables tracing globally when set to true.
	TracingDisabled bool

	// CheckTLSCertificate enables TLS certificate verification.
	// Defaults to true.
	CheckTLSCertificate bool
}

// NewConfig creates a new Config with default values.
func NewConfig() *Config {
	return &Config{
		URL:                 "",
		APIKey:              "",
		Workspace:           DefaultWorkspace,
		ProjectName:         DefaultProjectName,
		TracingDisabled:     false,
		CheckTLSCertificate: true,
	}
}

// LoadConfig loads configuration from environment variables and config file.
// Priority order (highest to lowest):
// 1. Explicitly set values (via options)
// 2. Environment variables
// 3. Config file (~/.opik.config)
// 4. Default values
func LoadConfig() *Config {
	cfg := NewConfig()

	// Load from config file first (lowest priority of external sources)
	cfg.loadFromFile()

	// Load from environment variables (overrides config file)
	cfg.loadFromEnv()

	// Set default URL based on whether API key is present
	if cfg.URL == "" {
		if cfg.APIKey != "" {
			cfg.URL = DefaultCloudURL
		} else {
			cfg.URL = DefaultLocalURL
		}
	}

	return cfg
}

// loadFromEnv loads configuration from environment variables.
func (c *Config) loadFromEnv() {
	if url := os.Getenv(EnvURLOverride); url != "" {
		c.URL = url
	}
	if apiKey := os.Getenv(EnvAPIKey); apiKey != "" {
		c.APIKey = apiKey
	}
	if workspace := os.Getenv(EnvWorkspace); workspace != "" {
		c.Workspace = workspace
	}
	if projectName := os.Getenv(EnvProjectName); projectName != "" {
		c.ProjectName = projectName
	}
	if disable := os.Getenv(EnvTraceDisable); disable != "" {
		c.TracingDisabled = strings.ToLower(disable) == "true" || disable == "1"
	}
}

// loadFromFile loads configuration from the config file.
func (c *Config) loadFromFile() {
	configPath := DefaultConfigFile
	if strings.HasPrefix(configPath, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return
		}
		configPath = filepath.Join(home, configPath[1:])
	}

	file, err := os.Open(configPath)
	if err != nil {
		return // Config file doesn't exist, which is fine
	}
	defer file.Close()

	// Parse INI-style config file
	scanner := bufio.NewScanner(file)
	section := ""
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		// Section header
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.ToLower(strings.TrimSpace(line[1 : len(line)-1]))
			continue
		}

		// Key-value pair
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Only process [opik] section or no section
		if section != "" && section != "opik" {
			continue
		}

		switch strings.ToLower(key) {
		case "url_override", "url":
			if c.URL == "" {
				c.URL = value
			}
		case "api_key", "apikey":
			if c.APIKey == "" {
				c.APIKey = value
			}
		case "workspace":
			if c.Workspace == DefaultWorkspace {
				c.Workspace = value
			}
		case "project_name", "projectname":
			if c.ProjectName == DefaultProjectName {
				c.ProjectName = value
			}
		}
	}
}

// IsCloud returns true if the configuration is for Opik Cloud.
func (c *Config) IsCloud() bool {
	return c.APIKey != "" || strings.Contains(c.URL, "comet.com")
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c.URL == "" {
		return ErrMissingURL
	}
	if c.IsCloud() && c.APIKey == "" {
		return ErrMissingAPIKey
	}
	return nil
}

// SaveConfig saves the configuration to the config file.
func SaveConfig(cfg *Config) error {
	configPath := DefaultConfigFile
	if strings.HasPrefix(configPath, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		configPath = filepath.Join(home, configPath[1:])
	}

	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	if _, err := w.WriteString("[opik]\n"); err != nil {
		return err
	}
	if cfg.URL != "" {
		if _, err := w.WriteString("url_override = " + cfg.URL + "\n"); err != nil {
			return err
		}
	}
	if cfg.APIKey != "" {
		if _, err := w.WriteString("api_key = " + cfg.APIKey + "\n"); err != nil {
			return err
		}
	}
	if cfg.Workspace != "" {
		if _, err := w.WriteString("workspace = " + cfg.Workspace + "\n"); err != nil {
			return err
		}
	}
	if cfg.ProjectName != "" {
		if _, err := w.WriteString("project_name = " + cfg.ProjectName + "\n"); err != nil {
			return err
		}
	}
	return w.Flush()
}
