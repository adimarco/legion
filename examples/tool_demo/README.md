# Tool Integration Demo

This example demonstrates how to register and use tools with the Hive framework. Tools allow an agent to interact with the outside world to gather information or perform actions.

## What This Example Shows

1. **Tool Registration**: How to register different types of tools with various function signatures
2. **Tool Categories**: Organizing tools by category (file, system, date/time)
3. **Tool Usage**: How the agent identifies when to use tools and invokes them
4. **Tool Loop**: How the agent can use tools, get results, and use those results to inform its response

## Types of Tools Demonstrated

### File Tools
- `readFile` - Reads the contents of a file
- `listDir` - Lists the contents of a directory

### System Tools
- `getWorkingDir` - Gets the current working directory
- `getEnv` - Gets the value of an environment variable

### Date/Time Tools
- `getDateTime` - Gets the current date and time
- `formatDate` - Formats a date string

## Running the Example

```bash
go run examples/tool_demo/main.go
```

## Sample Conversation

Here's a sample of what you can ask the agent:

- "What files are in the current directory?"
- "What's my current working directory?"
- "What's the value of the HOME environment variable?"
- "What's the current date and time?"

The agent will use the appropriate tools to answer your questions.

## How Tool Integration Works

1. The agent receives user input
2. The agent decides if a tool is needed to answer the question
3. The agent calls the appropriate tool with the required arguments
4. The tool executes and returns a result
5. The agent incorporates the tool result into its response
6. This loop continues until the agent has gathered all needed information

This pattern is similar to function calling in other AI frameworks but implemented in a way that's idiomatic to Go. 