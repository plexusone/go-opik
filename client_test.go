package opik

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/grokify/go-comet-ml-opik/testutil"
)

func TestNewClient(t *testing.T) {
	t.Run("with defaults", func(t *testing.T) {
		// Create mock server
		ms := testutil.NewMockServer()
		defer ms.Close()

		client, err := NewClient(
			WithURL(ms.URL()),
			WithAPIKey("test-key"),
			WithWorkspace("test-workspace"),
		)
		if err != nil {
			t.Fatalf("NewClient error: %v", err)
		}

		if client == nil {
			t.Fatal("client should not be nil")
		}
		if client.Config() == nil {
			t.Error("Config() should not be nil")
		}
	})

	t.Run("with project name", func(t *testing.T) {
		ms := testutil.NewMockServer()
		defer ms.Close()

		client, err := NewClient(
			WithURL(ms.URL()),
			WithAPIKey("test-key"),
			WithProjectName("my-project"),
		)
		if err != nil {
			t.Fatalf("NewClient error: %v", err)
		}

		if client.ProjectName() != "my-project" {
			t.Errorf("ProjectName() = %q, want %q", client.ProjectName(), "my-project")
		}
	})

	t.Run("with custom timeout", func(t *testing.T) {
		ms := testutil.NewMockServer()
		defer ms.Close()

		client, err := NewClient(
			WithURL(ms.URL()),
			WithAPIKey("test-key"),
			WithTimeout(5*time.Second),
		)
		if err != nil {
			t.Fatalf("NewClient error: %v", err)
		}

		if client == nil {
			t.Fatal("client should not be nil")
		}
	})

	t.Run("with tracing disabled", func(t *testing.T) {
		ms := testutil.NewMockServer()
		defer ms.Close()

		client, err := NewClient(
			WithURL(ms.URL()),
			WithAPIKey("test-key"),
			WithTracingDisabled(true),
		)
		if err != nil {
			t.Fatalf("NewClient error: %v", err)
		}

		if client.IsTracingEnabled() {
			t.Error("IsTracingEnabled() should be false")
		}
	})

	t.Run("invalid config", func(t *testing.T) {
		// Without URL, should fail validation
		_, err := NewClient(
			WithURL(""),
		)
		if err == nil {
			t.Error("expected error for invalid config")
		}
	})
}

func TestClientConfig(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	client, err := NewClient(
		WithURL(ms.URL()),
		WithAPIKey("my-api-key"),
		WithWorkspace("my-workspace"),
		WithProjectName("my-project"),
	)
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	config := client.Config()

	if config.URL != ms.URL() {
		t.Errorf("Config.URL = %q, want %q", config.URL, ms.URL())
	}
	if config.APIKey != "my-api-key" {
		t.Errorf("Config.APIKey = %q, want %q", config.APIKey, "my-api-key")
	}
	if config.Workspace != "my-workspace" {
		t.Errorf("Config.Workspace = %q, want %q", config.Workspace, "my-workspace")
	}
	if config.ProjectName != "my-project" {
		t.Errorf("Config.ProjectName = %q, want %q", config.ProjectName, "my-project")
	}
}

func TestClientSetProjectName(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	client, err := NewClient(
		WithURL(ms.URL()),
		WithAPIKey("test-key"),
		WithProjectName("initial"),
	)
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	if client.ProjectName() != "initial" {
		t.Errorf("initial ProjectName() = %q, want %q", client.ProjectName(), "initial")
	}

	client.SetProjectName("updated")

	if client.ProjectName() != "updated" {
		t.Errorf("updated ProjectName() = %q, want %q", client.ProjectName(), "updated")
	}
}

func TestClientIsTracingEnabled(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	t.Run("enabled by default", func(t *testing.T) {
		client, _ := NewClient(
			WithURL(ms.URL()),
			WithAPIKey("test-key"),
		)
		if !client.IsTracingEnabled() {
			t.Error("tracing should be enabled by default")
		}
	})

	t.Run("can be disabled", func(t *testing.T) {
		client, _ := NewClient(
			WithURL(ms.URL()),
			WithAPIKey("test-key"),
			WithTracingDisabled(true),
		)
		if client.IsTracingEnabled() {
			t.Error("tracing should be disabled")
		}
	})
}

func TestAuthHTTPClient(t *testing.T) {
	var capturedHeaders http.Header

	// Create a test server that captures headers
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeaders = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	authClient := &authHTTPClient{
		client:    http.DefaultClient,
		apiKey:    "test-api-key",
		workspace: "test-workspace",
	}

	req, _ := http.NewRequest("GET", ts.URL, nil)
	_, err := authClient.Do(req)
	if err != nil {
		t.Fatalf("Do error: %v", err)
	}

	t.Run("sets Authorization header", func(t *testing.T) {
		auth := capturedHeaders.Get("Authorization")
		if auth != "test-api-key" {
			t.Errorf("Authorization = %q, want %q", auth, "test-api-key")
		}
	})

	t.Run("sets Comet-Workspace header", func(t *testing.T) {
		workspace := capturedHeaders.Get("Comet-Workspace")
		if workspace != "test-workspace" {
			t.Errorf("Comet-Workspace = %q, want %q", workspace, "test-workspace")
		}
	})

	t.Run("sets SDK version header", func(t *testing.T) {
		version := capturedHeaders.Get("X-OPIK-DEBUG-SDK-VERSION")
		if version != Version {
			t.Errorf("X-OPIK-DEBUG-SDK-VERSION = %q, want %q", version, Version)
		}
	})

	t.Run("sets SDK lang header", func(t *testing.T) {
		lang := capturedHeaders.Get("X-OPIK-DEBUG-SDK-LANG")
		if lang != "go" {
			t.Errorf("X-OPIK-DEBUG-SDK-LANG = %q, want %q", lang, "go")
		}
	})

	t.Run("sets Accept-Encoding header", func(t *testing.T) {
		encoding := capturedHeaders.Get("Accept-Encoding")
		if encoding != "gzip" {
			t.Errorf("Accept-Encoding = %q, want %q", encoding, "gzip")
		}
	})
}

func TestAuthHTTPClientNoAPIKey(t *testing.T) {
	var capturedHeaders http.Header

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeaders = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// No API key
	authClient := &authHTTPClient{
		client:    http.DefaultClient,
		apiKey:    "",
		workspace: "",
	}

	req, _ := http.NewRequest("GET", ts.URL, nil)
	_, _ = authClient.Do(req)

	// Headers should be empty when not set
	if capturedHeaders.Get("Authorization") != "" {
		t.Error("Authorization should be empty when apiKey is empty")
	}
	if capturedHeaders.Get("Comet-Workspace") != "" {
		t.Error("Comet-Workspace should be empty when workspace is empty")
	}
}

func TestClientTraceWithTracingDisabled(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	client, _ := NewClient(
		WithURL(ms.URL()),
		WithAPIKey("test-key"),
		WithTracingDisabled(true),
	)

	_, err := client.Trace(context.Background(), "test-trace")
	if err != ErrTracingDisabled {
		t.Errorf("expected ErrTracingDisabled, got %v", err)
	}
}

func TestClientAPI(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	client, _ := NewClient(
		WithURL(ms.URL()),
		WithAPIKey("test-key"),
	)

	api := client.API()
	if api == nil {
		t.Error("API() should not return nil")
	}
}

func TestProjectOption(t *testing.T) {
	opts := &projectOptions{}
	WithProjectDescription("test description")(opts)

	if opts.description != "test description" {
		t.Errorf("description = %q, want %q", opts.description, "test description")
	}
}

func TestVersion(t *testing.T) {
	if Version == "" {
		t.Error("Version should not be empty")
	}
}

// TestClientWithMockServer demonstrates using the MockServer for API tests.
func TestClientWithMockServer(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	// Mock the traces endpoint
	ms.OnPost("/v1/private/traces").RespondJSON(201, map[string]any{})

	// Note: Testing actual API calls requires matching the ogen client's request format
	// which is more complex. This test shows the mock server setup pattern.

	t.Run("mock server records requests", func(t *testing.T) {
		// Make a direct request to verify mock server works
		resp, err := http.Post(ms.URL()+"/v1/private/traces", "application/json", nil)
		if err != nil {
			t.Fatalf("request error: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 201 {
			t.Errorf("StatusCode = %d, want 201", resp.StatusCode)
		}

		if ms.RequestCount() != 1 {
			t.Errorf("RequestCount = %d, want 1", ms.RequestCount())
		}

		last := ms.LastRequest()
		if last.Method != "POST" {
			t.Errorf("Method = %q, want POST", last.Method)
		}
		if last.Path != "/v1/private/traces" {
			t.Errorf("Path = %q, want /v1/private/traces", last.Path)
		}
	})
}

// TestClientWithMatchersExample demonstrates using matchers for assertions.
func TestClientWithMatchersExample(t *testing.T) {
	// Demonstrate the testutil matchers
	t.Run("any matcher", func(t *testing.T) {
		m := testutil.Any()
		testutil.AssertMatch(t, m, "anything")
		testutil.AssertMatch(t, m, 123)
		testutil.AssertMatch(t, m, nil)
	})

	t.Run("any string matcher", func(t *testing.T) {
		m := testutil.AnyString().WithPrefix("trace-")
		if !m.Match("trace-123") {
			t.Error("should match string with prefix")
		}
		if m.Match("span-123") {
			t.Error("should not match string without prefix")
		}
	})

	t.Run("any map with keys", func(t *testing.T) {
		m := testutil.AnyMap("id", "name")
		if !m.Match(map[string]any{"id": "1", "name": "test", "extra": "data"}) {
			t.Error("should match map with required keys")
		}
		if m.Match(map[string]any{"id": "1"}) {
			t.Error("should not match map missing required key")
		}
	})

	t.Run("any float near", func(t *testing.T) {
		m := testutil.AnyFloat().Near(0.5, 0.01)
		if !m.Match(0.505) {
			t.Error("should match float within tolerance")
		}
		if m.Match(0.6) {
			t.Error("should not match float outside tolerance")
		}
	})
}
