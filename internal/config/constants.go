package config

// Environment variable prefixes
const (
	EnvPrefix = "FASTAGENT_"
)

// Valid logger types
var validLoggerTypes = map[string]bool{
	"none":    true,
	"console": true,
	"file":    true,
}

// Valid log levels
var validLogLevels = map[string]bool{
	"debug":   true,
	"info":    true,
	"warning": true,
	"error":   true,
}

// Valid transport types
var validTransportTypes = map[string]bool{
	"stdio": true,
	"sse":   true,
}
