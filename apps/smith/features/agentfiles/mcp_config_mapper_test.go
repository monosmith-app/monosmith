package agentfiles

import (
	"errors"
	"strings"
	"testing"
)

const claudeGolden = `{
  "mcpServers": {
    "smith": {
      "command": "smith",
      "args": [
        "mcp"
      ]
    }
  }
}
`

const vscodeGolden = `{
  "servers": {
    "smith": {
      "type": "stdio",
      "command": "smith",
      "args": [
        "mcp"
      ]
    }
  }
}
`

func TestAnEmptyFileGetsTheServer(t *testing.T) {
	for _, c := range []struct {
		form MCPForm
		want string
	}{{ClaudeForm, claudeGolden}, {VSCodeForm, vscodeGolden}} {
		got, state, err := MergeMCP(nil, c.form, DirectLauncher)
		if err != nil || state != MCPMissing || string(got) != c.want {
			t.Errorf("%s: %v %v\n%s", c.form.Path, state, err, got)
		}
	}
}

func TestMergingKeepsOtherServersAndKeysInOrder(t *testing.T) {
	existing := `{"zeta": 1, "mcpServers": {"b": {"command": "b"}, "a": {"url": "https://x"}}, "alpha": [1,2]}`
	got, state, err := MergeMCP([]byte(existing), ClaudeForm, DirectLauncher)
	if err != nil || state != MCPMissing {
		t.Fatal(state, err)
	}
	text := string(got)
	order := []string{`"zeta"`, `"mcpServers"`, `"b"`, `"a"`, `"smith"`, `"alpha"`}
	last := -1
	for _, k := range order {
		i := strings.Index(text, k)
		if i < last {
			t.Fatalf("%s is out of order in\n%s", k, text)
		}
		last = i
	}
	again, state, err := MergeMCP(got, ClaudeForm, DirectLauncher)
	if err != nil || state != MCPCurrent || string(again) != text {
		t.Fatalf("merging twice must change nothing: %v %v", state, err)
	}
}

func TestEitherLauncherIsOneSmithWrites(t *testing.T) {
	npx := `{"mcpServers": {"smith": {"command": "npx", "args": ["smith", "mcp"]}}}`
	got, state, _ := MergeMCP([]byte(npx), ClaudeForm, DirectLauncher)
	if state != MCPCurrent || string(got) != npx {
		t.Fatalf("%v %s", state, got)
	}
}

func TestAnEditedEntryIsReportedAndShownWithoutBeingWritten(t *testing.T) {
	edited := `{"mcpServers": {"smith": {"command": "/opt/smith", "args": ["mcp", "--verbose"]}}}`
	got, state, err := MergeMCP([]byte(edited), ClaudeForm, DirectLauncher)
	if err != nil || state != MCPEdited {
		t.Fatal(state, err)
	}
	if !strings.Contains(string(got), `"smith"`) || strings.Contains(string(got), "/opt/smith") {
		t.Fatalf("the returned file is what smith would make:\n%s", got)
	}
}

func TestFilesSmithCannotMergeIntoAreRefused(t *testing.T) {
	for _, bad := range []string{"{ // comment\n}", "[]", "not json", `{"mcpServers": []}`} {
		if _, _, err := MergeMCP([]byte(bad), ClaudeForm, DirectLauncher); !errors.Is(err, ErrNotJSON) && err == nil {
			t.Errorf("%q was accepted", bad)
		}
	}
}
