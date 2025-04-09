# Concurrent Specialists Demo

This demo showcases the power of FastAgent's channel-based concurrent agent system by implementing multiple specialized AI agents that process queries simultaneously.

## Overview

The demo creates three specialist agents (Science, History, and Technology) that run concurrently, each processing their own stream of domain-specific questions. It demonstrates:

- Non-blocking message handling
- Concurrent processing
- Pretty terminal output
- Clean coordination and shutdown

## System Architecture

```
                                ┌──────────────────┐
                                │                  │
                                │   Anthropic LLM  │
                                │                  │
                                └─────────┬────────┘
                                          │
                                          │
              ┌──────────────────────┬────┴────┬───────────────────────┐
              │                      │         │                       │
     ┌────────▼───────┐     ┌────────▼───────┐ │              ┌────────▼───────┐
     │  Science Agent │     │ History Agent  │ │              │   Tech Agent   │
     │                │     │                │ │              │                │
     │  Input Channel │     │  Input Channel │ │              │  Input Channel │
     │ Output Channel │     │ Output Channel │ │              │ Output Channel │
     │  Error Channel │     │  Error Channel │ │              │  Error Channel │
     └────────┬───────┘     └────────┬───────┘ │              └────────┬───────┘
              │                       │        │                       │
              │                       │        │                       │
     ┌────────▼───────┐     ┌────────▼───────┐ │              ┌────────▼───────┐
     │   Goroutine    │     │   Goroutine    │ │              │   Goroutine    │
     │ Response Proc. │     │ Response Proc. │ │              │ Response Proc. │
     └────────┬───────┘     └────────┬───────┘ │              └────────┬───────┘
              │                      │         │                       │
              │                      │         │                       │
              └──────────────────────┴─────────┴───────────────────────┘
                                     │
                               ┌─────▼─────┐
                               │ Terminal  │
                               │  Output   │
                               └───────────┘
```

## Architecture Details

### Agent Configuration

Each specialist is configured with a specific focus:

```go
func specialistConfig(name, specialty string) agent.AgentConfig {
    return agent.AgentConfig{
        Name: name,
        Type: agent.AgentTypeBasic,
        Instruction: fmt.Sprintf(`You are an AI specialist focused on %s.
Keep responses brief and to the point.
Always start with a relevant emoji.
Format important terms in *bold*.`, specialty),
    }
}
```

The instruction template creates agents that:
- Focus on their specialty domain
- Provide concise responses
- Use emojis for visual appeal
- Format key terms in bold

### Concurrent Processing

The demo uses Go's concurrency primitives to handle multiple agents and messages:

1. **Channel-Based Agents**:
   ```go
   specialists := map[string]*agent.ChannelAgent{
       "Science":    agent.NewChannelAgent(specialistConfig("Science Expert", "..."), llm),
       "History":    agent.NewChannelAgent(specialistConfig("History Expert", "..."), llm),
       "Technology": agent.NewChannelAgent(specialistConfig("Tech Expert", "..."), llm),
   }
   ```
   Each agent runs independently with its own input/output channels.

2. **Response Processing**:
   - Each agent gets its own goroutine for handling responses
   - Responses are processed asynchronously
   - Color coding helps distinguish between agents
   - Emoji extraction and bold formatting are handled in real-time

3. **Message Flow**:
   ```go
   // Send questions with slight delays
   for name, qs := range questions {
       specialist := specialists[name]
       for _, q := range qs {
           specialist.Send(q)
           time.Sleep(500 * time.Millisecond)
       }
   }
   ```
   Questions are sent with small delays to make the output more readable.

### Output Formatting

The demo uses the `color` package for rich terminal output:

1. **Color Coding**:
   - Science: Blue
   - History: Yellow
   - Technology: Magenta
   - System messages: Gray
   - Success messages: Green
   - Error messages: Red

2. **Message Structure**:
   ```
   [emoji] [Specialist Name]: [Formatted Response]
   ```
   Example:
   ```
   🔬 Science: The greenhouse effect is a natural process where atmospheric gases trap heat...
   ```

3. **Text Formatting**:
   - Important terms are bolded using ANSI escape codes
   - Emojis are extracted and positioned consistently
   - Clean separators between sections

## Implementation Details

### Error Handling

The demo includes comprehensive error handling:
- Channel full conditions
- Agent startup failures
- Message sending failures
- Response processing errors

### Synchronization

Uses several synchronization mechanisms:
- `context.Context` for cancellation
- `sync.WaitGroup` for coordinated shutdown
- Channel-based communication
- Mutex-protected state

### Cleanup

Implements a clean shutdown process:
1. Cancel context
2. Wait for in-flight messages
3. Close all channels
4. Wait for goroutines to finish

## Running the Demo

```bash
go run examples/concurrent_specialists/main.go
```

The demo will:
1. Start three specialist agents
2. Send domain-specific questions to each
3. Display responses as they arrive
4. Perform a clean shutdown

## Example Output

```
🤖 FastAgent Concurrent Specialists Demo
─────────────────────────────────────────
✓ Started Science specialist
✓ Started History specialist
✓ Started Technology specialist

Sending questions to specialists...
→ Asking Science: What is quantum entanglement?
...
```

## Key Features

1. **Concurrent Processing**:
   - Multiple agents running simultaneously
   - Non-blocking message handling
   - Parallel question processing

2. **Pretty Output**:
   - Color-coded responses
   - Emoji indicators
   - Bold formatting
   - Clean layout

3. **Robust Implementation**:
   - Error handling
   - Clean shutdown
   - Resource cleanup
   - Thread safety

4. **Extensible Design**:
   - Easy to add new specialists
   - Configurable message formatting
   - Flexible response handling

## Technical Foundation

Built on FastAgent's channel-based agent system:
- Uses `ChannelAgent` for message handling
- Leverages Go's concurrency features
- Implements proper synchronization
- Handles backpressure

## Future Enhancements

Possible improvements:
1. Interactive mode for user questions
2. Dynamic specialist loading
3. Response aggregation across specialists
4. Rate limiting and throttling
5. Persistent conversation history
6. Web interface 