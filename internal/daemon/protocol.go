package daemon

import (
	"encoding/json"
	"fmt"
)

// ProtocolVersion is the current IPC protocol version
const ProtocolVersion = 1

// MessageType represents the type of IPC message
type MessageType string

const (
	MessageTypeStatus      MessageType = "Status"
	MessageTypeQueryExecute MessageType = "QueryExecute"
	MessageTypeReindexPaths MessageType = "ReindexPaths"
	MessageTypeShutdown    MessageType = "Shutdown"
)

// Message represents an IPC message
type Message struct {
	Version int         `json:"version"`
	Type    MessageType `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// Response represents an IPC response
type Response struct {
	Version int         `json:"version"`
	Success bool        `json:"success"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// StatusRequest represents a status request
type StatusRequest struct{}

// StatusResponse represents a status response
type StatusResponse struct {
	Running bool `json:"running"`
	PID     int  `json:"pid,omitempty"`
}

// QueryExecuteRequest represents a query execution request
type QueryExecuteRequest struct {
	Query string `json:"query"`
}

// QueryExecuteResponse represents a query execution response
type QueryExecuteResponse struct {
	Results interface{} `json:"results"`
}

// ReindexPathsRequest represents a reindex paths request
type ReindexPathsRequest struct {
	Paths []string `json:"paths"`
}

// ReindexPathsResponse represents a reindex paths response
type ReindexPathsResponse struct {
	Processed int `json:"processed"`
}

// NewMessage creates a new message
func NewMessage(msgType MessageType, payload interface{}) (*Message, error) {
	var payloadJSON json.RawMessage
	var err error

	if payload != nil {
		payloadJSON, err = json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("marshaling payload: %w", err)
		}
	}

	return &Message{
		Version: ProtocolVersion,
		Type:    msgType,
		Payload: payloadJSON,
	}, nil
}

// NewResponse creates a new response
func NewResponse(success bool, data interface{}, err error) *Response {
	resp := &Response{
		Version: ProtocolVersion,
		Success: success,
		Data:    data,
	}

	if err != nil {
		resp.Error = err.Error()
	}

	return resp
}
