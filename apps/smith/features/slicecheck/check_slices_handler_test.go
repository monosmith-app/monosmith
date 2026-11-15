package slicecheck

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-go/scaffolding"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func project(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		"go.mod":              "module example.com/shop\n\ngo 1.26\n",
		"domain/order.go":     "package domain\nimport \"example.com/shop/adapters/sql\"\n",
		"adapters/sql/db.go":  "package sql\n",
		"application/call.go": "package application\nimport \"example.com/shop/domain\"\n",
	}
	for name, text := range files {
		p := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func runCheck(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	cmd := Command()
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs(append([]string{"check"}, args...))
	err := cmd.Execute()
	return out.String(), errOut.String(), err
}

func TestAnArchitectureCheckReportsTheForbiddenImport(t *testing.T) {
	_, stderr, err := runCheck(t, "--architecture", "hexagonal", "--lang", "go", project(t))
	if err == nil || !strings.Contains(err.Error(), "1 import(s) break the hexagonal architecture") {
		t.Fatalf("%v", err)
	}
	if !strings.Contains(stderr, "domain/order.go - imports example.com/shop/adapters/sql: core code must not depend on adapters code") {
		t.Fatalf("%s", stderr)
	}
}

func TestAnArchitectureThatPassesSaysSo(t *testing.T) {
	dir := project(t)
	if err := os.WriteFile(filepath.Join(dir, "domain", "order.go"), []byte("package domain\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, _, err := runCheck(t, "--architecture", "hexagonal", "--lang", "go", dir)
	if err != nil || !strings.Contains(out, "hexagonal architecture OK") {
		t.Fatalf("%q %v", out, err)
	}
}

func TestAnUnknownArchitectureNamesTheChoices(t *testing.T) {
	_, _, err := runCheck(t, "--architecture", "onion", project(t))
	if err == nil || !strings.Contains(err.Error(), "hexagonal, clean") {
		t.Fatalf("%v", err)
	}
}

// What smith add writes for Go must pass the check smith lint runs: the
// scaffold and the layout rules are one contract.
func TestTheGoScaffoldPassesTheSliceCheck(t *testing.T) {
	scaffolding.GoRunner = func(dir string, argv []string) error {
		// `go mod init` is the only thing the runner would write.
		if len(argv) > 2 && argv[1] == "mod" && argv[2] == "init" {
			return os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/w/x/shop\n\ngo 1.26\n"), 0o644)
		}
		return nil
	}
	for _, kind := range []protocol.ProjectKind{protocol.ProjectApp, protocol.ProjectLib} {
		root := t.TempDir()
		_, err := scaffolding.Generate(root, protocol.GenerateParams{Kind: kind, Name: "shop", Directory: "x/shop", Workspace: protocol.WorkspaceInfo{Name: "w", ModulePrefix: "example.com/w"}})
		if err != nil {
			t.Fatal(err)
		}
		out, errOut, err := runCheck(t, "--lang", "go", filepath.Join(root, "x", "shop"))
		if err != nil || !strings.Contains(out, "OK") {
			t.Errorf("%s scaffold breaks the slice rules: %v\n%s%s", kind, err, out, errOut)
		}
	}
}
