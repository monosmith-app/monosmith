package derivefile

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func referencesFile() protocol.DerivedFile {
	return protocol.DerivedFile{
		Path: "a/tsconfig.lib.json", Mode: protocol.DeriveKeys, Format: "json", Renderer: "r", Version: 1,
		Keys: map[string]any{"references": []any{map[string]any{"path": "../b"}}},
	}
}

func TestWantedKeysArePrettyPrintedSoADiffReadsLineByLine(t *testing.T) {
	got, err := Wanted(referencesFile())
	if err != nil {
		t.Fatal(err)
	}
	if want := "{\n  \"references\": [\n    {\n      \"path\": \"../b\"\n    }\n  ]\n}\n"; string(got) != want {
		t.Fatalf("got %q", got)
	}
}

func TestMergeKeepsThePositionOfAKeyItChanges(t *testing.T) {
	disk := "{\n  \"extends\": \"x\",\n  \"references\": [],\n  \"exclude\": [\"y\"]\n}\n"
	got, err := Merge(referencesFile(), []byte(disk), true)
	if err != nil {
		t.Fatal(err)
	}
	s := string(got)
	if !(strings.Index(s, `"extends"`) < strings.Index(s, `"references"`) && strings.Index(s, `"references"`) < strings.Index(s, `"exclude"`)) {
		t.Fatalf("the key moved:\n%s", s)
	}
}
