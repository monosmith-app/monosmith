package settingssource

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/architecture"
)

func TestChoosingAnArchitectureInTheTabWritesTheBytesSmithArchitectureWrites(t *testing.T) {
	tab, cliDir := twoWorkspaces(t)
	b := New(tab, t.TempDir(), func(string) string { return "" })
	if got := b.Load().Architecture; got != "none" {
		t.Fatalf("a workspace without the table uses %q", got)
	}
	for _, name := range []string{"vertical-slices", "none", "vertical-slices"} {
		if err := b.SetArchitecture(name); err != nil {
			t.Fatal(err)
		}
		cli(t, cliDir, architecture.Command(), name)
		same(t, tab, cliDir)
		if got := b.Load().Architecture; got != name {
			t.Fatalf("after choosing %s the tab reads %q", name, got)
		}
	}
	if !strings.Contains(read(t, tab), "[conventions.vertical-slices]") || !strings.Contains(read(t, tab), `name = "acme"   # stays`) {
		t.Fatalf("the table is in and the rest stayed:\n%s", read(t, tab))
	}
	if err := b.SetArchitecture("hexagonal"); err == nil {
		t.Fatal("an unknown architecture was accepted")
	}
}
