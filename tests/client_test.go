package tests

import (
	"testing"
	"time"

	"github.com/Yuni-sa/spacetimedb-go-sdk/client"
)

const (
	testBaseURL = "http://localhost:3000"
	testDBName  = "test"
)

func TestClientConnectivity(t *testing.T) {
	// Create a client using the builder pattern
	spacetimeClient, err := client.NewClientBuilder().
		WithBaseURL(testBaseURL).
		WithTimeout(5 * time.Second).
		Build()

	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer spacetimeClient.Close()

	// Test connectivity
	t.Log("Testing connectivity to SpacetimeDB")
	if err := spacetimeClient.Ping(); err != nil {
		t.Fatalf("Failed to ping SpacetimeDB: %v", err)
	}
	t.Log("Client connectivity test passed")
}

func TestIdentityOperations(t *testing.T) {
	spacetimeClient, err := client.NewClientBuilder().
		WithBaseURL(testBaseURL).
		WithTimeout(5 * time.Second).
		Build()

	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer spacetimeClient.Close()

	// Create a new identity
	t.Log("Creating new identity...")
	identityResp, err := spacetimeClient.Identity.Create()
	if err != nil {
		t.Fatalf("Failed to create identity: %v", err)
	}

	if identityResp.Identity == "" {
		t.Fatal("Identity should not be empty")
	}
	if identityResp.Token == "" {
		t.Fatal("Token should not be empty")
	}

	t.Logf("Created identity: %s", identityResp.Identity)
	t.Logf("Received token: %s...", identityResp.Token[:20])

	// Configure client with the new identity and token
	spacetimeClient.SetToken(identityResp.Token)
	spacetimeClient.SetIdentity(identityResp.Identity)

	// Verify the identity
	t.Log("Verifying identity...")
	if err := spacetimeClient.Identity.Verify(identityResp.Identity); err != nil {
		t.Fatalf("Failed to verify identity: %v", err)
	}
	t.Log("Identity verified successfully")

	// Get databases owned by this identity
	t.Log("Listing owned databases...")
	databases, err := spacetimeClient.Identity.GetDatabases(identityResp.Identity)
	if err != nil {
		t.Fatalf("Failed to get databases: %v", err)
	}
	t.Logf("Found %d owned databases: %v", len(databases), databases)
}

func TestDatabaseOperations(t *testing.T) {
	spacetimeClient, err := client.NewClientBuilder().
		WithBaseURL(testBaseURL).
		WithTimeout(5 * time.Second).
		Build()

	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer spacetimeClient.Close()

	// Create identity first
	identityResp, err := spacetimeClient.Identity.Create()
	if err != nil {
		t.Fatalf("Failed to create identity: %v", err)
	}
	spacetimeClient.SetToken(identityResp.Token)
	spacetimeClient.SetIdentity(identityResp.Identity)

	// Try to get database info
	t.Logf("Trying to get info for database '%s'...", testDBName)
	dbInfo, err := spacetimeClient.Database.GetInfo(testDBName)
	if err != nil {
		t.Logf("Database '%s' not found or accessible: %v", testDBName, err)
		t.Skip("Skipping database tests - database not accessible")
		return
	}

	t.Logf("Database identity: %s", dbInfo.DatabaseIdentity.Identity)
	t.Logf("Owner identity: %s", dbInfo.OwnerIdentity.Identity)
	t.Logf("Initial program hash: %s", dbInfo.InitialProgram)

	// Get database schema
	t.Log("Retrieving database schema...")
	schema, err := spacetimeClient.Database.GetSchema(testDBName, nil)
	if err != nil {
		t.Logf("Failed to get schema: %v", err)
	} else {
		t.Logf("Schema retrieved (type: %T)", schema)
		if len(schema.Tables) > 0 {
			t.Logf("Found %d tables in schema", len(schema.Tables))
		}
		if len(schema.Reducers) > 0 {
			t.Logf("Found %d reducers in schema", len(schema.Reducers))
		}
	}
}

func TestSQLQueries(t *testing.T) {
	spacetimeClient, err := client.NewClientBuilder().
		WithBaseURL(testBaseURL).
		WithTimeout(5 * time.Second).
		Build()

	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer spacetimeClient.Close()

	// Create identity first
	identityResp, err := spacetimeClient.Identity.Create()
	if err != nil {
		t.Fatalf("Failed to create identity: %v", err)
	}
	spacetimeClient.SetToken(identityResp.Token)
	spacetimeClient.SetIdentity(identityResp.Identity)

	// Check if database exists
	_, err = spacetimeClient.Database.GetInfo(testDBName)
	if err != nil {
		t.Skipf("Skipping SQL tests - database '%s' not accessible: %v", testDBName, err)
		return
	}

	// Example SQL queries
	t.Log("Executing test SQL queries...")

	// Test basic table queries
	testQueries := []string{
		"SELECT * FROM user LIMIT 5",
		"SELECT * FROM message LIMIT 5",
	}

	for i, query := range testQueries {
		t.Logf("Executing query %d: %s", i+1, query)
		results, err := spacetimeClient.Database.ExecuteSQL(testDBName, []string{query})
		if err != nil {
			t.Logf("Query %d failed: %v", i+1, err)
			continue
		}

		if len(results) > 0 {
			t.Logf("Query %d returned %d rows", i+1, len(results[0].Rows))
			// Log first few rows for debugging
			for j, row := range results[0].Rows {
				if j >= 3 { // Limit output
					break
				}
				t.Logf("  Row %d: %v", j+1, row)
			}
		}
	}
}

func TestWebSocketConnection(t *testing.T) {
	spacetimeClient, err := client.NewClientBuilder().
		WithBaseURL(testBaseURL).
		WithTimeout(5 * time.Second).
		Build()

	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer spacetimeClient.Close()

	// Create identity first
	identityResp, err := spacetimeClient.Identity.Create()
	if err != nil {
		t.Fatalf("Failed to create identity: %v", err)
	}
	spacetimeClient.SetToken(identityResp.Token)
	spacetimeClient.SetIdentity(identityResp.Identity)

	// Check if database exists
	_, err = spacetimeClient.Database.GetInfo(testDBName)
	if err != nil {
		t.Skipf("Skipping WebSocket tests - database '%s' not accessible: %v", testDBName, err)
		return
	}

	// Establish WebSocket connection
	t.Log("Establishing WebSocket connection...")
	wsConn, err := spacetimeClient.Database.ConnectWebSocket(testDBName, "v1.json.spacetimedb")
	if err != nil {
		t.Fatalf("Failed to connect WebSocket: %v", err)
	}
	defer wsConn.Close()

	t.Log("WebSocket connected successfully")

	// Subscribe to tables (similar to chat client)
	t.Log("Subscribing to tables...")
	subscribeMsg := map[string]any{
		"Subscribe": map[string]any{
			"query_strings": []string{"SELECT * FROM user", "SELECT * FROM message"},
			"request_id":    1,
		},
	}

	err = wsConn.SendMessage(subscribeMsg)
	if err != nil {
		t.Logf("Failed to send subscription message: %v", err)
	} else {
		t.Log("Subscription message sent successfully")
	}

	// Listen for initial messages
	t.Log("Listening for initial WebSocket messages...")
	messageCount := 0
	maxMessages := 5
	timeout := time.After(10 * time.Second)

	for messageCount < maxMessages {
		select {
		case <-timeout:
			t.Log("WebSocket listening timeout reached")
			goto endListening
		default:
			// Use a goroutine with timeout for message receiving
			messageChan := make(chan any, 1)
			errorChan := make(chan error, 1)

			go func() {
				message, err := wsConn.ReceiveMessage()
				if err != nil {
					errorChan <- err
				} else {
					messageChan <- message
				}
			}()

			select {
			case message := <-messageChan:
				messageCount++
				t.Logf("Received WebSocket message %d: %T", messageCount, message)

				// Try to parse the message to understand its structure
				if msgMap, ok := message.(map[string]any); ok {
					if _, hasIdentityToken := msgMap["IdentityToken"]; hasIdentityToken {
						t.Log("  -> IdentityToken message received")
					}
					if _, hasInitialSub := msgMap["InitialSubscription"]; hasInitialSub {
						t.Log("  -> InitialSubscription message received")
					}
					if txUpdate, hasTxUpdate := msgMap["TransactionUpdate"]; hasTxUpdate {
						t.Log("  -> TransactionUpdate message received")
						// Log some details about the transaction update
						if txMap, ok := txUpdate.(map[string]any); ok {
							if status, hasStatus := txMap["Status"]; hasStatus {
								t.Logf("     Transaction status: %T", status)
							}
						}
					}
				}
			case err := <-errorChan:
				t.Logf("WebSocket receive error (expected): %v", err)
				goto endListening
			case <-time.After(2 * time.Second):
				t.Log("Individual message timeout reached")
				goto endListening
			}
		}
	}

endListening:
	if messageCount > 0 {
		t.Logf("Successfully received %d WebSocket messages", messageCount)
	}

	// Test sending a reducer call and listening for response
	t.Log("Testing reducer call via WebSocket...")
	err = spacetimeClient.Database.CallReducer(testDBName, "SetName", []any{"TestUser_WebSocket"})
	if err != nil {
		t.Logf("Failed to call SetName reducer: %v", err)
	} else {
		t.Log("Successfully called SetName reducer")

		// Listen for any resulting messages
		t.Log("Listening for reducer response messages...")
		responseTimeout := time.After(5 * time.Second)
		responseCount := 0

		for responseCount < 3 {
			select {
			case <-responseTimeout:
				t.Log("Reducer response timeout reached")
				goto endResponseListening
			default:
				// Use a goroutine with timeout for message receiving
				responseChan := make(chan any, 1)
				responseErrorChan := make(chan error, 1)

				go func() {
					message, err := wsConn.ReceiveMessage()
					if err != nil {
						responseErrorChan <- err
					} else {
						responseChan <- message
					}
				}()

				select {
				case message := <-responseChan:
					responseCount++
					t.Logf("Received reducer response message %d: %T", responseCount, message)
				case err := <-responseErrorChan:
					t.Logf("No more reducer response messages: %v", err)
					goto endResponseListening
				case <-time.After(1 * time.Second):
					t.Log("Individual response timeout reached")
					goto endResponseListening
				}
			}
		}

	endResponseListening:
		if responseCount > 0 {
			t.Logf("Received %d reducer response messages", responseCount)
		}
	}

	// Test sending a message if SendMessage reducer exists
	t.Log("Testing SendMessage reducer...")
	err = spacetimeClient.Database.CallReducer(testDBName, "SendMessage", []any{"Test message from WebSocket test"})
	if err != nil {
		t.Logf("SendMessage reducer not available or failed: %v", err)
	} else {
		t.Log("Successfully called SendMessage reducer")
	}

	t.Log("WebSocket connection test completed successfully")
}

func TestReducerCalls(t *testing.T) {
	spacetimeClient, err := client.NewClientBuilder().
		WithBaseURL(testBaseURL).
		WithTimeout(5 * time.Second).
		Build()

	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer spacetimeClient.Close()

	// Create identity first
	identityResp, err := spacetimeClient.Identity.Create()
	if err != nil {
		t.Fatalf("Failed to create identity: %v", err)
	}
	spacetimeClient.SetToken(identityResp.Token)
	spacetimeClient.SetIdentity(identityResp.Identity)

	// Check if database exists
	_, err = spacetimeClient.Database.GetInfo(testDBName)
	if err != nil {
		t.Skipf("Skipping reducer tests - database '%s' not accessible: %v", testDBName, err)
		return
	}

	// Test common reducer calls
	testCases := []struct {
		name    string
		reducer string
		args    []any
	}{
		{
			name:    "SetName",
			reducer: "SetName",
			args:    []any{"TestUser"},
		},
		{
			name:    "SendMessage",
			reducer: "SendMessage",
			args:    []any{"Hello from Go SDK test!"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing reducer: %s", tc.reducer)
			err := spacetimeClient.Database.CallReducer(testDBName, tc.reducer, tc.args)
			if err != nil {
				t.Logf("Reducer %s failed (may not exist): %v", tc.reducer, err)
			} else {
				t.Logf("Successfully called reducer %s", tc.reducer)
			}
		})
	}
}
