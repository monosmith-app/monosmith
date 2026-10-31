package languageserver

import (
	"encoding/json"
	"net/url"
	"path"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// JSON-RPC error codes.
const (
	codeParseError     = -32700
	codeMethodNotFound = -32601
	codeInvalidParams  = -32602
)

// Server answers LSP messages for smith.toml documents. It holds the open
// documents and handles one message at a time.
type Server struct {
	version  string
	docs     *documentStore
	shutdown bool
	exited   bool
}

// NewServer returns a server that reports version as its own in `initialize`.
func NewServer(version string) *Server {
	return &Server{version: version, docs: newDocumentStore()}
}

// Exited reports whether the client sent the `exit` notification.
func (s *Server) Exited() bool { return s.exited }

// CleanExit reports whether `shutdown` came before `exit`, as the protocol
// requires for exit code 0.
func (s *Server) CleanExit() bool { return s.shutdown }

// Handle processes one message body and returns the bodies to send back: the
// answer to a request, and any notification (diagnostics) it caused.
func (s *Server) Handle(body []byte) [][]byte {
	var req request
	if err := json.Unmarshal(body, &req); err != nil {
		return encode(errorResponse{JSONRPC: "2.0", ID: json.RawMessage("null"), Error: responseError{Code: codeParseError, Message: err.Error()}})
	}
	isRequest := len(req.ID) > 0
	var params textDocumentParams
	if len(req.Params) > 0 {
		if err := json.Unmarshal(req.Params, &params); err != nil {
			if isRequest {
				return encode(errorResponse{JSONRPC: "2.0", ID: req.ID, Error: responseError{Code: codeInvalidParams, Message: err.Error()}})
			}
			return nil
		}
	}
	uri := params.TextDocument.URI

	switch req.Method {
	case "initialize":
		return encode(response{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{
			"capabilities": map[string]any{
				"textDocumentSync":   1,
				"completionProvider": map[string]any{"triggerCharacters": []string{"[", "."}},
				"hoverProvider":      true,
			},
			"serverInfo": map[string]any{"name": "smith", "version": s.version},
		}})
	case "shutdown":
		s.shutdown = true
		return encode(response{JSONRPC: "2.0", ID: req.ID})
	case "exit":
		s.exited = true
		return nil
	case "textDocument/didOpen":
		if !supported(uri) {
			return nil
		}
		s.docs.set(uri, params.TextDocument.Text)
		return s.publish(uri)
	case "textDocument/didChange":
		if !supported(uri) || len(params.ContentChanges) == 0 {
			return nil
		}
		// The server asks for full-text sync, so the last change is the text.
		s.docs.set(uri, params.ContentChanges[len(params.ContentChanges)-1].Text)
		return s.publish(uri)
	case "textDocument/didClose":
		if !supported(uri) {
			return nil
		}
		s.docs.remove(uri)
		return encode(notification{JSONRPC: "2.0", Method: "textDocument/publishDiagnostics", Params: map[string]any{"uri": uri, "diagnostics": []Diagnostic{}}})
	case "textDocument/completion":
		items := []CompletionItem{}
		if text, ok := s.docs.get(uri); ok {
			items = Complete(text, params.Position)
		}
		return encode(response{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{"isIncomplete": false, "items": items}})
	case "textDocument/hover":
		var hover *Hover
		if text, ok := s.docs.get(uri); ok {
			hover = HoverAt(text, params.Position)
		}
		if hover == nil {
			return encode(response{JSONRPC: "2.0", ID: req.ID})
		}
		return encode(response{JSONRPC: "2.0", ID: req.ID, Result: hover})
	}
	if isRequest && !strings.HasPrefix(req.Method, "$/") {
		return encode(errorResponse{JSONRPC: "2.0", ID: req.ID, Error: responseError{Code: codeMethodNotFound, Message: "method not supported: " + req.Method}})
	}
	return nil
}

// publish diagnoses the stored text of uri and returns the notification.
func (s *Server) publish(uri string) [][]byte {
	text, _ := s.docs.get(uri)
	diags := Diagnose(text)
	if diags == nil {
		diags = []Diagnostic{}
	}
	return encode(notification{JSONRPC: "2.0", Method: "textDocument/publishDiagnostics", Params: map[string]any{"uri": uri, "diagnostics": diags}})
}

// supported reports whether uri names a smith.toml file; other documents are
// ignored until the server learns recipe files.
func supported(uri string) bool {
	u, err := url.Parse(uri)
	if err != nil {
		return false
	}
	return path.Base(u.Path) == workspace.FileName
}

// encode marshals one outgoing message into a one-element list.
func encode(v any) [][]byte {
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return [][]byte{b}
}
