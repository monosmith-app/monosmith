package languageserver

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
)

func TestCommandServesOverFramedStdio(t *testing.T) {
	var in bytes.Buffer
	for _, m := range []string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`,
		`{"jsonrpc":"2.0","method":"initialized","params":{}}`,
		`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///w/smith.toml","text":"[workspace]\nname = \"a\"\nbad = 1\n"}}}`,
		`{"jsonrpc":"2.0","id":2,"method":"shutdown"}`,
		`{"jsonrpc":"2.0","method":"exit"}`,
	} {
		if err := WriteFrame(&in, []byte(m)); err != nil {
			t.Fatal(err)
		}
	}
	var out bytes.Buffer
	cmd := Command("test")
	cmd.SetIn(&in)
	cmd.SetOut(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	r := bufio.NewReader(&out)
	var bodies []string
	for {
		b, err := ReadFrame(r)
		if err != nil {
			break
		}
		bodies = append(bodies, string(b))
	}
	if len(bodies) != 3 {
		t.Fatalf("want initialize, diagnostics, shutdown; got %v", bodies)
	}
	if !strings.Contains(bodies[1], "unknown key bad") && !strings.Contains(bodies[1], "unknown key workspace.bad") {
		t.Errorf("diagnostics %s", bodies[1])
	}
}

func TestCommandEndsAtEndOfInput(t *testing.T) {
	cmd := Command("test")
	cmd.SetIn(strings.NewReader(""))
	cmd.SetOut(&bytes.Buffer{})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestCommandFailsOnExitBeforeShutdown(t *testing.T) {
	var in bytes.Buffer
	_ = WriteFrame(&in, []byte(`{"jsonrpc":"2.0","method":"exit"}`))
	cmd := Command("test")
	cmd.SetIn(&in)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SilenceUsage, cmd.SilenceErrors = true, true
	if err := cmd.Execute(); err == nil {
		t.Fatal("want an error")
	}
}
