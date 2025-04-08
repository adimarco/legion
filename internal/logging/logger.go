package logging

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// zapLogger implements the Logger interface using Uber's Zap
type zapLogger struct {
	logger    *zap.Logger
	namespace string
}

var (
	// Global logger instance
	globalLogger *zap.Logger
	globalMu     sync.RWMutex
)

// Initialize sets up the global logger with the given configuration
func Initialize(cfg Config) error {
	globalMu.Lock()
	defer globalMu.Unlock()

	// Create encoder config
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Configure logger based on type
	var core zapcore.Core
	switch cfg.Type {
	case "console":
		core = zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			zapcore.AddSync(cfg.Writer),
			levelToZapLevel(cfg.Level),
		)
	case "file", "json":
		core = zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(cfg.Writer),
			levelToZapLevel(cfg.Level),
		)
	default:
		return fmt.Errorf("unsupported logger type: %s", cfg.Type)
	}

	// Create logger with namespace field
	globalLogger = zap.New(core)
	return nil
}

// GetLogger returns a new Logger instance with the given namespace
func GetLogger(namespace string) Logger {
	globalMu.RLock()
	defer globalMu.RUnlock()

	if globalLogger == nil {
		// Create a default logger if not initialized
		globalLogger = zap.NewExample()
	}

	return &zapLogger{
		logger:    globalLogger.Named(namespace),
		namespace: namespace,
	}
}

// Debug implements Logger.Debug
func (l *zapLogger) Debug(ctx context.Context, msg string, opts ...EventOption) error {
	event := newEvent(EventTypeDebug, l.namespace, msg)
	for _, opt := range opts {
		opt(event)
	}
	l.logger.Debug(msg, eventToFields(event)...)
	return nil
}

// Info implements Logger.Info
func (l *zapLogger) Info(ctx context.Context, msg string, opts ...EventOption) error {
	event := newEvent(EventTypeInfo, l.namespace, msg)
	for _, opt := range opts {
		opt(event)
	}
	l.logger.Info(msg, eventToFields(event)...)
	return nil
}

// Warning implements Logger.Warning
func (l *zapLogger) Warning(ctx context.Context, msg string, opts ...EventOption) error {
	event := newEvent(EventTypeWarning, l.namespace, msg)
	for _, opt := range opts {
		opt(event)
	}
	l.logger.Warn(msg, eventToFields(event)...)
	return nil
}

// Error implements Logger.Error
func (l *zapLogger) Error(ctx context.Context, msg string, opts ...EventOption) error {
	event := newEvent(EventTypeError, l.namespace, msg)
	for _, opt := range opts {
		opt(event)
	}
	l.logger.Error(msg, eventToFields(event)...)
	return nil
}

// Progress implements Logger.Progress
func (l *zapLogger) Progress(ctx context.Context, msg string, percentage float64, opts ...EventOption) error {
	event := newEvent(EventTypeProgress, l.namespace, msg)
	for _, opt := range opts {
		opt(event)
	}
	// Add percentage to event data
	if event.Data == nil {
		event.Data = make(map[string]interface{})
	}
	event.Data["percentage"] = percentage
	l.logger.Info(msg, eventToFields(event)...)
	return nil
}

// Event implements Logger.Event
func (l *zapLogger) Event(ctx context.Context, etype EventType, msg string, opts ...EventOption) error {
	event := newEvent(etype, l.namespace, msg)
	for _, opt := range opts {
		opt(event)
	}
	l.logger.Info(msg, eventToFields(event)...)
	return nil
}

// Helper functions

func newEvent(etype EventType, namespace, msg string) *Event {
	return &Event{
		Type:      etype,
		Namespace: namespace,
		Message:   msg,
		Timestamp: time.Now(),
	}
}

func eventToFields(event *Event) []zap.Field {
	fields := []zap.Field{
		zap.String("type", string(event.Type)),
		zap.String("namespace", event.Namespace),
		zap.Time("timestamp", event.Timestamp),
	}

	if event.Name != "" {
		fields = append(fields, zap.String("name", event.Name))
	}

	if event.Context != nil {
		if event.Context.SessionID != "" {
			fields = append(fields, zap.String("session_id", event.Context.SessionID))
		}
		if event.Context.WorkflowID != "" {
			fields = append(fields, zap.String("workflow_id", event.Context.WorkflowID))
		}
	}

	if event.Data != nil {
		fields = append(fields, zap.Any("data", event.Data))
	}

	if event.SpanID != "" {
		fields = append(fields, zap.String("span_id", event.SpanID))
	}

	if event.TraceID != "" {
		fields = append(fields, zap.String("trace_id", event.TraceID))
	}

	return fields
}

func levelToZapLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// Config holds logger configuration
type Config struct {
	Type   string
	Level  string
	Writer zapcore.WriteSyncer
}
