package slicecheck

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
