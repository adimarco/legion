package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSettings_Validate(t *testing.T) {
	tests := []struct {
		name        string
		settings    Settings
		wantErr     bool
		errContains string
	}{
		{
			name: "valid settings",
			settings: Settings{
				Logger: LoggerSettings{
					Type:            "file",
					Level:           "info",
					ProgressDisplay: true,
					Path:            "test.log",
					BatchSize:       100,
				},
				MCP: MCPSettings{
					Servers: map[string]MCPServerSettings{
						"test": {
							Transport: "stdio",
							Command:   "test",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid logger type",
			settings: Settings{
				Logger: LoggerSettings{
					Type:  "invalid",
					Level: "info",
				},
			},
			wantErr:     true,
			errContains: "invalid logger type",
		},
		{
			name: "invalid log level",
			settings: Settings{
				Logger: LoggerSettings{
					Type:  "file",
					Level: "invalid",
				},
			},
			wantErr:     true,
			errContains: "invalid log level",
		},
		{
			name: "missing file path",
			settings: Settings{
				Logger: LoggerSettings{
					Type:  "file",
					Level: "info",
					Path:  "",
				},
			},
			wantErr:     true,
			errContains: "path is required for file logger",
		},
		{
			name: "invalid batch size",
			settings: Settings{
				Logger: LoggerSettings{
					Type:      "file",
					Level:     "info",
					Path:      "test.log",
					BatchSize: 0,
				},
			},
			wantErr:     true,
			errContains: "batch size must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.settings.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMCPServerSettings_Validate(t *testing.T) {
	tests := []struct {
		name        string
		server      MCPServerSettings
		wantErr     bool
		errContains string
	}{
		{
			name: "valid stdio server",
			server: MCPServerSettings{
				Transport: "stdio",
				Command:   "test",
			},
			wantErr: false,
		},
		{
			name: "valid sse server",
			server: MCPServerSettings{
				Transport: "sse",
				URL:       "http://localhost:8080",
			},
			wantErr: false,
		},
		{
			name: "invalid transport",
			server: MCPServerSettings{
				Transport: "invalid",
			},
			wantErr:     true,
			errContains: "invalid transport",
		},
		{
			name: "missing command for stdio",
			server: MCPServerSettings{
				Transport: "stdio",
			},
			wantErr:     true,
			errContains: "command is required for stdio transport",
		},
		{
			name: "missing url for sse",
			server: MCPServerSettings{
				Transport: "sse",
			},
			wantErr:     true,
			errContains: "url is required for sse transport",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.server.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
