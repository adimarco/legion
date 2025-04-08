# FastAgent (Go Implementation)

FastAgent is a Go-based framework for building effective agents using the Model Context Protocol (MCP). This project is currently in early development, implementing a Go version of the FastAgent framework.

## Current Status

This is an early implementation with basic configuration and CLI functionality. More features are actively being developed.

### Implemented Features

- **CLI Framework**
  - Basic command structure using Cobra
  - `setup` command for new project initialization
  - `bootstrap` command for creating example applications
  - `config` command for managing configuration

- **Configuration System**
  - YAML-based configuration with environment variable overrides
  - Validation for settings and server configurations
  - Support for logging and MCP server settings
  - Comprehensive test coverage

## Getting Started

### Prerequisites

- Go 1.21 or later
- Git

### Installation

```bash
# Clone the repository
git clone <your-repo-url>
cd gofast

# Install dependencies
go mod tidy
```

### Basic Usage

```bash
# Build the CLI
go build -o gofast cmd/gofast/main.go

# View available commands
./gofast --help

# Set up a new project
./gofast setup myproject

# View configuration
./gofast config show
```

### Configuration

FastAgent uses a YAML-based configuration system with two main files:

- `fastagent.config.yaml`: Main configuration file
- `fastagent.secrets.yaml`: Secrets and API keys (never commit this file)

Configuration can be overridden using environment variables with the `FASTAGENT_` prefix.

Example configuration:
```yaml
default_model: haiku

logger:
  type: console
  level: info
  progress_display: true

mcp:
  servers:
    fetch:
      transport: stdio
      command: uvx
      args: ["mcp-server-fetch"]
```

## Development Status

This project is under active development. See [MIGRATION_PLAN.md](MIGRATION_PLAN.md) for the detailed development roadmap.

### Next Steps

- Context management implementation
- Logging system
- MCP server support
- Workflow execution

## Contributing

This project is in early development and not yet ready for contributions.

## License

[License details to be determined] 