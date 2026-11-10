package pipeline

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// tidyFixture is an app that imports a sibling module through a replace.
func tidyFixture(t *testing.T, marker string) string {
	t.Helper()
	ws := t.TempDir()
	files := map[string]string{
		"dep/go.mod":  "module example.com/dep\n\ngo 1.23\n",
		"dep/dep.go":  "package dep\n\nfunc Hi() {}\n",
		"app/go.mod":  "module example.com/app\n\ngo 1.23\n\nrequire example.com/dep v0.0.0" + marker + "\n\nreplace example.com/dep => ../dep\n",
		"app/main.go": "package main\n\nimport \"example.com/dep\"\n\nfunc main() { dep.Hi() }\n",
	}
	for name, body := range files {
		path := filepath.Join(ws, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return ws
}

func needGo(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go is not on PATH")
	}
}

func stubTidy(t *testing.T, stdout, stderr string, err error) {
	t.Helper()
	old := runTidy
	t.Cleanup(func() { runTidy = old })
	runTidy = func(string) (string, string, error) { return stdout, stderr, err }
}

var appProject = protocol.Project{Name: "app", Root: "app"}

func TestTidyModulePasses(t *testing.T) {
	needGo(t)
	f, ok := CheckModuleTidy(tidyFixture(t, ""), appProject)
	if !ok || !f.OK || f.Warning || f.Fix != nil {
		t.Fatalf("%+v", f)
	}
}

func TestTidyModuleWarnsForDirectImportMarkedIndirect(t *testing.T) {
	needGo(t)
	f, ok := CheckModuleTidy(tidyFixture(t, " // indirect"), appProject)
	if !ok || f.OK || !f.Warning || !strings.Contains(f.Detail, "indirect") {
		t.Fatalf("%+v", f)
	}
	if f.Fix == nil || f.Fix.Safe || f.Fix.Steps[0].Dir != "app" || strings.Join(f.Fix.Steps[0].Argv, " ") != "go mod tidy" || f.Fix.AppliesTo[0] != "app" {
		t.Fatalf("fix %+v", f.Fix)
	}
	if !strings.Contains(f.Remedy, "go mod tidy") {
		t.Fatal(f.Remedy)
	}
}

func TestTidyModuleSkipsWithoutGoMod(t *testing.T) {
	if _, ok := CheckModuleTidy(t.TempDir(), appProject); ok {
		t.Fatal("no go.mod, no check")
	}
}

func TestTidyModuleCutsLongDiff(t *testing.T) {
	needGo(t)
	var b strings.Builder
	for i := 0; i < 40; i++ {
		fmt.Fprintf(&b, "line %d\n", i)
	}
	stubTidy(t, b.String(), "", errors.New("exit status 1"))
	f, _ := CheckModuleTidy(tidyFixture(t, ""), appProject)
	if got := strings.Count(f.Detail, "\n"); got != tidyDiffLines || !strings.HasSuffix(f.Detail, "and 15 more lines") {
		t.Fatalf("%d lines: %q", got, f.Detail)
	}
}

func TestTidyModuleCouldNotCheckNeverFails(t *testing.T) {
	needGo(t)
	cases := map[string]struct {
		stderr string
		err    error
		want   string
	}{
		"old go":  {"flag provided but not defined: -diff", errors.New("exit status 2"), "Go 1.23"},
		"offline": {"module lookup disabled by GOPROXY=off", errors.New("exit status 1"), "offline"},
		"timeout": {"", toolrun.ErrTimeout, "longer than"},
		"other":   {"boom\nmore", errors.New("exit status 1"), "boom"},
	}
	for name, c := range cases {
		stubTidy(t, "", c.stderr, c.err)
		f, ok := CheckModuleTidy(tidyFixture(t, ""), appProject)
		if !ok || !f.OK || f.Warning || f.Fix != nil || !strings.Contains(f.Detail, "could not check") || !strings.Contains(f.Detail, c.want) {
			t.Errorf("%s: %+v", name, f)
		}
	}
}
