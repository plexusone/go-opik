package testutil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
)

// MockServer provides an HTTP test server that records requests and returns configured responses.
// This is similar to Python's respx library.
type MockServer struct {
	Server   *httptest.Server
	mu       sync.Mutex
	requests []*RecordedRequest
	routes   map[string]*Route
}

// RecordedRequest captures details of an incoming request.
type RecordedRequest struct {
	Method  string
	Path    string
	Headers http.Header
	Body    []byte
}

// Route defines a mock route with its response.
type Route struct {
	Method     string
	Path       string
	StatusCode int
	Response   any
	Headers    map[string]string
	Handler    http.HandlerFunc
	CallCount  int
}

// NewMockServer creates a new mock server.
func NewMockServer() *MockServer {
	ms := &MockServer{
		requests: make([]*RecordedRequest, 0),
		routes:   make(map[string]*Route),
	}

	ms.Server = httptest.NewServer(http.HandlerFunc(ms.handler))
	return ms
}

// URL returns the mock server URL.
func (ms *MockServer) URL() string {
	return ms.Server.URL
}

// Close shuts down the mock server.
func (ms *MockServer) Close() {
	ms.Server.Close()
}

// OnGet registers a GET route.
func (ms *MockServer) OnGet(path string) *Route {
	return ms.On("GET", path)
}

// OnPost registers a POST route.
func (ms *MockServer) OnPost(path string) *Route {
	return ms.On("POST", path)
}

// OnPut registers a PUT route.
func (ms *MockServer) OnPut(path string) *Route {
	return ms.On("PUT", path)
}

// OnPatch registers a PATCH route.
func (ms *MockServer) OnPatch(path string) *Route {
	return ms.On("PATCH", path)
}

// OnDelete registers a DELETE route.
func (ms *MockServer) OnDelete(path string) *Route {
	return ms.On("DELETE", path)
}

// On registers a route for any method.
func (ms *MockServer) On(method, path string) *Route {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	key := method + " " + path
	route := &Route{
		Method:     method,
		Path:       path,
		StatusCode: http.StatusOK,
	}
	ms.routes[key] = route
	return route
}

// Respond sets the response for a route.
func (r *Route) Respond(statusCode int, body any) *Route {
	r.StatusCode = statusCode
	r.Response = body
	return r
}

// RespondJSON sets a JSON response.
func (r *Route) RespondJSON(statusCode int, body any) *Route {
	r.StatusCode = statusCode
	r.Response = body
	if r.Headers == nil {
		r.Headers = make(map[string]string)
	}
	r.Headers["Content-Type"] = "application/json"
	return r
}

// WithHeaders adds response headers.
func (r *Route) WithHeaders(headers map[string]string) *Route {
	r.Headers = headers
	return r
}

// WithHandler sets a custom handler.
func (r *Route) WithHandler(handler http.HandlerFunc) *Route {
	r.Handler = handler
	return r
}

func (ms *MockServer) handler(w http.ResponseWriter, r *http.Request) {
	// Record the request
	body := make([]byte, 0)
	if r.Body != nil {
		body, _ = readBody(r)
	}

	ms.mu.Lock()
	ms.requests = append(ms.requests, &RecordedRequest{
		Method:  r.Method,
		Path:    r.URL.Path,
		Headers: r.Header,
		Body:    body,
	})

	// Find matching route
	key := r.Method + " " + r.URL.Path
	route, ok := ms.routes[key]
	if ok {
		route.CallCount++
	}
	ms.mu.Unlock()

	if !ok {
		http.NotFound(w, r)
		return
	}

	// Use custom handler if set
	if route.Handler != nil {
		route.Handler(w, r)
		return
	}

	// Set headers
	for k, v := range route.Headers {
		w.Header().Set(k, v)
	}

	// Write response
	w.WriteHeader(route.StatusCode)
	if route.Response != nil {
		if route.Headers["Content-Type"] == "application/json" {
			_ = json.NewEncoder(w).Encode(route.Response)
		} else if s, ok := route.Response.(string); ok {
			_, _ = w.Write([]byte(s))
		} else if b, ok := route.Response.([]byte); ok {
			_, _ = w.Write(b)
		}
	}
}

func readBody(r *http.Request) ([]byte, error) {
	if r.Body == nil {
		return nil, nil
	}
	defer r.Body.Close()

	buf := make([]byte, 0, 1024)
	for {
		chunk := make([]byte, 1024)
		n, err := r.Body.Read(chunk)
		buf = append(buf, chunk[:n]...)
		if err != nil {
			break
		}
	}
	return buf, nil
}

// Requests returns all recorded requests.
func (ms *MockServer) Requests() []*RecordedRequest {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	return append([]*RecordedRequest{}, ms.requests...)
}

// RequestCount returns the number of recorded requests.
func (ms *MockServer) RequestCount() int {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	return len(ms.requests)
}

// LastRequest returns the most recent request.
func (ms *MockServer) LastRequest() *RecordedRequest {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	if len(ms.requests) == 0 {
		return nil
	}
	return ms.requests[len(ms.requests)-1]
}

// RequestsForPath returns requests matching a specific path.
func (ms *MockServer) RequestsForPath(path string) []*RecordedRequest {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	result := make([]*RecordedRequest, 0)
	for _, req := range ms.requests {
		if req.Path == path {
			result = append(result, req)
		}
	}
	return result
}

// Reset clears all recorded requests.
func (ms *MockServer) Reset() {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.requests = make([]*RecordedRequest, 0)
	for _, route := range ms.routes {
		route.CallCount = 0
	}
}

// RouteCallCount returns how many times a route was called.
func (ms *MockServer) RouteCallCount(method, path string) int {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	key := method + " " + path
	if route, ok := ms.routes[key]; ok {
		return route.CallCount
	}
	return 0
}
