package client

import (
	"encoding/json"
	"fmt"
)

// ServerMessageType represents the type of server message
type ServerMessageType int

const ( // maybe use interface instead of enum (will still have boilerplate)
	ServerMessageTypeInitialSubscription ServerMessageType = iota
	ServerMessageTypeTransactionUpdate
	ServerMessageTypeTransactionUpdateLight
	ServerMessageTypeIdentityToken
	ServerMessageTypeOneOffQueryResponse
	ServerMessageTypeSubscribeApplied
	ServerMessageTypeUnsubscribeApplied
	ServerMessageTypeSubscriptionError
	ServerMessageTypeSubscribeMultiApplied
	ServerMessageTypeUnsubscribeMultiApplied
)

// ServerMessage represents all possible server-to-client messages
type ServerMessage struct {
	Type    ServerMessageType `json:"-"`
	Payload any               `json:"-"`
}

// Type-safe getters for each message type
func (sm *ServerMessage) AsInitialSubscription() (*InitialSubscription, bool) {
	if sm.Type == ServerMessageTypeInitialSubscription {
		return sm.Payload.(*InitialSubscription), true
	}
	return nil, false
}

func (sm *ServerMessage) AsTransactionUpdate() (*TransactionUpdate, bool) {
	if sm.Type == ServerMessageTypeTransactionUpdate {
		return sm.Payload.(*TransactionUpdate), true
	}
	return nil, false
}

func (sm *ServerMessage) AsTransactionUpdateLight() (*TransactionUpdateLight, bool) {
	if sm.Type == ServerMessageTypeTransactionUpdateLight {
		return sm.Payload.(*TransactionUpdateLight), true
	}
	return nil, false
}

func (sm *ServerMessage) AsIdentityToken() (*IdentityToken, bool) {
	if sm.Type == ServerMessageTypeIdentityToken {
		return sm.Payload.(*IdentityToken), true
	}
	return nil, false
}

func (sm *ServerMessage) AsOneOffQueryResponse() (*OneOffQueryResponse, bool) {
	if sm.Type == ServerMessageTypeOneOffQueryResponse {
		return sm.Payload.(*OneOffQueryResponse), true
	}
	return nil, false
}

func (sm *ServerMessage) AsSubscribeApplied() (*SubscribeApplied, bool) {
	if sm.Type == ServerMessageTypeSubscribeApplied {
		return sm.Payload.(*SubscribeApplied), true
	}
	return nil, false
}

func (sm *ServerMessage) AsUnsubscribeApplied() (*UnsubscribeApplied, bool) {
	if sm.Type == ServerMessageTypeUnsubscribeApplied {
		return sm.Payload.(*UnsubscribeApplied), true
	}
	return nil, false
}

func (sm *ServerMessage) AsSubscriptionError() (*SubscriptionError, bool) {
	if sm.Type == ServerMessageTypeSubscriptionError {
		return sm.Payload.(*SubscriptionError), true
	}
	return nil, false
}

func (sm *ServerMessage) AsSubscribeMultiApplied() (*SubscribeMultiApplied, bool) {
	if sm.Type == ServerMessageTypeSubscribeMultiApplied {
		return sm.Payload.(*SubscribeMultiApplied), true
	}
	return nil, false
}

func (sm *ServerMessage) AsUnsubscribeMultiApplied() (*UnsubscribeMultiApplied, bool) {
	if sm.Type == ServerMessageTypeUnsubscribeMultiApplied {
		return sm.Payload.(*UnsubscribeMultiApplied), true
	}
	return nil, false
}

// InitialSubscription represents the initial subscription response
type InitialSubscription struct {
	DatabaseUpdate             DatabaseUpdate `json:"database_update"`
	RequestID                  uint32         `json:"request_id"`
	TotalHostExecutionDuration TimeDuration   `json:"total_host_execution_duration"`
}

// TransactionUpdate represents a transaction update
type TransactionUpdate struct {
	Status                     UpdateStatus    `json:"status"`
	Timestamp                  Timestamp       `json:"timestamp"`
	CallerIdentity             Identity        `json:"caller_identity"`
	CallerConnectionID         ConnectionID    `json:"caller_connection_id"`
	ReducerCall                ReducerCallInfo `json:"reducer_call"`
	EnergyQuantaUsed           EnergyQuanta    `json:"energy_quanta_used"`
	TotalHostExecutionDuration TimeDuration    `json:"total_host_execution_duration"`
}

// TransactionUpdateLight represents a lightweight transaction update
type TransactionUpdateLight struct {
	RequestID uint32         `json:"request_id"`
	Update    DatabaseUpdate `json:"update"`
}

// IdentityToken represents an identity token message
type IdentityToken struct {
	Identity     Identity     `json:"identity"`
	Token        string       `json:"token"`
	ConnectionID ConnectionID `json:"connection_id"`
}

// OneOffQueryResponse represents a one-off query response
type OneOffQueryResponse struct {
	MessageID                  []byte        `json:"message_id"`
	Error                      *string       `json:"error,omitempty"`
	Tables                     []OneOffTable `json:"tables"`
	TotalHostExecutionDuration TimeDuration  `json:"total_host_execution_duration"`
}

// SubscribeApplied represents a subscription applied response
type SubscribeApplied struct {
	RequestID                        uint32        `json:"request_id"`
	TotalHostExecutionDurationMicros uint64        `json:"total_host_execution_duration_micros"`
	QueryID                          QueryID       `json:"query_id"`
	Rows                             SubscribeRows `json:"rows"`
}

// UnsubscribeApplied represents an unsubscription applied response
type UnsubscribeApplied struct {
	RequestID                        uint32        `json:"request_id"`
	TotalHostExecutionDurationMicros uint64        `json:"total_host_execution_duration_micros"`
	QueryID                          QueryID       `json:"query_id"`
	Rows                             SubscribeRows `json:"rows"`
}

// SubscriptionError represents a subscription error
type SubscriptionError struct {
	TotalHostExecutionDurationMicros uint64  `json:"total_host_execution_duration_micros"`
	RequestID                        *uint32 `json:"request_id,omitempty"`
	QueryID                          *uint32 `json:"query_id,omitempty"`
	TableID                          *uint32 `json:"table_id,omitempty"`
	Error                            string  `json:"error"`
}

// SubscribeMultiApplied represents a multi-subscription applied response
type SubscribeMultiApplied struct {
	RequestID                        uint32         `json:"request_id"`
	TotalHostExecutionDurationMicros uint64         `json:"total_host_execution_duration_micros"`
	QueryID                          QueryID        `json:"query_id"`
	Update                           DatabaseUpdate `json:"update"`
}

// UnsubscribeMultiApplied represents a multi-unsubscription applied response
type UnsubscribeMultiApplied struct {
	RequestID                        uint32         `json:"request_id"`
	TotalHostExecutionDurationMicros uint64         `json:"total_host_execution_duration_micros"`
	QueryID                          QueryID        `json:"query_id"`
	Update                           DatabaseUpdate `json:"update"`
}

// Supporting types

// DatabaseUpdate represents a database update
type DatabaseUpdate struct {
	Tables []TableUpdate `json:"tables"`
}

type TableUpdate struct {
	TableName string             `json:"table_name"`
	NumRows   uint32             `json:"num_rows"`
	TableID   uint32             `json:"table_id"`
	Updates   []TableUpdateEntry `json:"updates"`
}

type TableUpdateEntry struct {
	Inserts []string `json:"inserts,omitempty"`
	Deletes []string `json:"deletes,omitempty"`
}

// TableUpdate represents a table update
//type TableUpdate struct {
//	TableID   uint32                    `json:"table_id"`
//	TableName string                    `json:"table_name"`
//	NumRows   uint64                    `json:"num_rows"`
//	Updates   []CompressableQueryUpdate `json:"updates"`
//}

// CompressableQueryUpdate represents a compressable query update
type CompressableQueryUpdate struct {
	Uncompressed *QueryUpdate `json:"Uncompressed,omitempty"`
	Brotli       []byte       `json:"Brotli,omitempty"`
	Gzip         []byte       `json:"Gzip,omitempty"`
}

// QueryUpdate represents a query update
type QueryUpdate struct {
	Deletes BsatnRowList `json:"deletes"`
	Inserts BsatnRowList `json:"inserts"`
}

// BsatnRowList represents a BSATN row list
type BsatnRowList struct {
	SizeHint RowSizeHint `json:"size_hint"`
	RowsData []byte      `json:"rows_data"`
}

// RowSizeHint represents a row size hint
type RowSizeHint struct {
	FixedSize  *uint16  `json:"FixedSize,omitempty"`
	RowOffsets []uint64 `json:"RowOffsets,omitempty"`
}

// OneOffTable represents a one-off table
type OneOffTable struct {
	TableName string       `json:"table_name"`
	Rows      BsatnRowList `json:"rows"`
}

// SubscribeRows represents subscription rows
type SubscribeRows struct {
	TableID   uint32      `json:"table_id"`
	TableName string      `json:"table_name"`
	TableRows TableUpdate `json:"table_rows"`
}

// ReducerCallInfo represents reducer call information
type ReducerCallInfo struct {
	ReducerName    string          `json:"reducer_name"`
	Args           json.RawMessage `json:"args"`
	Status         string          `json:"status"`
	ReducerID      uint32          `json:"reducer_id"`
	RequestID      uint32          `json:"request_id"`
	CallerIdentity string          `json:"caller_identity,omitempty"`
	Error          *string         `json:"error,omitempty"`
}

// UpdateStatus represents an update status
type UpdateStatus struct {
	Committed   *DatabaseUpdate `json:"Committed,omitempty"`
	Failed      *string         `json:"Failed,omitempty"`
	OutOfEnergy any             `json:"OutOfEnergy,omitempty"`
}

// EnergyQuanta represents energy quanta
type EnergyQuanta struct {
	Quanta uint64 `json:"quanta"`
}

// SpacetimeDB primitive types
type Identity struct {
	Identity string `json:"__identity__"`
}

type ConnectionID struct {
	ConnectionID float64 `json:"__connection_id__"`
}
type Timestamp struct {
	Timestamp uint64 `json:"__timestamp_micros_since_unix_epoch__"`
}

type TimeDuration struct {
	Duration uint64 `json:"__time_duration_micros__"`
}

// ParseServerMessage parses a raw JSON message into a ServerMessage
func ParseServerMessage(data []byte) (*ServerMessage, error) {
	// First, try to parse as a tagged enum (new format)
	var taggedMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &taggedMsg); err == nil {
		for msgType, payload := range taggedMsg {
			switch msgType {
			case "InitialSubscription":
				var v InitialSubscription
				if err := json.Unmarshal(payload, &v); err != nil {
					return nil, fmt.Errorf("failed to unmarshal InitialSubscription: %w", err)
				}
				return &ServerMessage{
					Type:    ServerMessageTypeInitialSubscription,
					Payload: &v,
				}, nil
			case "TransactionUpdate":
				var v TransactionUpdate
				if err := json.Unmarshal(payload, &v); err != nil {
					return nil, fmt.Errorf("failed to unmarshal TransactionUpdate: %w", err)
				}
				return &ServerMessage{
					Type:    ServerMessageTypeTransactionUpdate,
					Payload: &v,
				}, nil
			case "TransactionUpdateLight":
				var v TransactionUpdateLight
				if err := json.Unmarshal(payload, &v); err != nil {
					return nil, fmt.Errorf("failed to unmarshal TransactionUpdateLight: %w", err)
				}
				return &ServerMessage{
					Type:    ServerMessageTypeTransactionUpdateLight,
					Payload: &v,
				}, nil
			case "IdentityToken":
				var v IdentityToken
				if err := json.Unmarshal(payload, &v); err != nil {
					return nil, fmt.Errorf("failed to unmarshal IdentityToken: %w", err)
				}
				return &ServerMessage{
					Type:    ServerMessageTypeIdentityToken,
					Payload: &v,
				}, nil
			case "OneOffQueryResponse":
				var v OneOffQueryResponse
				if err := json.Unmarshal(payload, &v); err != nil {
					return nil, fmt.Errorf("failed to unmarshal OneOffQueryResponse: %w", err)
				}
				return &ServerMessage{
					Type:    ServerMessageTypeOneOffQueryResponse,
					Payload: &v,
				}, nil
			case "SubscribeApplied":
				var v SubscribeApplied
				if err := json.Unmarshal(payload, &v); err != nil {
					return nil, fmt.Errorf("failed to unmarshal SubscribeApplied: %w", err)
				}
				return &ServerMessage{
					Type:    ServerMessageTypeSubscribeApplied,
					Payload: &v,
				}, nil
			case "UnsubscribeApplied":
				var v UnsubscribeApplied
				if err := json.Unmarshal(payload, &v); err != nil {
					return nil, fmt.Errorf("failed to unmarshal UnsubscribeApplied: %w", err)
				}
				return &ServerMessage{
					Type:    ServerMessageTypeUnsubscribeApplied,
					Payload: &v,
				}, nil
			case "SubscriptionError":
				var v SubscriptionError
				if err := json.Unmarshal(payload, &v); err != nil {
					return nil, fmt.Errorf("failed to unmarshal SubscriptionError: %w", err)
				}
				return &ServerMessage{
					Type:    ServerMessageTypeSubscriptionError,
					Payload: &v,
				}, nil
			case "SubscribeMultiApplied":
				var v SubscribeMultiApplied
				if err := json.Unmarshal(payload, &v); err != nil {
					return nil, fmt.Errorf("failed to unmarshal SubscribeMultiApplied: %w", err)
				}
				return &ServerMessage{
					Type:    ServerMessageTypeSubscribeMultiApplied,
					Payload: &v,
				}, nil
			case "UnsubscribeMultiApplied":
				var v UnsubscribeMultiApplied
				if err := json.Unmarshal(payload, &v); err != nil {
					return nil, fmt.Errorf("failed to unmarshal UnsubscribeMultiApplied: %w", err)
				}
				return &ServerMessage{
					Type:    ServerMessageTypeUnsubscribeMultiApplied,
					Payload: &v,
				}, nil
			default:
				return nil, fmt.Errorf("unknown message type: %s", msgType)
			}
		}
	}

	return nil, fmt.Errorf("failed to parse server message")
}
