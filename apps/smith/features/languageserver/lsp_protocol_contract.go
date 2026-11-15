package languageserver

import "encoding/json"

// Position is a zero-based line and character in a document. Characters count
// runes, which equals the protocol's UTF-16 units for text in the Basic
// Multilingual Plane.
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// Range is a span of a document, from Start up to End.
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Diagnostic is one problem the editor underlines.
type Diagnostic struct {
	Range    Range  `json:"range"`
	Severity int    `json:"severity"`
	Source   string `json:"source"`
	Message  string `json:"message"`
}

// CompletionItem is one suggestion in the completion list.
type CompletionItem struct {
	Label         string `json:"label"`
	Kind          int    `json:"kind"`
	Detail        string `json:"detail,omitempty"`
	Documentation string `json:"documentation,omitempty"`
	InsertText    string `json:"insertText,omitempty"`
}

// MarkupContent is formatted text, here Markdown.
type MarkupContent struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

// Hover is the answer to textDocument/hover.
type Hover struct {
	Contents MarkupContent `json:"contents"`
}

// Severity and completion kind numbers from the protocol.
const (
	// SeverityError marks a problem smith itself rejects.
	SeverityError = 1
	// CompletionKindModule is the kind of a table name.
	CompletionKindModule = 9
	// CompletionKindProperty is the kind of a key.
	CompletionKindProperty = 10
	// CompletionKindValue is the kind of an enumerated value.
	CompletionKindValue = 12
)

// request is an incoming JSON-RPC message: a request when ID is present, a
// notification when it is not.
type request struct {
	ID     json.RawMessage `json:"id,omitempty"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params,omitempty"`
}

// response is an outgoing JSON-RPC answer; a nil Result encodes as null.
type response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result"`
}

// errorResponse is an outgoing JSON-RPC error, which carries no result.
type errorResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Error   responseError   `json:"error"`
}

// responseError is the error member of an errorResponse.
type responseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// notification is an outgoing JSON-RPC notification.
type notification struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  any    `json:"params"`
}

// textDocumentParams covers the params of didOpen, didChange, didClose,
// completion and hover that this server reads.
type textDocumentParams struct {
	TextDocument struct {
		URI  string `json:"uri"`
		Text string `json:"text"`
	} `json:"textDocument"`
	ContentChanges []struct {
		Text string `json:"text"`
	} `json:"contentChanges"`
	Position Position `json:"position"`
}
