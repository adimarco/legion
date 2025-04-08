package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

type testBuffer struct {
	bytes.Buffer
}

func (b *testBuffer) Sync() error {
	return nil
}

func TestLogger_Initialization(t *testing.T) {
	// Test console logger initialization
	buf := &testBuffer{}
	err := Initialize(Config{
		Type:   "console",
		Level:  "info",
		Writer: buf,
	})
	require.NoError(t, err)

	// Test file logger initialization
	err = Initialize(Config{
		Type:   "file",
		Level:  "debug",
		Writer: buf,
	})
	require.NoError(t, err)

	// Test invalid type
	err = Initialize(Config{
		Type:   "invalid",
		Level:  "info",
		Writer: buf,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported logger type")
}

func TestLogger_Levels(t *testing.T) {
	buf := &testBuffer{}
	err := Initialize(Config{
		Type:   "json",
		Level:  "debug",
		Writer: buf,
	})
	require.NoError(t, err)

	logger := GetLogger("test")
	ctx := context.Background()

	tests := []struct {
		name     string
		logFunc  func(string, ...EventOption) error
		level    string
		message  string
		wantType EventType
	}{
		{
			name: "debug message",
			logFunc: func(msg string, opts ...EventOption) error {
				return logger.Debug(ctx, msg, opts...)
			},
			level:    "debug",
			message:  "debug test",
			wantType: EventTypeDebug,
		},
		{
			name: "info message",
			logFunc: func(msg string, opts ...EventOption) error {
				return logger.Info(ctx, msg, opts...)
			},
			level:    "info",
			message:  "info test",
			wantType: EventTypeInfo,
		},
		{
			name: "warning message",
			logFunc: func(msg string, opts ...EventOption) error {
				return logger.Warning(ctx, msg, opts...)
			},
			level:    "warn",
			message:  "warning test",
			wantType: EventTypeWarning,
		},
		{
			name: "error message",
			logFunc: func(msg string, opts ...EventOption) error {
				return logger.Error(ctx, msg, opts...)
			},
			level:    "error",
			message:  "error test",
			wantType: EventTypeError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			err := tt.logFunc(tt.message)
			require.NoError(t, err)

			var logEntry map[string]interface{}
			err = json.Unmarshal(buf.Bytes(), &logEntry)
			require.NoError(t, err)

			assert.Equal(t, tt.level, logEntry["level"])
			assert.Equal(t, tt.message, logEntry["msg"])
			assert.Equal(t, string(tt.wantType), logEntry["type"])
			assert.Equal(t, "test", logEntry["namespace"])
		})
	}
}

func TestLogger_EventOptions(t *testing.T) {
	buf := &testBuffer{}
	err := Initialize(Config{
		Type:   "json",
		Level:  "info",
		Writer: buf,
	})
	require.NoError(t, err)

	logger := GetLogger("test")
	ctx := context.Background()

	// Test with all event options
	eventCtx := EventContext{
		SessionID:  "test-session",
		WorkflowID: "test-workflow",
	}
	data := map[string]interface{}{
		"key": "value",
	}

	err = logger.Info(ctx, "test message",
		WithName("test-event"),
		WithContext(eventCtx),
		WithData(data),
	)
	require.NoError(t, err)

	var logEntry map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	assert.Equal(t, "test-event", logEntry["name"])
	assert.Equal(t, "test-session", logEntry["session_id"])
	assert.Equal(t, "test-workflow", logEntry["workflow_id"])
	assert.Equal(t, "value", logEntry["data"].(map[string]interface{})["key"])
}

func TestLogger_Progress(t *testing.T) {
	buf := &testBuffer{}
	err := Initialize(Config{
		Type:   "json",
		Level:  "info",
		Writer: buf,
	})
	require.NoError(t, err)

	logger := GetLogger("test")
	ctx := context.Background()

	err = logger.Progress(ctx, "progress test", 50.5)
	require.NoError(t, err)

	var logEntry map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	assert.Equal(t, "progress test", logEntry["msg"])
	assert.Equal(t, string(EventTypeProgress), logEntry["type"])
	assert.Equal(t, 50.5, logEntry["data"].(map[string]interface{})["percentage"])
}

func TestLogger_LevelConversion(t *testing.T) {
	tests := []struct {
		level    string
		expected zapcore.Level
	}{
		{"debug", zapcore.DebugLevel},
		{"info", zapcore.InfoLevel},
		{"warning", zapcore.WarnLevel},
		{"error", zapcore.ErrorLevel},
		{"invalid", zapcore.InfoLevel}, // Default to info
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			actual := levelToZapLevel(tt.level)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestLogger_Namespacing(t *testing.T) {
	buf := &testBuffer{}
	err := Initialize(Config{
		Type:   "json",
		Level:  "info",
		Writer: buf,
	})
	require.NoError(t, err)

	// Create loggers with different namespaces
	logger1 := GetLogger("namespace1")
	logger2 := GetLogger("namespace2")
	ctx := context.Background()

	// Log messages from both loggers
	err = logger1.Info(ctx, "test message 1")
	require.NoError(t, err)

	var logEntry1 map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &logEntry1)
	require.NoError(t, err)
	assert.Equal(t, "namespace1", logEntry1["namespace"])

	buf.Reset()

	err = logger2.Info(ctx, "test message 2")
	require.NoError(t, err)

	var logEntry2 map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &logEntry2)
	require.NoError(t, err)
	assert.Equal(t, "namespace2", logEntry2["namespace"])
}
