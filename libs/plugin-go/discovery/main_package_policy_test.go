package discovery

import (
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestAModuleOutsideTheKindFoldersIsAnAppWhenItHoldsAMainPackage(t *testing.T) {
	ws := t.TempDir()
	write(t, ws, "api/go.mod", "module example.com/api\n")
	write(t, ws, "api/main.go", "package main\n\nfunc main() {}\n")
	write(t, ws, "kit/go.mod", "module example.com/kit\n")
	write(t, ws, "kit/kit.go", "package kit\n")
	write(t, ws, "kit/kit_test.go", "package main\n")
	write(t, ws, "tools/go.mod", "// smith:kind lib\nmodule example.com/tools\n")
	write(t, ws, "tools/main.go", "package main\n")

	res, err := Discover(ws, []string{"api/go.mod", "kit/go.mod", "tools/go.mod"})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]protocol.ProjectKind{"api": protocol.ProjectApp, "kit": protocol.ProjectLib, "tools": protocol.ProjectLib}
	for _, p := range res.Projects {
		if p.Kind != want[p.Name] {
			t.Errorf("%s: kind %s, want %s", p.Name, p.Kind, want[p.Name])
		}
	}
}
