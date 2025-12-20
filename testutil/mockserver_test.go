package testutil

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestMockServer(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	t.Run("URL is set", func(t *testing.T) {
		if ms.URL() == "" {
			t.Error("URL should not be empty")
		}
	})

	t.Run("default 404 for unregistered routes", func(t *testing.T) {
		resp, err := http.Get(ms.URL() + "/unknown")
		if err != nil {
			t.Fatalf("request error: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusNotFound)
		}
	})
}

func TestMockServerRoutes(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	t.Run("GET route", func(t *testing.T) {
		ms.OnGet("/api/test").Respond(200, "OK")

		resp, err := http.Get(ms.URL() + "/api/test")
		if err != nil {
			t.Fatalf("request error: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
		}
	})

	t.Run("POST route with JSON", func(t *testing.T) {
		ms.OnPost("/api/create").RespondJSON(201, map[string]string{"id": "123"})

		body := bytes.NewBufferString(`{"name":"test"}`)
		resp, err := http.Post(ms.URL()+"/api/create", "application/json", body)
		if err != nil {
			t.Fatalf("request error: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 201 {
			t.Errorf("StatusCode = %d, want 201", resp.StatusCode)
		}
		if resp.Header.Get("Content-Type") != "application/json" {
			t.Error("Content-Type should be application/json")
		}

		var result map[string]string
		_ = json.NewDecoder(resp.Body).Decode(&result)
		if result["id"] != "123" {
			t.Errorf("response id = %q, want %q", result["id"], "123")
		}
	})

	t.Run("custom handler", func(t *testing.T) {
		callCount := 0
		ms.OnGet("/api/custom").WithHandler(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte("custom"))
		})

		resp, err := http.Get(ms.URL() + "/api/custom")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusAccepted {
			t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusAccepted)
		}
		if callCount != 1 {
			t.Errorf("handler called %d times, want 1", callCount)
		}
	})

	t.Run("with headers", func(t *testing.T) {
		ms.OnGet("/api/headers").Respond(200, "OK").WithHeaders(map[string]string{
			"X-Custom": "value",
		})

		resp, err := http.Get(ms.URL() + "/api/headers")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.Header.Get("X-Custom") != "value" {
			t.Error("custom header not set")
		}
	})
}

func TestMockServerRecording(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	ms.OnPost("/api/record").Respond(200, "OK")

	// Make some requests
	body := bytes.NewBufferString(`{"data":"test"}`)
	req, _ := http.NewRequest("POST", ms.URL()+"/api/record", body)
	req.Header.Set("Authorization", "Bearer token123")
	_, _ = http.DefaultClient.Do(req)

	t.Run("records request count", func(t *testing.T) {
		if ms.RequestCount() != 1 {
			t.Errorf("RequestCount = %d, want 1", ms.RequestCount())
		}
	})

	t.Run("records request details", func(t *testing.T) {
		last := ms.LastRequest()
		if last == nil {
			t.Fatal("LastRequest returned nil")
		}
		if last.Method != "POST" {
			t.Errorf("Method = %q, want POST", last.Method)
		}
		if last.Path != "/api/record" {
			t.Errorf("Path = %q, want /api/record", last.Path)
		}
		if last.Headers.Get("Authorization") != "Bearer token123" {
			t.Error("Authorization header not recorded")
		}
	})

	t.Run("requests for path", func(t *testing.T) {
		reqs := ms.RequestsForPath("/api/record")
		if len(reqs) != 1 {
			t.Errorf("RequestsForPath length = %d, want 1", len(reqs))
		}
	})

	t.Run("route call count", func(t *testing.T) {
		count := ms.RouteCallCount("POST", "/api/record")
		if count != 1 {
			t.Errorf("RouteCallCount = %d, want 1", count)
		}
	})
}

func TestMockServerReset(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	ms.OnGet("/test").Respond(200, "OK")
	_, _ = http.Get(ms.URL() + "/test")
	_, _ = http.Get(ms.URL() + "/test")

	if ms.RequestCount() != 2 {
		t.Errorf("RequestCount before reset = %d, want 2", ms.RequestCount())
	}

	ms.Reset()

	if ms.RequestCount() != 0 {
		t.Errorf("RequestCount after reset = %d, want 0", ms.RequestCount())
	}
	if ms.RouteCallCount("GET", "/test") != 0 {
		t.Error("RouteCallCount should be 0 after reset")
	}
}

func TestMockServerAllRequests(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	ms.OnGet("/a").Respond(200, "")
	ms.OnGet("/b").Respond(200, "")

	_, _ = http.Get(ms.URL() + "/a")
	_, _ = http.Get(ms.URL() + "/b")
	_, _ = http.Get(ms.URL() + "/a")

	reqs := ms.Requests()
	if len(reqs) != 3 {
		t.Errorf("Requests length = %d, want 3", len(reqs))
	}
}

func TestMockServerBodyRecording(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	ms.OnPost("/api/data").Respond(200, "OK")

	payload := map[string]any{"key": "value", "num": 42}
	jsonData, _ := json.Marshal(payload)
	resp, err := http.Post(ms.URL()+"/api/data", "application/json", bytes.NewReader(jsonData))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	_, _ = io.ReadAll(resp.Body)

	last := ms.LastRequest()
	if last == nil {
		t.Fatal("no request recorded")
	}

	var recorded map[string]any
	_ = json.Unmarshal(last.Body, &recorded)
	if recorded["key"] != "value" {
		t.Errorf("recorded body key = %v, want %q", recorded["key"], "value")
	}
}

func TestMockServerHTTPMethods(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	ms.OnGet("/get").Respond(200, "get")
	ms.OnPost("/post").Respond(201, "post")
	ms.OnPut("/put").Respond(200, "put")
	ms.OnPatch("/patch").Respond(200, "patch")
	ms.OnDelete("/delete").Respond(204, "")

	tests := []struct {
		method string
		path   string
		status int
	}{
		{"GET", "/get", 200},
		{"POST", "/post", 201},
		{"PUT", "/put", 200},
		{"PATCH", "/patch", 200},
		{"DELETE", "/delete", 204},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, ms.URL()+tt.path, nil)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("request error: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.status {
				t.Errorf("StatusCode = %d, want %d", resp.StatusCode, tt.status)
			}
		})
	}
}
