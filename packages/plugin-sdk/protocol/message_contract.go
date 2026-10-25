package protocol

import "encoding/json"

// Version is the protocol major version this SDK speaks. A breaking change
// raises it; additive changes are signalled through capabilities instead.
const Version = 1

// Request is a JSON-RPC 2.0 request, or a notification when ID is nil.
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *int64          `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Response is a JSON-RPC 2.0 response. Exactly one of Result and Error is set.
type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *int64          `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *ResponseError  `json:"error,omitempty"`
}

// ResponseError is a JSON-RPC 2.0 error object.
type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Error returns the message so a ResponseError can be used as a Go error.
func (e *ResponseError) Error() string { return e.Message }

// Standard JSON-RPC 2.0 error codes used by the protocol.
const (
	CodeParseError     = -32700
	CodeInvalidRequest = -32600
	CodeMethodNotFound = -32601
	CodeInvalidParams  = -32602
	CodeInternalError  = -32603
	// CodePluginFailure is returned when a method ran but failed for a
	// reason the user must act on (a missing toolchain, a bad manifest).
	CodePluginFailure = -32000
	// CodeRequestCancelled answers a request the plugin stopped because the
	// host sent a cancel notification for its id. A plugin that finished the
	// work before it saw the cancel answers with the result instead.
	CodeRequestCancelled = -32800
)
