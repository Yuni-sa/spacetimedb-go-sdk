# SpacetimeDB Go SDK

## Disclaimer

> **Note**: This project is an independent, unofficial Go SDK for SpacetimeDB. It is not affiliated with or endorsed by Clockwork Labs or the SpacetimeDB team.

---

⚠️ **WARNING: This package is in very early stages and bound to have breaking changes.** ⚠️

Go client SDK for SpacetimeDB that provides access to all HTTP APIs and WebSocket functionality.

## Installation

```bash
go get github.com/Yuni-sa/spacetimedb-go-sdk/client
```

## Quick Start

### Simple Client Usage

```go
package main

import (
    "fmt"
    "log"
    "time"

    "github.com/Yuni-sa/spacetimedb-go-sdk/client"
)

func main() {
	// Create a client
	spacetimeClient, err := client.NewClientBuilder().
		WithBaseURL("http://localhost:3000").
		WithTimeout(30 * time.Second).
		Build()
	if err != nil {
		log.Fatal("Failed to create client:", err)
	}
	defer spacetimeClient.Close()

	// Create identity and get token
	identityResp, err := spacetimeClient.Identity.Create()
	if err != nil {
		log.Fatal("Failed to create identity:", err)
	}

	// Configure client with token
	spacetimeClient.SetToken(identityResp.Token)
	spacetimeClient.SetIdentity(identityResp.Identity)

	fmt.Printf("Connected with identity: %s\n", identityResp.Identity)

	// Get database info
	dbInfo, err := spacetimeClient.Database.GetInfo("my_database")
	if err != nil {
		log.Printf("Database not found: %v", err)
		return
	}

	fmt.Printf("Database owner: %s\n", dbInfo.OwnerIdentity.Identity)

	// Execute SQL query
	results, err := spacetimeClient.Database.ExecuteSQL("my_database", []string{
		"SELECT * FROM users LIMIT 5",
	})
	if err != nil {
		log.Printf("SQL query failed: %v", err)
		return
	}

	fmt.Printf("Query returned %d rows\n", len(results[0].Rows))

	// Call a reducer
	err = spacetimeClient.Database.CallReducer("my_database", "my_message_reducer", []any{
		"Hello from Go SDK!",
	})
	if err != nil {
		log.Printf("Reducer call failed: %v", err)
		return
	}

	fmt.Println("Message sent successfully!")
}
```

### Authentication Token Management

```go
// Initialize auth token with default settings
authToken, err := client.NewAuthToken()
if err != nil {
    log.Fatal(err)
}

// Or with custom configuration
authToken, err := client.NewAuthToken(
    client.WithAuthConfigFolder(".my_app"),
    client.WithAuthConfigFile("my_app_settings.ini"),
    client.WithAuthConfigRoot("/custom/root/path"),
)
if err != nil {
    log.Fatal(err)
}

// Get existing token
token := authToken.GetToken()

// Save a new token to file
err = authToken.SaveToken(token)
if err != nil {
    log.Fatal(err)
}
```

### WebSocket Real-time Connection

```go
// Connect to database via WebSocket
wsConn, err := spacetimeClient.Database.ConnectWebSocket("my-database", client.SatsProtocol)
if err != nil {
    log.Fatal("WebSocket connection failed:", err)
}
defer wsConn.Close()

// Subscribe to tables
err = wsConn.SendSubscribe([]string{"SELECT * FROM my_table"}, 1)
if err != nil {
    log.Fatal("Failed to subscribe:", err)
}

// Listen for messages
for {
    message, err := wsConn.ReceiveMessage()
    if err != nil {
        log.Printf("WebSocket error: %v", err)
        break
    }
    
    fmt.Printf("Received: %T\n", message)
}
```

## Running Tests

```bash
# Run all tests (automatically publishes test database)
make test

# Run tests with verbose output
make test-verbose

# Run specific test
go test ./tests -v
```

## Running Quickstart Chat Example

```bash
# Publish and run the quickstart chat client
make quickstart

# Or manually:
make publish DBNAME=quickstart-chat
go run ./examples/quickstart-chat/client/main.go

# Run multiple clients (for testing)
go run ./examples/quickstart-chat/client/main.go --client 1
go run ./examples/quickstart-chat/client/main.go --client 2
```

## API Reference

### Client Builder

```go
client, err := client.NewClientBuilder().
    WithBaseURL("http://localhost:3000").
    WithToken("your-auth-token").
    WithIdentity("your-identity").
    WithTimeout(30 * time.Second).
    Build()
```

### Identity Service

- `Create()` - Generate new identity and token
- `CreateWebSocketToken()` - Generate short-lived token
- `GetPublicKey()` - Get verification public key
- `SetEmail(identity, email)` - Associate email with identity
- `Verify(identity)` - Verify identity/token pair
- `GetDatabases(identity)` - List owned databases

### Database Service

- `Publish(wasmModule)` - Publish anonymous database
- `PublishTo(name, wasmModule, clear)` - Publish to named database
- `GetInfo(nameOrIdentity)` - Get database information
- `Delete(nameOrIdentity)` - Delete database
- `GetNames(nameOrIdentity)` - Get database names
- `AddName(nameOrIdentity, newName)` - Add database name
- `SetNames(nameOrIdentity, names)` - Set all database names
- `GetIdentity(nameOrIdentity)` - Get database identity
- `ConnectWebSocket(nameOrIdentity, protocol)` - WebSocket connection
- `CallReducer(nameOrIdentity, reducer, args)` - Invoke reducer
- `GetSchema(nameOrIdentity, version)` - Get database schema
- `GetLogs(nameOrIdentity, numLines, follow)` - Get database logs
- `ExecuteSQL(nameOrIdentity, queries)` - Execute SQL queries

### WebSocket Connection

- `SendMessage(message)` - Send WebSocket message
- `ReceiveMessage()` - Receive WebSocket message
- `Close()` - Close connection
- `GracefulClose()` - Gracefully close connection with proper handshake
- `SendSubscribe(queries, requestID)` - Send subscription request for multiple queries
- `SendCallReducer(reducerName, args, requestID)` - Send reducer call request
- `SendOneOffQuery(messageID, queryString)` - Send one-off query request
- `SendSubscribeSingle(query, requestID, queryID)` - Subscribe to single query with ID
- `SendSubscribeMulti(queries, requestID, queryID)` - Subscribe to multiple queries with ID
- `SendUnsubscribe(requestID, queryID)` - Unsubscribe from single query
- `SendUnsubscribeMulti(requestID, queryID)` - Unsubscribe from multiple queries
- `SendSubscribeAll(requestID)` - Subscribe to all tables



## Protocol Support

### Currently Supported
- **JSON Protocol**: `client.SatsProtocol` (`v1.json.spacetimedb`)
- **HTTP APIs**: All SpacetimeDB HTTP endpoints
- **WebSocket**: Real-time subscriptions and updates

### Not Yet Supported
- **BSATN Protocol**: `client.BsatnProtocol` (`v1.bsatn.spacetimedb`) - Binary encoding

## Project Goals

This SDK is actively being developed with the following goals:

- [ ] **Complete HTTP API Coverage** - All SpacetimeDB HTTP endpoints
- [ ] **WebSocket Support** - Real-time database subscriptions
- [ ] **SATS Type System** - Full SpacetimeDB Algebraic Type System support
- [ ] **Authentication Management** - Token persistence and management
- [ ] **BSATN Protocol Support** - Binary encoding for better performance
- [ ] **Enhanced WebSocket Features** - Better message handling and functions
- [ ] **Comprehensive Test Coverage** - More robust testing across all features
- [ ] **Improved Project Structure** - Better organization as the project grows
- [ ] **Documentation** - More examples and detailed API documentation
- [ ] **Performance Optimizations** - Connection pooling, caching, etc.
- [ ] **Code Generation** - Module Bindings and ClientAPI

## Examples

See the `examples/` directory for complete working examples:

- **`examples/quickstart-chat/`** - Chat application with WebSocket real-time updates
- **`tests/client_test.go`** - Test suite 

## Development Status

**Stability**: ⚠️ **Breaking changes expected** ⚠️

**Production Ready**: No - Use for development and testing only

**Feedback**: Issues and contributions welcome!

## Contributing

This project is in early development. Contributions, bug reports, and feature requests are welcome!

## License

This SDK is licensed under the [MIT License](LICENSE).

### About SpacetimeDB's License

SpacetimeDB is licensed under the Business Source License (BSL) 1.1. While this is not an open-source or free software license, it transitions to the GNU Affero General Public License (AGPL) v3.0 with a custom linking exception after a specified period.

The custom linking exception allows developers to link their own code with SpacetimeDB without the requirement to open-source their proprietary code, which is a typical stipulation under the standard AGPL license. :contentReference[oaicite:1]{index=1}

For detailed information, please refer to the [SpacetimeDB License](https://github.com/clockworklabs/spacetimedb/blob/master/LICENSE.txt).
