package agentfiles

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func both(root string) Options {
	return Options{Root: root, Claude: true, VSCode: true, Agent: true}
}

func statuses(t *testing.T, o Options) map[string]Status {
	t.Helper()
	files, err := Plan(o)
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]Status{}
	for _, f := range files {
		out[f.Path] = f.Status
	}
	return out
}

func TestWritingMakesEveryFileCurrentAndTheNextRunChangesNothing(t *testing.T) {
	root := t.TempDir()
	files, _ := Plan(both(root))
	written, err := Write(root, files)
	if err != nil || len(written) != 3 {
		t.Fatal(written, err)
	}
	for path, s := range statuses(t, both(root)) {
		if s != Current {
			t.Errorf("%s is %v", path, s)
		}
	}
	files, _ = Plan(both(root))
	if written, _ := Write(root, files); len(written) != 0 {
		t.Errorf("a second run wrote %v", written)
	}
}

func TestTheMCPFileKeepsTheOtherServersOnDisk(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".mcp.json"), []byte(`{"mcpServers":{"mine":{"command":"mine"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	files, _ := Plan(both(root))
	if _, err := Write(root, files); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(filepath.Join(root, ".mcp.json"))
	if !bytes.Contains(raw, []byte(`"mine"`)) || !bytes.Contains(raw, []byte(`"smith"`)) {
		t.Fatalf("%s", raw)
	}
}

func TestAnAgentFileTheUserEditedIsNeverWritten(t *testing.T) {
	root := t.TempDir()
	files, _ := Plan(both(root))
	_, _ = Write(root, files)
	p := filepath.Join(root, filepath.FromSlash(AgentPath))
	raw, _ := os.ReadFile(p)
	edited := strings.Replace(string(raw), "Habits", "My habits", 1)
	_ = os.WriteFile(p, []byte(edited), 0o644)
	if s := statuses(t, both(root))[AgentPath]; s != Edited {
		t.Fatalf("status %v", s)
	}
	files, _ = Plan(both(root))
	_, _ = Write(root, files)
	if after, _ := os.ReadFile(p); string(after) != edited {
		t.Fatal("the edit was overwritten")
	}
	// Taking the marker out is the way to own the file.
	owned := strings.Replace(edited, markerPrefix, "<!-- mine ", 1)
	_ = os.WriteFile(p, []byte(owned), 0o644)
	if s := statuses(t, both(root))[AgentPath]; s != Edited {
		t.Fatalf("a file without the marker is %v", s)
	}
}

func TestAnOlderUneditedAgentFileIsStaleAndRewritten(t *testing.T) {
	root := t.TempDir()
	old := RenderAgent([]Tool{{Name: "smith_projects", Summary: "List projects", ReadOnly: true}})
	p := filepath.Join(root, filepath.FromSlash(AgentPath))
	_ = os.MkdirAll(filepath.Dir(p), 0o755)
	_ = os.WriteFile(p, []byte(old), 0o644)
	if s := statuses(t, both(root))[AgentPath]; s != Stale {
		t.Fatalf("status %v", s)
	}
	changed, kept, err := Refresh(root, true)
	if err != nil || kept || len(changed) != 1 {
		t.Fatal(changed, kept, err)
	}
	if raw, _ := os.ReadFile(p); string(raw) != old {
		t.Fatal("--dry-run style check wrote")
	}
	if changed, _, _ = Refresh(root, false); len(changed) != 1 {
		t.Fatal(changed)
	}
	if s := statuses(t, both(root))[AgentPath]; s != Current {
		t.Fatalf("status %v", s)
	}
}

func TestRefreshCreatesNothing(t *testing.T) {
	root := t.TempDir()
	if changed, _, err := Refresh(root, false); err != nil || len(changed) != 0 {
		t.Fatal(changed, err)
	}
	if _, err := os.Stat(filepath.Join(root, ".claude")); err == nil {
		t.Fatal("upgrade must offer the agent, not add it")
	}
}

func TestDoctorWarnsOfMissingFilesWithASafeFixAndPassesOnceInstalled(t *testing.T) {
	root := t.TempDir()
	f := Finding(root)
	if f.OK || !f.Warning || f.Fix == nil || !f.Fix.Safe || len(f.Fix.Files) != 3 {
		t.Fatalf("%+v", f)
	}
	files, _ := Plan(both(root))
	_, _ = Write(root, files)
	if f := Finding(root); !f.OK {
		t.Fatalf("%+v", f)
	}
}

func TestClientsFollowTheWorkspace(t *testing.T) {
	root := t.TempDir()
	if c, v := PresentClients(root, nil); !c || !v {
		t.Error("with nothing present both get files")
	}
	_ = os.Mkdir(filepath.Join(root, ".vscode"), 0o755)
	if c, v := PresentClients(root, nil); c || !v {
		t.Errorf("only .vscode: %v %v", c, v)
	}
}

func TestLauncherIsNpxOnlyForASmithInTheWorkspacesNodeModules(t *testing.T) {
	root := t.TempDir()
	if got := LauncherFor(root, filepath.Join(root, "node_modules", "monosmith", "bin", "smith")); got.Command != "npx" {
		t.Errorf("%v", got)
	}
	if got := LauncherFor(root, "/usr/local/bin/smith"); got.Command != "smith" {
		t.Errorf("%v", got)
	}
}
