# AX Error Codes

The ax package defines 19 domain-specific error codes for the Opik API.

## Error Code Constants

### Not Found Errors

| Constant | Value | Description |
|----------|-------|-------------|
| `ErrTraceNotFound` | `TRACE_NOT_FOUND` | The requested trace was not found |
| `ErrSpanNotFound` | `SPAN_NOT_FOUND` | The requested span was not found |
| `ErrDatasetNotFound` | `DATASET_NOT_FOUND` | The requested dataset was not found |
| `ErrExperimentNotFound` | `EXPERIMENT_NOT_FOUND` | The requested experiment was not found |
| `ErrPromptNotFound` | `PROMPT_NOT_FOUND` | The requested prompt was not found |
| `ErrProjectNotFound` | `PROJECT_NOT_FOUND` | The requested project was not found |
| `ErrFeedbackNotFound` | `FEEDBACK_NOT_FOUND` | The requested feedback was not found |
| `ErrAttachmentNotFound` | `ATTACHMENT_NOT_FOUND` | The requested attachment was not found |
| `ErrWorkspaceNotFound` | `WORKSPACE_NOT_FOUND` | The workspace was not found |
| `ErrEvaluatorNotFound` | `EVALUATOR_NOT_FOUND` | The evaluator was not found |
| `ErrAlertNotFound` | `ALERT_NOT_FOUND` | The alert was not found |
| `ErrQueueNotFound` | `QUEUE_NOT_FOUND` | The annotation queue was not found |
| `ErrDashboardNotFound` | `DASHBOARD_NOT_FOUND` | The dashboard was not found |

### Authentication Errors

| Constant | Value | Description |
|----------|-------|-------------|
| `ErrUnauthorized` | `UNAUTHORIZED` | Authentication is required or failed |
| `ErrForbidden` | `FORBIDDEN` | User lacks permission for the operation |

### Validation Errors

| Constant | Value | Description |
|----------|-------|-------------|
| `ErrInvalidInput` | `INVALID_INPUT` | Request input validation failed |
| `ErrConflict` | `CONFLICT` | Resource conflict (e.g., duplicate name) |

### Server Errors

| Constant | Value | Description |
|----------|-------|-------------|
| `ErrRateLimited` | `RATE_LIMITED` | Request was rate limited (retryable) |
| `ErrInternalError` | `INTERNAL_ERROR` | Internal server error (retryable) |

## Error Metadata

Each error code has associated metadata:

```go
info := ax.GetErrorInfo(ax.ErrTraceNotFound)
// info.Code        = "TRACE_NOT_FOUND"
// info.Description = "The requested trace was not found"
// info.Category    = "not_found"
// info.Retryable   = false
// info.HTTPStatus  = 404
```

## Category Helpers

```go
// Check error category
ax.IsAuthError(code)       // auth: UNAUTHORIZED, FORBIDDEN
ax.IsNotFoundError(code)   // not_found: all *_NOT_FOUND codes
ax.IsValidationError(code) // validation: INVALID_INPUT
ax.IsConflictError(code)   // conflict: CONFLICT
ax.IsRetryableError(code)  // rate_limit, server: retryable errors
```

## Using Error Codes

### Check Specific Error

```go
if opik.IsAXError(err, ax.ErrTraceNotFound) {
    // Handle trace not found
}
```

### Extract Error Code

```go
code, ok := opik.GetAXErrorCode(err)
if ok {
    switch code {
    case ax.ErrTraceNotFound:
        // ...
    case ax.ErrProjectNotFound:
        // ...
    }
}
```

### Get Error Info

```go
info := opik.GetAXErrorInfo(err)
if info != nil {
    fmt.Printf("Category: %s, Retryable: %v\n",
        info.Category, info.Retryable)
}
```

## HTTP Status Mapping

When no specific error code is available, you can infer from HTTP status:

```go
code := ax.ErrorCodeForHTTPStatus(statusCode)
// 400 -> INVALID_INPUT
// 401 -> UNAUTHORIZED
// 403 -> FORBIDDEN
// 409 -> CONFLICT
// 429 -> RATE_LIMITED
// 500 -> INTERNAL_ERROR
```

## Complete Example

```go
func handleError(err error) {
    code, ok := opik.GetAXErrorCode(err)
    if !ok {
        log.Printf("Unknown error: %v", err)
        return
    }

    info := ax.GetErrorInfo(code)

    switch info.Category {
    case "not_found":
        log.Printf("Resource not found: %s", info.Description)
        // Consider creating the resource

    case "auth":
        log.Printf("Auth error: %s", info.Description)
        // Re-authenticate or escalate

    case "validation":
        log.Printf("Invalid request: %s", info.Description)
        // Fix the request

    case "conflict":
        log.Printf("Resource conflict: %s", info.Description)
        // Fetch existing resource

    case "rate_limit", "server":
        if info.Retryable {
            log.Printf("Retryable error: %s", info.Description)
            // Back off and retry
        }
    }
}
```
