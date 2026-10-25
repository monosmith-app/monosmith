package scaffold

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestNextStepsSkipTheCdWhenTheWorkspaceIsTheCurrentDirectory(t *testing.T) {
	here := nextSteps("acme", ".")
	if strings.Contains(here, "cd ") {
		t.Errorf("no cd for the current directory:\n%s", here)
	}
	if !strings.Contains(here, "smith add app web") {
		t.Errorf("the next step is still printed:\n%s", here)
	}
	if there := nextSteps("acme", "acme"); !strings.Contains(there, "  cd acme\n") {
		t.Errorf("a new directory needs the cd:\n%s", there)
	}
}

func TestAddSaysItLeavesItsFilesUncommitted(t *testing.T) {
	if long := AddCommand("dev").Long; !strings.Contains(long, "uncommitted") {
		t.Errorf("smith add --help must say it leaves its files uncommitted:\n%s", long)
	}
}

func TestTryHintNamesATaskTheProjectHas(t *testing.T) {
	project := func(tasks ...string) *projectgraph.Project {
		p := &projectgraph.Project{Project: protocol.Project{Name: "mods", Tasks: map[string]protocol.Task{}}}
		for _, task := range tasks {
			p.Tasks[task] = protocol.Task{Command: []string{"x"}}
		}
		return p
	}
	for name, tc := range map[string]struct {
		p    *projectgraph.Project
		want string
	}{
		"a project with tests":     {project("build", "test"), "Try: smith test -p mods"},
		"a Bicep lib, no tests":    {project("build", "lint"), "Try: smith build -p mods"},
		"a project with no tasks":  {project(), ""},
		"a project not discovered": {nil, ""},
	} {
		if got := tryHint(tc.p); got != tc.want {
			t.Errorf("%s: %q, want %q", name, got, tc.want)
		}
	}
}

func TestTheWorkspaceConfigDoesNotSayVerticalSlicesIsForGoAndPythonOnly(t *testing.T) {
	if strings.Contains(configTemplate, "Go and Python") {
		t.Errorf("the plugin covers more languages than that:\n%s", configTemplate)
	}
	if !strings.Contains(configTemplate, "Dart, Rust, Ruby") || !strings.Contains(configTemplate, "Node projects") {
		t.Errorf("the comment names the languages the convention applies to:\n%s", configTemplate)
	}
}
