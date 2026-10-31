package languageserver

import (
	"encoding/json"
	"strings"
	"testing"
)

const tomlURI = "file:///w/smith.toml"

func send(t *testing.T, s *Server, msg string) []map[string]any {
	t.Helper()
	var out []map[string]any
	for _, b := range s.Handle([]byte(msg)) {
		var m map[string]any
		if err := json.Unmarshal(b, &m); err != nil {
			t.Fatalf("bad output %s: %v", b, err)
		}
		out = append(out, m)
	}
	return out
}

func TestInitializeAdvertisesCapabilities(t *testing.T) {
	out := send(t, NewServer("1.2.3"), `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`)
	if len(out) != 1 {
		t.Fatalf("got %v", out)
	}
	res := out[0]["result"].(map[string]any)
	caps := res["capabilities"].(map[string]any)
	if caps["hoverProvider"] != true || caps["textDocumentSync"] != float64(1) || caps["completionProvider"] == nil {
		t.Errorf("capabilities %v", caps)
	}
	if res["serverInfo"].(map[string]any)["version"] != "1.2.3" {
		t.Errorf("serverInfo %v", res["serverInfo"])
	}
}

func TestDidOpenAndDidChangePublishDiagnostics(t *testing.T) {
	s := NewServer("x")
	out := send(t, s, `{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"`+tomlURI+`","text":"[workspace]\nnme = \"a\"\n"}}}`)
	if len(out) != 1 || out[0]["method"] != "textDocument/publishDiagnostics" {
		t.Fatalf("got %v", out)
	}
	diags := out[0]["params"].(map[string]any)["diagnostics"].([]any)
	if len(diags) != 1 {
		t.Fatalf("diagnostics %v", diags)
	}
	out = send(t, s, `{"jsonrpc":"2.0","method":"textDocument/didChange","params":{"textDocument":{"uri":"`+tomlURI+`"},"contentChanges":[{"text":"[workspace]\nname = \"a\"\n"}]}}`)
	if d := out[0]["params"].(map[string]any)["diagnostics"].([]any); len(d) != 0 {
		t.Errorf("after the fix: %v", d)
	}
	out = send(t, s, `{"jsonrpc":"2.0","method":"textDocument/didClose","params":{"textDocument":{"uri":"`+tomlURI+`"}}}`)
	if d := out[0]["params"].(map[string]any)["diagnostics"].([]any); len(d) != 0 {
		t.Errorf("after close: %v", d)
	}
}

func TestOtherFilesAreIgnored(t *testing.T) {
	out := send(t, NewServer("x"), `{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///w/other.toml","text":"x"}}}`)
	if len(out) != 0 {
		t.Fatalf("got %v", out)
	}
}

func TestCompletionAndHoverRequests(t *testing.T) {
	s := NewServer("x")
	send(t, s, `{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"`+tomlURI+`","text":"[workspace]\nsc"}}}`)
	out := send(t, s, `{"jsonrpc":"2.0","id":2,"method":"textDocument/completion","params":{"textDocument":{"uri":"`+tomlURI+`"},"position":{"line":1,"character":2}}}`)
	items := out[0]["result"].(map[string]any)["items"].([]any)
	if len(items) == 0 || items[0].(map[string]any)["label"] != "scope" {
		t.Errorf("completion %v", items)
	}
	out = send(t, s, `{"jsonrpc":"2.0","id":3,"method":"textDocument/hover","params":{"textDocument":{"uri":"`+tomlURI+`"},"position":{"line":0,"character":3}}}`)
	if h, _ := out[0]["result"].(map[string]any); h == nil || !strings.Contains(h["contents"].(map[string]any)["value"].(string), "workspace") {
		t.Errorf("hover %v", out[0])
	}
	out = send(t, s, `{"jsonrpc":"2.0","id":4,"method":"textDocument/hover","params":{"textDocument":{"uri":"`+tomlURI+`"},"position":{"line":5,"character":0}}}`)
	if v, present := out[0]["result"]; !present || v != nil {
		t.Errorf("hover off the text: want result null, got %v", out[0])
	}
}

func TestUnknownRequestGetsAnErrorAndUnknownNotificationNone(t *testing.T) {
	s := NewServer("x")
	out := send(t, s, `{"jsonrpc":"2.0","id":9,"method":"textDocument/rename","params":{}}`)
	if e := out[0]["error"].(map[string]any); e["code"] != float64(codeMethodNotFound) {
		t.Errorf("got %v", out[0])
	}
	if _, has := out[0]["result"]; has {
		t.Error("an error response carries a result")
	}
	if out := send(t, s, `{"jsonrpc":"2.0","method":"workspace/didChangeConfiguration","params":{}}`); len(out) != 0 {
		t.Errorf("got %v", out)
	}
	if out := send(t, s, `{"jsonrpc":"2.0","id":10,"method":"$/cancelRequest","params":{}}`); len(out) != 0 {
		t.Errorf("got %v", out)
	}
}

func TestBadJSONGetsAParseError(t *testing.T) {
	out := send(t, NewServer("x"), `{nope`)
	if out[0]["error"].(map[string]any)["code"] != float64(codeParseError) {
		t.Errorf("got %v", out)
	}
}

func TestShutdownThenExit(t *testing.T) {
	s := NewServer("x")
	out := send(t, s, `{"jsonrpc":"2.0","id":5,"method":"shutdown"}`)
	if v, present := out[0]["result"]; !present || v != nil {
		t.Errorf("shutdown: %v", out[0])
	}
	send(t, s, `{"jsonrpc":"2.0","method":"exit"}`)
	if !s.Exited() || !s.CleanExit() {
		t.Error("want a clean exit")
	}
	s2 := NewServer("x")
	send(t, s2, `{"jsonrpc":"2.0","method":"exit"}`)
	if !s2.Exited() || s2.CleanExit() {
		t.Error("exit without shutdown is not clean")
	}
}
