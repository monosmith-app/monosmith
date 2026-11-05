package pipeline

import (
	"os"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

const pinLine = "golangci-lint has version 2.15.0 built with go1.26 from abc on 2026-01-01"

// fakeMachine replaces the PATH search and the probe: onPath are the files the
// PATH holds in order, says maps a file to what it prints, bin is the folder go
// install writes to ("" when go does not say).
func fakeMachine(t *testing.T, onPath []string, says map[string]string, bin string) {
	t.Helper()
	realProbe, realLook := probe, lookAll
	t.Cleanup(func() { probe, lookAll = realProbe, realLook })
	lookAll = func(string) []string { return onPath }
	probe = func(_ string, argv []string) (string, error) {
		if argv[0] == "go" && len(argv) > 1 && argv[1] == "env" {
			if bin == "" {
				return "", os.ErrNotExist
			}
			return bin + "\n" + filepath.Join(bin, "..") + "\n", nil
		}
		if v, ok := says[argv[0]]; ok {
			return v, nil
		}
		return "", os.ErrNotExist
	}
}

// putLint puts a golangci-lint file in a folder (a fresh one when dir is "")
// and returns the folder and the file.
func putLint(t *testing.T, dir string) (string, string) {
	t.Helper()
	if dir == "" {
		dir = t.TempDir()
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(dir, lintProgram+exeSuffix(goruntime.GOOS))
	if err := os.WriteFile(file, []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	return dir, file
}

func TestLintVersionReadsEveryOutputFormat(t *testing.T) {
	for output, want := range map[string]string{
		pinLine: "2.15.0",
		"golangci-lint has version v2.15.0 built with go1.26":       "2.15.0",
		"golangci-lint has version 1.64.8 built with go1.24 from x": "1.64.8",
		"golangci-lint version 2.15.0 (commit abc)":                 "2.15.0",
		"something else entirely":                                   "",
	} {
		if got := lintVersion(output); got != want {
			t.Errorf("lintVersion(%q) = %q, want %q", output, got, want)
		}
	}
}

func TestLintFindingPassesWhenThePathFindsThePin(t *testing.T) {
	fakeMachine(t, []string{"/usr/bin/golangci-lint"}, map[string]string{"/usr/bin/golangci-lint": pinLine}, "")
	f := lintFinding(t.TempDir(), "2.15.0")
	if !f.OK || f.Fix != nil || !strings.Contains(f.Detail, "found /usr/bin/golangci-lint ("+pinLine+")") {
		t.Fatalf("equal: the detail names the file and the version line: %+v", f)
	}
}

func TestLintFindingOffersTheWaysOnForANewerAndAnOlderVersion(t *testing.T) {
	for version, word := range map[string]string{"2.16.0": "newer", "2.13.0": "older", "1.64.8": "older"} {
		root := t.TempDir()
		if err := os.WriteFile(filepath.Join(root, "smith.toml"), []byte("[workspace]\nname = \"w\"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		fakeMachine(t, []string{"/old/golangci-lint"}, map[string]string{"/old/golangci-lint": "golangci-lint has version " + version + " built with go1.27.2 from 25c5369"}, "")
		f := lintFinding(root, "2.15.0")
		if f.OK || f.Fix == nil || len(f.Fix.Choices) != 1 {
			t.Fatalf("%s: a different version asks how to go on: %+v", version, f)
		}
		if !strings.Contains(f.Detail, "found /old/golangci-lint (golangci-lint has version "+version) || !strings.Contains(f.Detail, word+" than the 2.15.0") {
			t.Errorf("%s: the detail names the file, the line and the order: %s", version, f.Detail)
		}
		options := f.Fix.Choices[0].Options
		if len(options) != 2 || !strings.Contains(options[0].Label, "recommended") || !strings.Contains(options[1].Label, "Keep /old/golangci-lint ("+version+")") {
			t.Fatalf("%s: install into smith's folder first, keep the one found second: %+v", version, options)
		}
		step := options[0].Fix.Steps[0]
		if step.Env[0] != "GOBIN="+toolrun.ManagedBinDir(root) || step.Argv[0] != "go" || step.Argv[1] != "install" || !strings.HasSuffix(step.Argv[2], "@v2.15.0") {
			t.Errorf("%s: go install of the pin into the managed folder: %+v", version, step)
		}
		if len(options[0].Fix.Files) != 0 || len(options[1].Fix.Steps) != 0 {
			t.Errorf("%s: only the keep answer writes smith.toml, and it runs nothing", version)
		}
	}
}

// The check and the lint task both look in smith's own folder first.
func TestLintFindingUsesTheManagedCopyBeforePath(t *testing.T) {
	root := t.TempDir()
	_, managed := putLint(t, toolrun.ManagedBinDir(root))
	fakeMachine(t, []string{"/old/golangci-lint"}, map[string]string{
		managed:              pinLine,
		"/old/golangci-lint": "golangci-lint has version 2.16.0 built with go1.27.2",
	}, "")
	f := lintFinding(root, "2.15.0")
	if !f.OK || !strings.Contains(f.Detail, "found "+managed) || !strings.Contains(f.Detail, "also on PATH, after it: /old/golangci-lint") {
		t.Fatalf("the copy smith installed wins over PATH: %+v", f)
	}
}

// (a) the pinned copy is in the go bin folder, but another one comes first on PATH.
func TestLintFindingExplainsACopyThatComesFirstOnPath(t *testing.T) {
	root := t.TempDir()
	bin, file := putLint(t, "")
	old := filepath.Join("old", "golangci-lint")
	fakeMachine(t, []string{old, file}, map[string]string{
		old:  "golangci-lint has version 1.64.8 built with go1.24",
		file: pinLine,
	}, bin)
	f := lintFinding(root, "2.15.0")
	if f.OK || f.Fix == nil {
		t.Fatalf("shadowed: still not fine, and the ways on are offered: %+v", f)
	}
	for _, want := range []string{old, "1.64.8", bin, "2.15.0", "move", "earlier in PATH", "does not edit PATH or delete files", "smith doctor --fix"} {
		if !strings.Contains(f.Remedy+f.Detail, want) {
			t.Errorf("the explanation lacks %q:\nDetail: %s\nRemedy: %s", want, f.Detail, f.Remedy)
		}
	}
}

// (b) other output formats: a v prefix and a different distribution still pass.
func TestLintFindingReadsAVersionPrefixAndAnotherDistribution(t *testing.T) {
	for _, says := range []string{"golangci-lint has version v2.15.0 built with go1.26", "golangci-lint version 2.15.0 (nix)"} {
		fakeMachine(t, []string{"/x/golangci-lint"}, map[string]string{"/x/golangci-lint": says}, "")
		if f := lintFinding(t.TempDir(), "2.15.0"); !f.OK {
			t.Errorf("%q should pass: %+v", says, f)
		}
	}
}

// (d) installed in the go bin folder, which is not on PATH at all.
func TestLintFindingSaysTheGoBinFolderIsNotOnPath(t *testing.T) {
	bin, file := putLint(t, "")
	fakeMachine(t, nil, map[string]string{file: pinLine}, bin)
	f := lintFinding(t.TempDir(), "2.15.0")
	if f.OK || !strings.HasPrefix(f.Detail, "golangci-lint is not on PATH; the pinned 2.15.0 is installed in "+bin) {
		t.Fatalf("not on PATH: %+v", f)
	}
	if !strings.Contains(f.Remedy, "add it to PATH") || !strings.Contains(f.Remedy, bin) {
		t.Errorf("the remedy names the folder: %s", f.Remedy)
	}
}

func TestLintFindingOffersTheInstallWhenNothingIsInstalled(t *testing.T) {
	fakeMachine(t, nil, nil, t.TempDir())
	f := lintFinding(t.TempDir(), "2.15.0")
	if f.OK || f.Detail != "golangci-lint is not on PATH" || f.Fix == nil {
		t.Fatalf("missing: %+v", f)
	}
	options := f.Fix.Choices[0].Options
	if len(options) != 2 || !strings.Contains(options[0].Label, "smith's own folder") || !strings.Contains(options[1].Label, "go install") {
		t.Fatalf("missing: smith's folder, or go install: %+v", options)
	}
}

func TestKeepingTheVersionFoundIsRecordedInSmithTomlAndAcceptedAfterwards(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "smith.toml"), []byte("[workspace]\nname = \"w\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	fakeMachine(t, []string{"/old/golangci-lint"}, map[string]string{"/old/golangci-lint": "golangci-lint has version 2.16.0 built with go1.27.2"}, "")
	f := lintFinding(root, "2.15.0")
	keep := f.Fix.Choices[0].Options[1].Fix
	if len(keep.Files) != 1 || keep.Files[0].Path != "smith.toml" || !strings.Contains(keep.Files[0].Content, "[toolchains]\ngolangci-lint = \"2.16.0\"") {
		t.Fatalf("the keep answer writes the accepted version: %+v", keep)
	}
	if err := os.WriteFile(filepath.Join(root, "smith.toml"), []byte(keep.Files[0].Content), 0o644); err != nil {
		t.Fatal(err)
	}
	after := lintFinding(root, "2.15.0")
	if !after.OK || !strings.Contains(after.Detail, "kept on purpose by [toolchains] in smith.toml") || !strings.Contains(after.Detail, "may differ from CI") {
		t.Fatalf("after keeping, the check passes and says why: %+v", after)
	}
	// A different version than the accepted one still warns.
	fakeMachine(t, []string{"/old/golangci-lint"}, map[string]string{"/old/golangci-lint": "golangci-lint has version 2.16.1 built with go1.27.2"}, "")
	if again := lintFinding(root, "2.15.0"); again.OK {
		t.Fatalf("another version is not accepted: %+v", again)
	}
}

func TestPathMatchesListsFilesInPathOrder(t *testing.T) {
	a, b := t.TempDir(), t.TempDir()
	name := lintProgram + exeSuffix(goruntime.GOOS)
	for _, dir := range []string{b, a} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	path := strings.Join([]string{a, t.TempDir(), b}, string(os.PathListSeparator))
	got := pathMatches(lintProgram, path, goruntime.GOOS)
	if len(got) != 2 || got[0] != filepath.Join(a, name) || got[1] != filepath.Join(b, name) {
		t.Fatalf("the first folder on PATH comes first: %v", got)
	}
}
