package tui

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/projectlist"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

func TestAddedProjectIsTheNameOfAnAddLine(t *testing.T) {
	if n, ok := addedProject([]string{"add", "package", "demo-pkg", "--lang", "node"}); !ok || n != "demo-pkg" {
		t.Fatalf("%q %v", n, ok)
	}
	for _, argv := range [][]string{{"add"}, {"add", "app"}, {"run", "test"}, {"add", "app", "--help"}} {
		if _, ok := addedProject(argv); ok {
			t.Errorf("%v is not an add of a project", argv)
		}
	}
}

func TestNextStepsNameTheKeysAndTheProject(t *testing.T) {
	line := addedNextSteps(theme.Default(), "demo-pkg")
	for _, want := range []string{"runs its tests", "adds another", "demo-pkg"} {
		if !strings.Contains(line, want) {
			t.Errorf("%q missing from %q", want, line)
		}
	}
}

func TestMainLanguageIsTheOneMostProjectsUse(t *testing.T) {
	s := Shell{list: projectlist.New([]projectlist.Project{
		{Name: "a", Language: "go"}, {Name: "b", Language: "node"}, {Name: "c", Language: "node"},
	})}
	if got := s.mainLanguage(); got != "node" {
		t.Fatalf("%q", got)
	}
	if (Shell{list: projectlist.New(nil)}).mainLanguage() != "" {
		t.Fatal("no projects")
	}
}
