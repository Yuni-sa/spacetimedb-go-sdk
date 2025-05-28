package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Yuni-sa/spacetimedb-go-sdk/client"

	"github.com/google/uuid"
)

type User struct {
	Identity string  `json:"__identity__"`
	Name     *string `json:"name"`
	Online   bool    `json:"online"`
}

type Message struct {
	Sender string    `json:"sender"`
	Sent   time.Time `json:"sent"`
	Text   string    `json:"text"`
}

type ChatClient struct {
	localIdentity string
	authToken     *client.AuthToken
	users         map[string]*User
	usersMutex    sync.RWMutex
	messagesMutex sync.RWMutex
	inputQueue    chan Command
	ctx           context.Context
	cancel        context.CancelFunc
	wsConn        *client.WebSocketConnection
	conn          *client.Client
}

type Command struct {
	Type string
	Args string
}

const (
	host   = "http://localhost:3000"
	dbName = "quickstart-chat"
)

func NewChatClient(authToken *client.AuthToken) *ChatClient {
	ctx, cancel := context.WithCancel(context.Background())
	return &ChatClient{
		authToken:  authToken,
		users:      make(map[string]*User),
		inputQueue: make(chan Command, 100),
		ctx:        ctx,
		cancel:     cancel,
	}
}

func main() {
	// Initialize AuthToken with user-specific filename
	authToken, err := client.NewAuthToken(client.WithAuthConfigFolder(".spacetime_go_quickstart_chat"))
	if err != nil {
		log.Fatalf("Failed to create auth token: %v", err)
	}

	// Create chat client
	chatClient := NewChatClient(authToken)
	defer chatClient.cancel()

	// Connect to database
	conn, err := chatClient.connectToDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close()

	chatClient.conn = conn

	// Start processing thread
	go chatClient.processThread()

	// Start input loop
	go chatClient.inputLoop()

	<-chatClient.ctx.Done()

	if chatClient.wsConn != nil {
		chatClient.wsConn.GracefulClose()
	}
}

// connectToDB creates and configures the database connection
func (c *ChatClient) connectToDB() (*client.Client, error) {
	builder := client.NewClientBuilder().
		WithBaseURL(host).
		WithTimeout(30 * time.Second)

	if token := c.authToken.GetToken(); token != "" {
		builder = builder.WithToken(token)
	}

	conn, err := builder.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build client: %w", err)
	}

	// Test connection
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return conn, nil
}

func (c *ChatClient) processThread() {
	defer func() {
		if c.wsConn != nil {
			c.wsConn.GracefulClose()
		}
	}()

	// Create identity if we don't have a token
	if c.authToken.GetToken() == "" {
		if err := c.createIdentity(); err != nil {
			log.Printf("Failed to create identity: %v", err)
			return
		}
	}

	// Establish WebSocket connection
	var err error
	c.wsConn, err = c.conn.Database.ConnectWebSocket(dbName, client.SatsProtocol)
	if err != nil {
		log.Printf("Failed to connect WebSocket: %v", err)
		c.cancel()
		return
	}

	fmt.Println("SpacetimeDB websocket connected")

	// Subscribe to all tables
	if err := c.wsConn.SendSubscribeAll(uuid.New().ID()); err != nil {
		log.Printf("Failed to subscribe: %v", err)
		return
	}

	// Start message processing goroutines
	go c.handleWebSocketMessages()
	go c.processCommands()

	// Keep the process thread alive
	<-c.ctx.Done()
}

func (c *ChatClient) printMessagesInOrder(initSub *client.InitialSubscription) {
	c.messagesMutex.RLock()
	defer c.messagesMutex.RUnlock()

	for _, table := range initSub.DatabaseUpdate.Tables {
		if table.TableName == "message" {
			for _, update := range table.Updates {
				for _, insert := range update.Inserts {
					c.printMessage(parseJSONMessage(insert))
				}
			}
		}
	}
}

// createIdentity creates a new identity and saves the token
func (c *ChatClient) createIdentity() error {
	resp, err := c.conn.Identity.Create()
	if err != nil {
		return fmt.Errorf("failed to create identity: %w", err)
	}

	c.localIdentity = resp.Identity
	c.authToken.SaveToken(resp.Token)
	c.conn.SetToken(resp.Token)
	c.conn.SetIdentity(resp.Identity)

	return nil
}

func (c *ChatClient) handleWebSocketMessages() {
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			msg, err := c.wsConn.ReceiveMessage()
			if err != nil {
				log.Printf("WebSocket receive error: %v", err)
				return
			}

			if err := c.processWebSocketMessage(msg); err != nil {
				log.Printf("Error processing WebSocket message: %v", err)
			}
		}
	}
}

// processWebSocketMessage handles different types of WebSocket messages
func (c *ChatClient) processWebSocketMessage(msg any) error {
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	serverMsg, err := client.ParseServerMessage(msgBytes)
	if err != nil {
		return err
	}

	if txUpdate, ok := serverMsg.AsTransactionUpdate(); ok {
		return c.handleTransactionUpdate(txUpdate)
	}

	if initSub, ok := serverMsg.AsInitialSubscription(); ok {
		c.printMessagesInOrder(initSub)
		return nil
	}

	if _, ok := serverMsg.AsIdentityToken(); ok {
		fmt.Println("Connected to SpacetimeDB")
		return nil
	}

	return nil
}

// handleTransactionUpdate processes transaction update messages
func (c *ChatClient) handleTransactionUpdate(txUpdate *client.TransactionUpdate) error {
	// Only process committed transactions
	if txUpdate.Status.Committed != nil {
		for _, tableUpdate := range txUpdate.Status.Committed.Tables {
			switch txUpdate.ReducerCall.ReducerName {
			case "SendMessage":
				c.onSentMessage(tableUpdate)
			case "SetName":
				c.onSetName(tableUpdate)
			case "ClientConnected":
				c.onUserStatusChange(tableUpdate, true)
			case "ClientDisconnected":
				c.onUserStatusChange(tableUpdate, false)
			}
		}
	}

	return nil
}

func (c *ChatClient) onSetName(update client.TableUpdate) {
	for _, updateEntry := range update.Updates {
		for _, insertStr := range updateEntry.Inserts {
			user := parseUser(insertStr)
			if user == nil {
				continue
			}
			c.usersMutex.Lock()
			c.users[user.Identity] = user
			c.usersMutex.Unlock()
		}
	}
}

func (c *ChatClient) onUserStatusChange(update client.TableUpdate, isConnecting bool) {
	for _, updateEntry := range update.Updates {
		for _, insertStr := range updateEntry.Inserts {
			user := parseUser(insertStr)
			if user == nil {
				continue
			}
			c.usersMutex.Lock()
			user.Online = isConnecting
			c.users[user.Identity] = user
			c.usersMutex.Unlock()

			// Print user status change
			if isConnecting {
				fmt.Printf("%s is online\n", c.getUserDisplayName(user.Identity))
			} else {
				fmt.Printf("%s is offline\n", c.getUserDisplayName(user.Identity))
			}
		}
	}
}

func (c *ChatClient) onSentMessage(update client.TableUpdate) {
	c.messagesMutex.Lock()
	defer c.messagesMutex.Unlock()

	for _, updateEntry := range update.Updates {
		for _, insertStr := range updateEntry.Inserts {
			if message := parseMessage(insertStr); message != nil {
				// Get display name for sender
				c.printMessage(message)
			}
		}
	}
}

func (c *ChatClient) printMessage(message *Message) {
	senderName := c.getUserDisplayName(message.Sender)
	fmt.Printf("%s: \"%s\"\n", senderName, message.Text)
}

func parseJSONMessage(jsonRaw string) *Message {
	var raw struct {
		Sender struct {
			Identity string `json:"__identity__"`
		} `json:"Sender"`
		Sent struct {
			Micros int64 `json:"__timestamp_micros_since_unix_epoch__"`
		} `json:"Sent"`
		Text string `json:"Text"`
	}
	if err := json.Unmarshal([]byte(jsonRaw), &raw); err != nil {
		return nil
	}
	return &Message{
		Sender: raw.Sender.Identity,
		Sent:   time.UnixMicro(raw.Sent.Micros),
		Text:   raw.Text,
	}
}

func parseMessage(insertStr string) *Message {
	// Expected format: [["sender_identity"], [timestamp_micros], "text"]
	var raw []any
	if err := json.Unmarshal([]byte(insertStr), &raw); err != nil {
		return nil
	}
	if len(raw) != 3 {
		return nil
	}

	// Parse sender (inside a slice)
	senderSlice, ok := raw[0].([]any)
	if !ok || len(senderSlice) != 1 {
		return nil
	}
	sender, ok := senderSlice[0].(string)
	if !ok {
		return nil
	}

	// Parse timestamp (inside a slice)
	timestampSlice, ok := raw[1].([]any)
	if !ok || len(timestampSlice) != 1 {
		return nil
	}
	timestampFloat, ok := timestampSlice[0].(float64)
	if !ok {
		return nil
	}
	sent := time.UnixMicro(int64(timestampFloat))

	// Parse text
	text, ok := raw[2].(string)
	if !ok {
		return nil
	}

	return &Message{
		Sender: sender,
		Sent:   sent,
		Text:   text,
	}
}

func parseUser(insertStr string) *User {
	// Expected format: [[identity], [tag, name], online]
	var raw []any
	if err := json.Unmarshal([]byte(insertStr), &raw); err != nil {
		return nil
	}

	if len(raw) != 3 {
		return nil
	}

	// Parse identity (inside a slice)
	identitySlice, ok := raw[0].([]any)
	if !ok || len(identitySlice) != 1 {
		return nil
	}
	identity, ok := identitySlice[0].(string)
	if !ok {
		return nil
	}

	// Parse name (inside a slice with tag and value)
	var name *string
	nameSlice, ok := raw[1].([]any)
	if ok && len(nameSlice) == 2 {
		// [tag, value] format where tag is 0 for Some, 1 for None
		if tag, ok := nameSlice[0].(float64); ok && tag == 0 {
			if nameStr, ok := nameSlice[1].(string); ok {
				name = &nameStr
			}
		}
	}

	// Parse online
	online, ok := raw[2].(bool)
	if !ok {
		return nil
	}

	return &User{
		Identity: identity,
		Name:     name,
		Online:   online,
	}
}

func (c *ChatClient) getUserDisplayName(identity string) string {
	if identity == "" {
		return "unknown"
	}

	c.usersMutex.RLock()
	user, exists := c.users[identity]
	c.usersMutex.RUnlock()

	if exists && user.Name != nil && *user.Name != "" {
		return *user.Name
	}

	if result, err := c.conn.Database.ExecuteSQL(dbName,
		[]string{
			fmt.Sprintf("SELECT Name FROM user WHERE Identity = '%s'", identity),
		},
	); err == nil && len(result) > 0 && len(result[0].Rows) > 0 {
		if row, ok := result[0].Rows[0].([]any); ok && len(row) > 0 {
			if nameVal := row[0].([]any); nameVal != nil {
				// [tag, value]
				if nameStr, ok := nameVal[1].(string); ok && nameStr != "" {
					c.usersMutex.RLock()
					if user, exists := c.users[identity]; exists {
						user.Name = &nameStr
					}
					c.usersMutex.RUnlock()
					return nameStr
				}
			}
		}
	}

	// Return first 8 characters of identity as fallback
	if len(identity) >= 8 {
		return identity[:8]
	}

	return identity
}

func (c *ChatClient) processCommands() {
	for {
		select {
		case <-c.ctx.Done():
			return
		case cmd := <-c.inputQueue:
			switch cmd.Type {
			case "message":
				if err := c.wsConn.SendCallReducer("SendMessage", fmt.Sprintf(`{"text":%q}`, cmd.Args), 0); err != nil {
					log.Printf("Failed to send message %v: %v", cmd.Args, err)
				}
			case "name":
				if err := c.wsConn.SendCallReducer("SetName", fmt.Sprintf(`{"name":%q}`, cmd.Args), 0); err != nil {
					log.Printf("Failed to set name %v: %v", cmd.Args, err)
				}
			}
		}
	}
}

// inputLoop reads user input and queues commands
func (c *ChatClient) inputLoop() {
	fmt.Println("SpacetimeDB Go Chat Client")
	fmt.Println("Commands:")
	fmt.Println("  /name <your_name>  - Set your name")
	fmt.Println("  <message>          - Send a message")
	fmt.Println("  /quit              - Exit")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		if input == "/quit" {
			c.cancel()
			break
		}

		if strings.HasPrefix(input, "/name ") {
			name := strings.TrimSpace(input[6:])
			if name != "" {
				select {
				case c.inputQueue <- Command{Type: "name", Args: name}:
				default:
					log.Println("Command queue full")
				}
			}
		} else if input != "" {
			select {
			case c.inputQueue <- Command{Type: "message", Args: input}:
			default:
				log.Println("Command queue full")
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Input error: %v", err)
	}
}
