package workspacerecipe

import (
	"errors"
	"slices"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

var nodeJest = protocol.Stack{Language: "node", Scaffold: "typescript", Toolchain: []string{"typescript", "eslint", "jest"}}

var provides = map[string][]string{
	"typescript": {"build", "typecheck"},
	"eslint":     {"lint"},
	"jest":       {"test"},
	"vitest":     {"test"},
}

func node(stack string, st protocol.Stack) ProjectFact {
	return ProjectFact{Language: "node", StackName: stack, Stack: st}
}

func TestReadFixesTheMainLanguageItsStackAndItsSlots(t *testing.T) {
	f := Facts{
		Projects: []ProjectFact{
			node("node-jest", nodeJest), node("node-jest", nodeJest), {Language: "go"},
		},
		Layouts:  []Layout{{Name: "vertical-slices"}},
		Provides: provides,
	}
	r, err := Read(f, "")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"language=node", "scaffold=typescript", "toolchain=typescript", "lint=eslint", "test=jest", "structure=vertical-slices"}
	if !slices.Equal(r.Fixed, want) {
		t.Errorf("fixed %v, want %v", r.Fixed, want)
	}
	if r.Language != "node" || r.Projects != 2 || r.StackName != "node-jest" {
		t.Errorf("reading %+v", r)
	}
}

func TestReadKeepsTheToolchainWhenSlotsWouldChangeItsOrder(t *testing.T) {
	// jest before eslint: as slots they would apply as lint, then test.
	st := protocol.Stack{Language: "node", Toolchain: []string{"typescript", "jest", "eslint"}}
	r, err := Read(Facts{Projects: []ProjectFact{node("mine", st)}, Provides: provides}, "node")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"language=node", "toolchain=typescript,jest,eslint"}
	if !slices.Equal(r.Fixed, want) {
		t.Errorf("fixed %v, want %v", r.Fixed, want)
	}
}

func TestReadKeepsATaskTwoPluginsProvideInTheToolchain(t *testing.T) {
	st := protocol.Stack{Language: "node", Toolchain: []string{"jest", "vitest"}}
	r, _ := Read(Facts{Projects: []ProjectFact{node("both", st)}, Provides: provides}, "node")
	if want := []string{"language=node", "toolchain=jest,vitest"}; !slices.Equal(r.Fixed, want) {
		t.Errorf("fixed %v, want %v", r.Fixed, want)
	}
}

func TestReadTakesTheStackMostProjectsUse(t *testing.T) {
	vitest := protocol.Stack{Language: "node", Toolchain: []string{"vitest"}, Deploy: []string{"docker"}}
	f := Facts{Projects: []ProjectFact{node("node-jest", nodeJest), node("v", vitest), node("v", vitest)}, Provides: provides}
	r, _ := Read(f, "node")
	if want := []string{"language=node", "test=vitest", "deploy=docker"}; !slices.Equal(r.Fixed, want) || r.StackName != "v" {
		t.Errorf("fixed %v (stack %q), want %v", r.Fixed, r.StackName, want)
	}
}

func TestReadLeavesOutALayoutThatDoesNotCheckTheLanguage(t *testing.T) {
	f := Facts{Projects: []ProjectFact{{Language: "go", StackName: protocol.NoStack}}, Layouts: []Layout{{Name: "mvc", Languages: []string{"node"}}}}
	r, _ := Read(f, "")
	if want := []string{"language=go"}; !slices.Equal(r.Fixed, want) || r.StackName != "" {
		t.Errorf("fixed %v (stack %q), want %v", r.Fixed, r.StackName, want)
	}
}

func TestReadRefusesAnEmptyWorkspaceAndAnUnusedLanguage(t *testing.T) {
	if _, err := Read(Facts{}, ""); !errors.Is(err, ErrNoProjects) {
		t.Errorf("empty workspace: %v", err)
	}
	if _, err := Read(Facts{Projects: []ProjectFact{{Language: "go"}}}, "python"); err == nil {
		t.Error("a language no project uses was read")
	}
}

func TestLanguagesAndReadEachPutTheMainLanguageFirst(t *testing.T) {
	f := Facts{Projects: []ProjectFact{{Language: "python"}, {Language: "go"}, {Language: "go"}, {Language: "node"}}}
	if got := Languages(f); !slices.Equal(got, []string{"go", "node", "python"}) {
		t.Errorf("languages %v", got)
	}
	each := ReadEach(f)
	if len(each) != 3 || each[0].Language != "go" || each[0].Projects != 2 {
		t.Errorf("readings %+v", each)
	}
}
