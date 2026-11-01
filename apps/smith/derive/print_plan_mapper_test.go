package derive

import (
	"bytes"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestExplainShowsAKeysDiffOneLinePerLine(t *testing.T) {
	f := protocol.DerivedFile{
		Path: "a/tsconfig.lib.json", Mode: protocol.DeriveKeys, Format: "json", Renderer: "r", Version: 1,
		Keys: map[string]any{"references": []any{map[string]any{"path": "../b"}, map[string]any{"path": "../c"}}},
	}
	item := Classify(f, nil, []byte(`{"references": [{"path": "../b"}]}`), true)
	var out bytes.Buffer
	PrintExplanation(&out, []Item{item})
	for _, want := range []string{"diverged", "+      \"path\": \"../c\""} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("missing %q in:\n%s", want, out.String())
		}
	}
	for _, line := range strings.Split(out.String(), "\n") {
		if len(line) > 120 {
			t.Fatalf("a minified line in the diff: %q", line)
		}
	}
}
