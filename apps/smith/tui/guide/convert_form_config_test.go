package guide

import (
	"strings"
	"testing"
)

func convertMenu(t *testing.T) Model {
	t.Helper()
	m, ok := NewModel(Actions([]string{"node"}, []string{"node"}, []string{"node"})).Pick("Convert an Nx or Turborepo workspace")
	if !ok {
		t.Fatal("the convert form is in the menu")
	}
	return m
}

func TestTheConvertFormAsksOnlyWhatTheLevelNeedsAndBuildsTheCommand(t *testing.T) {
	// Conservative (first answer), then no commit (first answer): the plain command.
	m := keys(convertMenu(t), "enter", "enter").(Model)
	if got := strings.Join(m.Argv, " "); got != "convert" {
		t.Fatalf("argv %q", got)
	}
	// Full asks lint, test, build and CI; every answer is a presented choice.
	m = keys(convertMenu(t), "down", "enter", "down", "enter", "down", "down", "enter", "enter", "enter", "down", "enter").(Model)
	if got := strings.Join(m.Argv, " "); got != "convert --level full --lint smith --test vitest --build smith --ci smith --commit" {
		t.Fatalf("argv %q", got)
	}
}

func TestTheConvertFormIsOnlyOfferedOutsideAWorkspace(t *testing.T) {
	for _, a := range Actions([]string{"node"}, []string{"node"}, []string{"node"}) {
		if a.Presents == "convert" && (!a.OutsideOnly || !a.NeedsContent) {
			t.Fatalf("convert works on a repository that has files and no smith.toml: %+v", a)
		}
	}
}
