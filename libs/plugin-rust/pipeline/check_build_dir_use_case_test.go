package pipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func rustWorkspace(t *testing.T, files map[string]string) string {
	t.Helper()
	t.Setenv("CARGO_TARGET_DIR", "")
	ws := t.TempDir()
	for rel, body := range files {
		p := filepath.Join(ws, filepath.FromSlash(rel))
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return ws
}

var kitCrate = []protocol.Project{{Name: "kit", Root: "crates/kit"}, {Name: "cli", Root: "apps/cli"}}

func TestBuildDirReportsACrateThatBuildsIntoTargetAndOffersTheMove(t *testing.T) {
	ws := rustWorkspace(t, map[string]string{
		"Cargo.toml":            "[workspace]\nmembers = [\"crates/kit\", \"apps/cli\"]\n",
		"crates/kit/Cargo.toml": "[package]\nname = \"kit\"\n",
		".cargo/config.toml":    "[alias]\nb = \"build\"\n",
		"apps/cli/Cargo.toml":   "[package]\nname = \"cli\"\n",
	})
	findings, ok := CheckBuildDir(ws, kitCrate)
	if !ok || len(findings) != 1 {
		t.Fatalf("one finding for the workspace: %+v", findings)
	}
	f := findings[0]
	if f.OK || f.Warning || !strings.Contains(f.Detail, "kit, cli build into target/") || !strings.Contains(f.Detail, "no .cargo/config.toml sets target-dir") {
		t.Fatalf("names where it builds now: %+v", f)
	}
	if f.Fix == nil || len(f.Fix.Choices) != 1 || len(f.Fix.Choices[0].Options) != 2 {
		t.Fatalf("a question, move or leave: %+v", f.Fix)
	}
	move, leave := f.Fix.Choices[0].Options[0].Fix, f.Fix.Choices[0].Options[1].Fix
	if len(leave.Files) != 0 || len(leave.Steps) != 0 || len(move.Files) != 1 || move.Files[0].Path != ".cargo/config.toml" {
		t.Fatalf("leave changes nothing, move writes the config: %+v", f.Fix)
	}
	if got := move.Files[0].Content; got != "[alias]\nb = \"build\"\n\n[build]\ntarget-dir = \"dist/cargo\"\n" {
		t.Fatalf("the other lines stay: %q", got)
	}
}

func TestBuildDirPassesWhenTargetDirIsInDist(t *testing.T) {
	ws := rustWorkspace(t, map[string]string{
		"Cargo.toml":            "[workspace]\nmembers = [\"crates/kit\"]\n",
		"crates/kit/Cargo.toml": "[package]\nname = \"kit\"\n",
		".cargo/config.toml":    "[build]\ntarget-dir = \"dist/cargo\"\n",
	})
	findings, _ := CheckBuildDir(ws, kitCrate[:1])
	if len(findings) != 1 || !findings[0].OK {
		t.Fatalf("%+v", findings)
	}
}

func TestBuildDirUpdatesAnotherTargetDirAndKeepsTheComment(t *testing.T) {
	ws := rustWorkspace(t, map[string]string{
		"Cargo.toml":            "[workspace]\nmembers = [\"crates/kit\"]\n",
		"crates/kit/Cargo.toml": "[package]\nname = \"kit\"\n",
		".cargo/config.toml":    "# mine\n[build]\ntarget-dir = \"out/cargo\" # shared\njobs = 4\n",
	})
	f, _ := CheckBuildDir(ws, kitCrate[:1])
	if !strings.Contains(f[0].Detail, "builds into out/cargo/") || f[0].Fix == nil {
		t.Fatalf("%+v", f[0])
	}
	if got := f[0].Fix.Choices[0].Options[0].Fix.Files[0].Content; got != "# mine\n[build]\ntarget-dir = \"dist/cargo\" # shared\njobs = 4\n" {
		t.Fatalf("%q", got)
	}
}

func TestBuildDirOfACrateWithItsOwnWorkspaceIsItsOwnFindingAndWritesBesideIt(t *testing.T) {
	ws := rustWorkspace(t, map[string]string{
		"Cargo.toml":             "[workspace]\nmembers = []\n",
		"tools/probe/Cargo.toml": "[package]\nname = \"probe\"\n\n[workspace]\n",
	})
	f, _ := CheckBuildDir(ws, []protocol.Project{{Name: "probe", Root: "tools/probe"}})
	if len(f) != 1 || f[0].Check != "probe builds into dist/cargo" || f[0].Fix == nil {
		t.Fatalf("%+v", f)
	}
	file := f[0].Fix.Choices[0].Options[0].Fix.Files[0]
	if file.Path != "tools/probe/.cargo/config.toml" || !strings.Contains(file.Content, `target-dir = "../../dist/cargo"`) {
		t.Fatalf("relative to the folder that holds .cargo: %+v", file)
	}
}

func TestBuildDirRefusesWithAWarningWhenAFileNamesTheOldFolder(t *testing.T) {
	ws := rustWorkspace(t, map[string]string{
		"Cargo.toml":            "[workspace]\nmembers = [\"crates/kit\"]\n",
		"crates/kit/Cargo.toml": "[package]\nname = \"kit\"\n",
		"Dockerfile":            "COPY target/release/kit /usr/bin/kit\n",
	})
	f, _ := CheckBuildDir(ws, kitCrate[:1])
	if f[0].Fix != nil || !f[0].Warning || f[0].OK || !strings.Contains(f[0].Remedy, "Dockerfile names target/") {
		t.Fatalf("a warning with the file named and no fix: %+v", f[0])
	}
}

func TestBuildDirTheEnvironmentWinsAndIsNeverEdited(t *testing.T) {
	ws := rustWorkspace(t, map[string]string{
		"Cargo.toml":            "[workspace]\nmembers = [\"crates/kit\"]\n",
		"crates/kit/Cargo.toml": "[package]\nname = \"kit\"\n",
		".cargo/config.toml":    "[build]\ntarget-dir = \"dist/cargo\"\n",
	})
	t.Setenv("CARGO_TARGET_DIR", filepath.Join(ws, "..", "elsewhere"))
	f, _ := CheckBuildDir(ws, kitCrate[:1])
	if f[0].OK || !f[0].Warning || f[0].Fix != nil || !strings.Contains(f[0].Detail, "CARGO_TARGET_DIR") {
		t.Fatalf("%+v", f[0])
	}
}

func TestBuildDirLeavesAnUnreadableTargetDirUnchecked(t *testing.T) {
	ws := rustWorkspace(t, map[string]string{
		"Cargo.toml":            "[workspace]\nmembers = [\"crates/kit\"]\n",
		"crates/kit/Cargo.toml": "[package]\nname = \"kit\"\n",
		".cargo/config.toml":    "build = { target-dir = \"x\" }\n",
	})
	if f, ok := CheckBuildDir(ws, kitCrate[:1]); ok {
		t.Fatalf("nothing is guessed: %+v", f)
	}
}
