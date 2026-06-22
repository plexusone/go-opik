// Package agentobs provides generic agent observability interfaces for use with Opik.
//
// This package defines generic event types and interfaces that allow agent frameworks
// to send observability data to Opik without depending on specific framework implementations.
// It is designed to be imported by agent framework integrations while keeping opik-go
// free of framework-specific dependencies.
//
// # Core Types
//
// [AgentEvent] represents a generic agent lifecycle event with type, timestamp,
// session/trace/span identifiers, and arbitrary data. Framework integrations convert
// their native events into AgentEvent before sending to Opik.
//
// [TraceManager] manages the lifecycle of Opik traces associated with agent sessions,
// including automatic cleanup of stale traces and correlation of spans within traces.
//
// # Utilities
//
// The package includes utilities for common observability tasks:
//
//   - [ExtractMedia] extracts media references from message content
//   - [Sanitize] removes internal markers and sensitive content
//   - [SanitizeMap] recursively sanitizes map data structures
//
// # Usage
//
// Framework integrations should:
//
//  1. Convert framework-specific events to [AgentEvent]
//  2. Use [TraceManager] to manage trace lifecycle
//  3. Use [Sanitize] to clean content before sending to Opik
//  4. Use [ExtractMedia] to extract attachments from message content
//
// Example:
//
//	manager, _ := agentobs.NewTraceManager(client, agentobs.TraceManagerConfig{
//	    StaleTimeout: 5 * time.Minute,
//	})
//	defer manager.Close()
//
//	// On session creation
//	trace, _ := manager.StartTrace(ctx, "session-123", "my-agent")
//
//	// On message
//	event := agentobs.AgentEvent{
//	    Type:      agentobs.EventMessageReceived,
//	    SessionID: "session-123",
//	    Data: map[string]any{
//	        "content": agentobs.Sanitize(content, agentobs.DefaultSanitizeConfig()),
//	    },
//	}
package agentobs
