package client

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocket connection methods

// WebSocketConnection represents a WebSocket connection to a database
type WebSocketConnection struct {
	conn   *websocket.Conn
	client *Client
	dbName string
}

// ConnectWebSocket establishes a WebSocket connection to a database
func (s *DatabaseService) ConnectWebSocket(nameOrIdentity string, protocol string) (*WebSocketConnection, error) {
	// Parse the base URL to extract just the host
	baseURL, err := url.Parse(s.client.baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	// Determine WebSocket scheme based on HTTP scheme
	wsScheme := "ws"
	if baseURL.Scheme == "https" {
		wsScheme = "wss"
	}

	wsURL := url.URL{
		Scheme: wsScheme,
		Host:   baseURL.Host,
		Path:   fmt.Sprintf("/v1/database/%s/subscribe", nameOrIdentity),
	}

	// Validate protocol
	if protocol == "" {
		protocol = SatsProtocol
	}
	if protocol != SatsProtocol && protocol != BsatnProtocol {
		return nil, fmt.Errorf("invalid protocol: %s", protocol)
	}

	// Set up required headers for SpacetimeDB WebSocket connection
	headers := http.Header{
		"Sec-WebSocket-Protocol": []string{protocol},
		"Sec-WebSocket-Version":  []string{"13"},
	}

	if s.client.token != "" {
		headers["Authorization"] = []string{fmt.Sprintf("Bearer %s", s.client.token)}
	}

	dialer := websocket.Dialer{
		HandshakeTimeout: 45 * time.Second,
		Subprotocols:     []string{protocol},
	}

	conn, resp, err := dialer.Dial(wsURL.String(), headers)
	if err != nil {
		if resp != nil {
			return nil, fmt.Errorf("WebSocket handshake failed. Status: %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("error connecting to WebSocket: %w", err)
	}

	return &WebSocketConnection{
		conn:   conn,
		client: s.client,
		dbName: nameOrIdentity,
	}, nil
}

// Close closes the WebSocket connection
func (ws *WebSocketConnection) Close() error {
	if ws.conn != nil {
		return ws.conn.Close()
	}
	return nil
}

func (ws *WebSocketConnection) GracefulClose() error {
	if ws.conn != nil {
		// Send a close message with normal closure code (1000)
		err := ws.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			return fmt.Errorf("error sending close message: %w", err)
		}

		// Wait for the peer to respond
		time.Sleep(100 * time.Millisecond)

		err = ws.conn.Close()
		if err != nil {
			return fmt.Errorf("error closing websocket connection: %w", err)
		}
	}
	return nil
}

// SendSubscribe sends a subscription request
func (ws *WebSocketConnection) SendSubscribe(queries []string, requestID uint32) error {
	subscribeMsg := ClientMessage{
		Subscribe: &Subscribe{
			QueryStrings: queries,
			RequestID:    requestID,
		},
	}
	return ws.SendMessage(subscribeMsg)
}

// SendCallReducer sends a reducer call request
func (ws *WebSocketConnection) SendCallReducer(reducerName string, args string, requestID uint32) error {
	callMsg := ClientMessage{
		CallReducer: &CallReducer{
			Reducer:   reducerName,
			Args:      args,
			RequestID: requestID,
			Flags:     0,
		},
	}
	return ws.SendMessage(callMsg)
}

func (ws *WebSocketConnection) SendOneOffQuery(messageID []byte, queryString string) error {
	queryMsg := ClientMessage{
		OneOffQuery: &OneOffQuery{
			MessageID:   messageID,
			QueryString: queryString,
		},
	}
	return ws.SendMessage(queryMsg)
}

func (ws *WebSocketConnection) SendSubscribeSingle(query string, requestID uint32, queryID QueryID) error {
	subscribeMsg := ClientMessage{
		SubscribeSingle: &SubscribeSingle{
			Query:     query,
			RequestID: requestID,
			QueryID:   queryID,
		},
	}
	return ws.SendMessage(subscribeMsg)
}

func (ws *WebSocketConnection) SendSubscribeMulti(queries []string, requestID uint32, queryID QueryID) error {
	subscribeMsg := ClientMessage{
		SubscribeMulti: &SubscribeMulti{
			QueryStrings: queries,
			RequestID:    requestID,
			QueryID:      queryID,
		},
	}
	return ws.SendMessage(subscribeMsg)
}

func (ws *WebSocketConnection) SendUnsubscribe(requestID uint32, queryID QueryID) error {
	unsubscribeMsg := ClientMessage{
		Unsubscribe: &Unsubscribe{
			RequestID: requestID,
			QueryID:   queryID,
		},
	}
	return ws.SendMessage(unsubscribeMsg)
}

func (ws *WebSocketConnection) SendUnsubscribeMulti(requestID uint32, queryID QueryID) error {
	unsubscribeMsg := ClientMessage{
		UnsubscribeMulti: &UnsubscribeMulti{
			RequestID: requestID,
			QueryID:   queryID,
		},
	}
	return ws.SendMessage(unsubscribeMsg)
}

func (ws *WebSocketConnection) SendSubscribeAll(requestID uint32) error {
	subscribeMsg := ClientMessage{
		Subscribe: &Subscribe{
			RequestID:    requestID,
			QueryStrings: []string{"SELECT * FROM *"},
		},
	}
	return ws.SendMessage(subscribeMsg)
}

// Basic websocket send and receive

// SendMessage sends a message through the WebSocket connection
func (ws *WebSocketConnection) SendMessage(message any) error {
	if ws.conn == nil {
		return fmt.Errorf("WebSocket connection not established")
	}
	return ws.conn.WriteJSON(message)
}

// ReceiveMessage receives a message from the WebSocket connection
func (ws *WebSocketConnection) ReceiveMessage() (any, error) {
	if ws.conn == nil {
		return nil, fmt.Errorf("WebSocket connection not established")
	}

	var message any
	err := ws.conn.ReadJSON(&message)
	if err != nil {
		return nil, fmt.Errorf("error reading message: %w", err)
	}

	return message, nil
}
