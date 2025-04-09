# Basic Anthropic LLM Example

This example demonstrates basic usage of the Anthropic LLM integration in gofast.

## Prerequisites

1. You need an Anthropic API key. You can get one from [Anthropic's website](https://www.anthropic.com/).
2. Set your API key in an environment variable:
   ```bash
   export ANTHROPIC_API_KEY=your-api-key-here
   ```
   Or create a `.env` file in this directory with:
   ```
   ANTHROPIC_API_KEY=your-api-key-here
   ```

## Running the Example

From this directory:

```bash
go run main.go
```

The example will:
1. Initialize an Anthropic LLM instance
2. Send a simple question to the model
3. Print the response

This demonstrates the basic workflow of using the Anthropic LLM in your own applications. 