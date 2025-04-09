# Hive

Hive is an intelligent agent framework that enables the creation, composition, and deployment of AI agents. Think of it as "npm for AI agents" meets "Slack App Directory" - a decentralized registry and platform where developers can build, share, and deploy intelligent agents.

## Features

- 🐝 **Agent Registry** - Publish and version agent configurations
- 🔄 **Team Composition** - Compose agents into specialized teams
- 🌐 **Multi-Channel** - Deploy agents across Slack, Discord, Terminal, and more
- 🛠️ **Tool Integration** - Rich ecosystem of tools and capabilities
- 🔒 **Enterprise Ready** - Built for production with security and scaling in mind

## Quick Start

1. Install Hive:
```bash
go install github.com/adimarco/hive/cmd/hive@latest
```

2. Set up your Anthropic API key:
```bash
export ANTHROPIC_API_KEY=your-api-key-here
```

3. Create a new project:
```bash
hive setup myproject
cd myproject
```

4. Run examples:
```bash
hive bootstrap workflow
```

## Documentation

- [Getting Started](docs/getting-started.md)
- [Agent Development](docs/agent-development.md)
- [Tool Integration](docs/tool-integration.md)
- [Channel Support](docs/channels.md)

## Development

1. Clone the repository:
```bash
git clone https://github.com/adimarco/hive.git
cd hive
```

2. Build the CLI:
```bash
go build -o hive cmd/hive/main.go
```

3. Run tests:
```bash
go test ./...
```

## Examples

- `examples/simple_agent` - Basic agent creation and interaction
- `examples/anthropic_basic` - Direct LLM integration example
- More examples coming soon!

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

## License

MIT License - see [LICENSE](LICENSE) for details. 