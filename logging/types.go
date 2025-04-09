package logging

import (
	"context"
	"time"
)

// EventType represents the severity or role of an event
type EventType string

const (
	EventTypeDebug    EventType = "debug"
	EventTypeInfo     EventType = "info"
	EventTypeWarning  EventType = "warning"
	EventTypeError    EventType = "error"
	EventTypeProgress EventType = "progress"
)

// EventContext stores correlation or cross-cutting data
type EventContext struct {
	SessionID  string `json:"session_id,omitempty"`
	WorkflowID string `json:"workflow_id,omitempty"`
	// Additional fields can be added as needed
}

// Event represents a log event with metadata and payload
type Event struct {
	Type      EventType              `json:"type"`
	Name      string                 `json:"name,omitempty"`
	Namespace string                 `json:"namespace"`
	Message   string                 `json:"message"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Context   *EventContext          `json:"context,omitempty"`
	SpanID    string                 `json:"span_id,omitempty"`
	TraceID   string                 `json:"trace_id,omitempty"`
}

// EventFilter defines criteria for filtering events
type EventFilter struct {
	Types      map[EventType]bool // Allowed event types
	Names      map[string]bool    // Allowed event names
	Namespaces map[string]bool    // Allowed namespace prefixes
	MinLevel   EventType          // Minimum severity level
}

// EventListener processes events asynchronously
type EventListener interface {
	// HandleEvent processes an incoming event
	HandleEvent(ctx context.Context, event Event) error
}

// EventTransport sends events to external systems
type EventTransport interface {
	// SendEvent sends an event to the external system
	SendEvent(ctx context.Context, event Event) error
}

// Logger provides a developer-friendly logging interface
type Logger interface {
	// Debug logs a debug message
	Debug(ctx context.Context, msg string, opts ...EventOption) error

	// Info logs an info message
	Info(ctx context.Context, msg string, opts ...EventOption) error

	// Warning logs a warning message
	Warning(ctx context.Context, msg string, opts ...EventOption) error

	// Error logs an error message
	Error(ctx context.Context, msg string, opts ...EventOption) error

	// Progress logs a progress message with optional percentage
	Progress(ctx context.Context, msg string, percentage float64, opts ...EventOption) error

	// Event emits a custom event
	Event(ctx context.Context, etype EventType, msg string, opts ...EventOption) error
}

// EventOption allows for optional event parameters
type EventOption func(*Event)

// WithName sets the event name
func WithName(name string) EventOption {
	return func(e *Event) {
		e.Name = name
	}
}

// WithContext sets the event context
func WithContext(ctx EventContext) EventOption {
	return func(e *Event) {
		e.Context = &ctx
	}
}

// WithData sets additional event data
func WithData(data map[string]interface{}) EventOption {
	return func(e *Event) {
		e.Data = data
	}
}
