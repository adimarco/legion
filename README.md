# FastAgent (Go Implementation)

FastAgent is a Go-based framework for building effective agents using the Model Context Protocol (MCP). It provides a robust foundation for creating AI agents that can use tools, manage context, and execute complex workflows.

## Core Features

- **Provider-Agnostic LLM Interface**
  - Support for multiple LLM providers (Anthropic, OpenAI)
  - Clean abstraction for message handling
  - Unified tool calling interface
  - Structured output support

- **Memory & Context Management**
  - Thread-safe context handling
  - Flexible memory management
  - Separate prompt/conversation storage
  - History serialization

- **Development Tools**
  - Passthrough LLM for testing
  - Playback LLM for simulations
  - Comprehensive testing utilities
  - Debug-friendly logging

- **CLI Framework**
  - Project setup and bootstrapping
  - Configuration management
  - Example application generation
  - Development utilities

## Current Status

This is an active implementation with the following components:

### Implemented Features
- **Core Framework**
  - CLI structure using Cobra
  - YAML configuration with validation
  - Environment variable support
  - Thread-safe context management
  - Structured logging with Zap

### In Development
- Message handling system
- Memory management
- Provider integration
- Agent runtime

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

## Development Roadmap

See [MIGRATION_PLAN.md](MIGRATION_PLAN.md) for the detailed development roadmap.

Current focus:
1. Message handling system
2. Memory management
3. Provider integration
4. Agent runtime

## Architecture

FastAgent follows these key principles:

1. **Type Safety**
   - Provider-specific type handling
   - Interface-based abstractions
   - Generic message types

2. **Clean Abstractions**
   - Provider-agnostic core
   - Pluggable components
   - Clear boundaries

3. **Testing First**
   - Comprehensive test coverage
   - Mock providers
   - Development tools

4. **Production Ready**
   - Thread safety
   - Error handling
   - Resource management

## Contributing

This project is in active development. See [MIGRATION_PLAN.md](MIGRATION_PLAN.md) for the current status and roadmap.

## License

[License details to be determined] 