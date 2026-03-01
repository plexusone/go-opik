# Release Notes v0.4.0

## Overview

This release includes critical bug fixes for span/trace project association and JSON encoding, plus a new `ListSpans()` convenience method.

## Bug Fixes

### Spans Now Correctly Associated with Project

**Critical Fix**: Spans were being created in "Default Project" instead of the configured project because `ProjectName` was not being set in the span creation request.

- Added `ProjectName` field to `SpanWrite` in span creation
- Spans now properly inherit the project from the client configuration
- Fixes issue where spans would not appear when querying by project name

### Fixed Malformed JSON in Batch Update Requests

**Critical Fix**: Batch update requests (PATCH) for traces and spans were returning HTTP 400 due to malformed JSON.

- **Root cause**: The ogen-generated `JsonListString.Encode()` method writes nothing when the byte slice is empty, but the field name is still written, creating invalid JSON like `{"input":"output":{"result":"test"}}`
- **Fix**: All `JsonListString` fields (Input, Output, Metadata) now use `[]byte("null")` instead of empty byte slices when no value is set
- Affects: `Trace.End()`, `Trace.Update()`, `Span.End()`, `Span.Update()`

## New Features

### ListSpans() Convenience Method

Added high-level SDK method for listing spans by trace ID:

```go
// SpanInfo represents basic span information
type SpanInfo struct {
    ID           string
    TraceID      string
    ParentSpanID string
    Name         string
    Type         string
    StartTime    time.Time
    EndTime      time.Time
    Model        string
    Provider     string
}

// List spans for a specific trace
spans, err := client.ListSpans(ctx, traceID, page, size)
```

## Debug Tools

Added `examples/debug-test/` directory with diagnostic tools:

- `main.go` - Lists traces and spans
- `get_trace.go` - Gets detailed trace info via API
- `test_full_flow.go` - Full create/end test with HTTP debug output
- `test_span_creation.go` - Tests span creation via SDK
- `test_json_encode.go` - Tests JSON encoding of API types
- `check_spans.go` - Comprehensive trace+span inspector using SDK methods

## Files Changed

- `span.go` - Added ProjectName to span creation, fixed JsonListString encoding
- `trace.go` - Fixed JsonListString encoding in End() and Update() methods
- `client.go` - Added SpanInfo struct and ListSpans() method
- `examples/debug-test/` - New diagnostic tools

## Upgrade Notes

This is a bug-fix release with no breaking API changes. Upgrade recommended for all users experiencing:

- Spans not appearing in their project
- HTTP 400 errors when ending traces or spans
- Need for listing spans by trace ID

## Testing

All changes have been verified against Comet Opik Cloud:

- Traces and spans now appear correctly in the configured project
- Batch update requests return HTTP 204 (success)
- ListSpans() correctly returns spans for a given trace
