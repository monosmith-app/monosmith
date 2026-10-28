package agentfiles

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func workspaceDir(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "smith.toml"), []byte("schema = 1\n[workspace]\nname = \"x\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SMITH_WORKSPACE_ROOT", root)
	return root
}

func install(in string, yes, interactive bool) (string, error) {
	var out bytes.Buffer
	err := Install{Agent: true, Claude: true, VSCode: true, Yes: yes, In: strings.NewReader(in), Out: &out, Interactive: func() bool { return interactive }}.Run()
	return out.String(), err
}

func TestInstallWithoutATerminalAsksForYes(t *testing.T) {
	root := workspaceDir(t)
	out, err := install("", false, false)
	if err == nil || !strings.Contains(err.Error(), "--yes") && !strings.Contains(err.Error(), "no terminal") {
		t.Fatalf("%v", err)
	}
	if !strings.Contains(out, AgentPath) {
		t.Errorf("the files are previewed:\n%s", out)
	}
	if _, err := os.Stat(filepath.Join(root, ".mcp.json")); err == nil {
		t.Fatal("wrote without asking")
	}
}

func TestInstallAsksOnATerminalAndYesWrites(t *testing.T) {
	root := workspaceDir(t)
	if out, err := install("n\n", false, true); err != nil || !strings.Contains(out, "Nothing written") {
		t.Fatalf("%v\n%s", err, out)
	}
	out, err := install("y\n", false, true)
	if err != nil || !strings.Contains(out, "Wrote .claude/agents/smith.md") {
		t.Fatalf("%v\n%s", err, out)
	}
	for _, p := range []string{AgentPath, ClaudeMCPPath, VSCodeMCPPath} {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(p))); err != nil {
			t.Error(err)
		}
	}
	if out, err := install("", true, false); err != nil || !strings.Contains(out, "already installed") {
		t.Fatalf("idempotent: %v\n%s", err, out)
	}
}

func TestInstallStopsAndShowsTheDiffWhenAFileWasEdited(t *testing.T) {
	root := workspaceDir(t)
	if _, err := install("", true, false); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(root, filepath.FromSlash(VSCodeMCPPath))
	_ = os.WriteFile(p, []byte(`{"servers":{"smith":{"command":"/mine/smith"}}}`), 0o644)
	out, err := install("", true, false)
	if err == nil || !strings.Contains(out, "/mine/smith") {
		t.Fatalf("%v\n%s", err, out)
	}
	if raw, _ := os.ReadFile(p); !strings.Contains(string(raw), "/mine/smith") {
		t.Fatal("overwritten")
	}
}

func TestMCPOnlyLeavesTheAgentFileOut(t *testing.T) {
	root := workspaceDir(t)
	var out bytes.Buffer
	err := Install{Claude: true, VSCode: true, Yes: true, In: strings.NewReader(""), Out: &out}.Run()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, ".claude")); err == nil {
		t.Fatal("smith mcp install wrote the agent")
	}
}
