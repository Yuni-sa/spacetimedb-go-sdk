# SpacetimeDB Go Chat Client

This is a Go implementation of the SpacetimeDB chat client.

## Features

- **Real-time messaging**: Send and receive chat messages in real-time
- **User management**: Set your name and see when users come online/offline
- **Authentication**: Automatic identity creation and token persistence
- **WebSocket connection**: Live updates through SpacetimeDB's WebSocket API
- **Command-line interface**: Simple text-based chat interface

## Prerequisites

- Go 1.24 or later
- A running SpacetimeDB instance (default: `http://localhost:3000`)
- The quickstart-chat module deployed to your SpacetimeDB instance

## Setup

1. **Ensure you have the chat module deployed**:
   Follow the [SpacetimeDB module quickstart](https://spacetimedb.com/docs/modules/c-sharp/quickstart) to deploy the chat module.

2. **Build the client**:
   ```bash
   go mod tidy
   go build -o chat-client main.go
   ```

3. **Run the client**:
   ```bash
   ./chat-client
   ```
   
   **Run multiple clients locally**:
   ```bash
   ./chat-client --client <client>
   ```

## Usage

### Starting the Client

When you run the client for the first time:

1. It will automatically create a new SpacetimeDB identity
2. Connect to the database via WebSocket
3. Subscribe to user and message updates
4. Display any existing messages in chronological order

### Sending Messages

Simply type your message and press Enter:
```
Hello, everyone!
```

### Setting Your Name

Use the `/name` command to set your display name:
```
/name Alice
```

### Example Session

```
SpacetimeDB Go Chat Client
Commands:
  /name <your_name>  - Set your name
  <message>          - Send a message
  /quit              - Exit

SpacetimeDB websocket connected
Connected to SpacetimeDB
Initial subscription applied
/name Alice
Bob is online
Alice: Hello, world!
Bob: Hi Alice!
Alice: Nice to meet you, Bob!
Charlie is online
Charlie: Hey everyone!
Bob is offline
```

## Configuration

You can modify these constants in `main.go`:

```go
const (
    host    = "http://localhost:3000"  // SpacetimeDB server URL
    dbName = "quickstart-chat"        // Database name
)
```

## Token Storage

The client automatically saves your authentication token to `~/.spacetime_go_quickstart` so you maintain the same identity across sessions.

## Error Handling

The client provides user-friendly error messages for:
- Failed name changes (empty names, etc.)
- Failed message sends (empty messages, etc.)
- Connection issues
- Authentication problems

## Commands

- **Regular text**: Send as a chat message
- **`/name <name>`**: Set your display name
- **`/quit`** or **Ctrl+C**: Exit the client
