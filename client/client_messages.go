package client

// ClientMessage represents all possible client-to-server messages
type ClientMessage struct {
	CallReducer      *CallReducer      `json:"CallReducer,omitempty"`
	Subscribe        *Subscribe        `json:"Subscribe,omitempty"`
	OneOffQuery      *OneOffQuery      `json:"OneOffQuery,omitempty"`
	SubscribeSingle  *SubscribeSingle  `json:"SubscribeSingle,omitempty"`
	SubscribeMulti   *SubscribeMulti   `json:"SubscribeMulti,omitempty"`
	Unsubscribe      *Unsubscribe      `json:"Unsubscribe,omitempty"`
	UnsubscribeMulti *UnsubscribeMulti `json:"UnsubscribeMulti,omitempty"`
}

// CallReducer represents a reducer call request
type CallReducer struct {
	Reducer   string `json:"reducer"`
	Args      string `json:"args"`
	RequestID uint32 `json:"request_id"`
	Flags     uint8  `json:"flags"`
}

// Subscribe represents a subscription request
type Subscribe struct {
	QueryStrings []string `json:"query_strings"`
	RequestID    uint32   `json:"request_id"`
}

// OneOffQuery represents a one-off query request
type OneOffQuery struct {
	MessageID   []byte `json:"message_id"`
	QueryString string `json:"query_string"`
}

// SubscribeSingle represents a single subscription request
type SubscribeSingle struct {
	Query     string  `json:"query"`
	RequestID uint32  `json:"request_id"`
	QueryID   QueryID `json:"query_id"`
}

// SubscribeMulti represents a multi-subscription request
type SubscribeMulti struct {
	QueryStrings []string `json:"query_strings"`
	RequestID    uint32   `json:"request_id"`
	QueryID      QueryID  `json:"query_id"`
}

// Unsubscribe represents an unsubscription request
type Unsubscribe struct {
	RequestID uint32  `json:"request_id"`
	QueryID   QueryID `json:"query_id"`
}

// UnsubscribeMulti represents a multi-unsubscription request
type UnsubscribeMulti struct {
	RequestID uint32  `json:"request_id"`
	QueryID   QueryID `json:"query_id"`
}

// QueryID represents a query identifier
type QueryID struct {
	ID uint32 `json:"id"`
}
