package tui

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/toolinstall"
	"github.com/monosmith-app/monosmith/apps/smith/tui/runner"
	"github.com/monosmith-app/monosmith/apps/smith/tui/sources"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// toolPlansJSON is what `smith tools install --dry-run --json` prints for one
// missing tool that a package manager installs without administrator rights.
func toolPlansJSON(t *testing.T) string {
	t.Helper()
	plan := toolinstall.ToolPlan{
		Status: toolinstall.Status{
			Need:  toolinstall.Need{ToolNeed: protocol.ToolNeed{Name: "go", Title: "Go", Why: "the go projects build with it", Consequence: "builds fail", Download: "https://go.dev/dl/"}},
			State: toolinstall.Missing,
		},
		Options: []toolinstall.Option{
			{Kind: toolinstall.OptionInstall, Value: "install:brew", Manager: "brew", Command: []string{"brew", "install", "go"}, Label: "Install with brew"},
			{Kind: toolinstall.OptionSkip, Value: "skip", Label: "Skip it"},
		},
		Default: "install:brew",
	}
	raw, err := json.Marshal([]toolinstall.ToolPlan{plan})
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func TestToolsInstallOpensTheDialogAndStreamsTheInstallInsteadOfLeavingTheShell(t *testing.T) {
	l := newLab(t, true, 110, 34)
	l.capture = func(argv []string) sources.Captured {
		return sources.Captured{OK: true, Text: toolPlansJSON(t)}
	}
	sh := l.sh()
	cmd := sh.execute([]string{"tools", "install"}, false)
	l.m = sh
	l.drive(cmd)
	if l.quit {
		t.Fatal("the shell stays open: the dialog asks the questions")
	}
	if got := strings.Join(l.captured[len(l.captured)-1], " "); got != "tools install --dry-run --json" {
		t.Fatalf("the choices are read without installing anything: %q", got)
	}
	l.has("Install the missing tools", "Go is missing", "Command: brew install go", "(default)")

	l.next = &scripted{lines: []string{"Pouring go"}, exit: runner.Exit{}}
	l.keys("enter")
	if l.lastStarted() != "tools install go --manager brew" {
		t.Fatalf("started %q", l.lastStarted())
	}
	l.has("done", "Installed.")
	l.keys("enter")
	l.hasNot("Install the missing tools")
}

func TestToolsInstallWithAFlagKeepsTheTerminalHandOver(t *testing.T) {
	l := newLab(t, true, 110, 34)
	sh := l.sh()
	sh.execute([]string{"tools", "install", "--yes"}, false)
	if !sh.quit || strings.Join(sh.Argv, " ") != "tools install --yes" {
		t.Fatalf("a command that answers for the person runs in the terminal: %v quit=%v", sh.Argv, sh.quit)
	}
}

func TestToolsInstallHandsOverWhenTheChoicesCannotBeRead(t *testing.T) {
	l := newLab(t, true, 110, 34)
	l.capture = func([]string) sources.Captured { return sources.Captured{Text: "boom"} }
	sh := l.sh()
	cmd := sh.execute([]string{"tools", "install", "go"}, false)
	l.m = sh
	l.drive(cmd)
	if !l.quit {
		t.Fatal("when the dialog cannot open, the terminal asks, as before")
	}
}

func TestIsToolsInstall(t *testing.T) {
	for argv, want := range map[string]bool{
		"tools install": true, "tools install go node": true, "tools install --yes": false,
		"tools install go --manager brew": false, "tools": false, "tools --json": false, "install": false,
	} {
		if got := isToolsInstall(strings.Fields(argv)); got != want {
			t.Errorf("%q: %v, want %v", argv, got, want)
		}
	}
}
