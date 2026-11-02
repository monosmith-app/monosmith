package packageprompt

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func libs(project string) []Library {
	if project == "api" {
		return nil
	}
	return []Library{{Name: "core", Kind: "lib", Language: "node", Version: "1.2.0"}, {Name: "ui", Kind: "package", Language: "node"}}
}

func withLibs(project string) Prompt {
	p := NewInstall(project, []string{"web", "api"}).WithLanguages(map[string]string{"web": "node", "api": "python"}).WithLibraries(libs)
	return p.Place(100, 40)
}

func arrow(p Prompt, t tea.KeyType) Prompt { p, _ = key(p, t); return p }

func TestTheFirstQuestionIsLibraryOrRegistryAndTheRegistryIsTheDefault(t *testing.T) {
	p := withLibs("web")
	if p.focus != focusSource || !strings.Contains(screen(p), "Add: ‹ A package from the registry (default) ›") {
		t.Fatalf("the first control is the question:\n%s", screen(p))
	}
	p = typeText(arrow(p, tea.KeyDown), "core")
	if got := strings.Join(p.Argv(), " "); got != "install -w web --registry core" {
		t.Fatalf("the registry answer must not be asked again: %q", got)
	}
}

func TestALibraryIsPickedFromTheListWithItsKindAndLanguage(t *testing.T) {
	p := withLibs("web")
	p = arrow(p, tea.KeyRight) // to "A library of this workspace"
	s := screen(p)
	for _, want := range []string{"Add: ‹ A library of this workspace ›", "Library: ‹ core · lib · node ›", "1/2 (Left, Right)", "Range: ‹ Follow the workspace (default) ›"} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing %q:\n%s", want, s)
		}
	}
	if got := strings.Join(p.Argv(), " "); got != "install -w web workspace:core" {
		t.Fatalf("argv %q", got)
	}
	p = arrow(p, tea.KeyDown) // to the library
	p = arrow(p, tea.KeyRight)
	if got := strings.Join(p.Argv(), " "); got != "install -w web workspace:ui" {
		t.Fatalf("argv %q", got)
	}
	_, out := key(p, tea.KeyEnter)
	if !out.Done || !out.Confirmed {
		t.Fatalf("Enter goes ahead: %+v", out)
	}
}

func TestTheRangeIsAChoiceWithAMeaningForNode(t *testing.T) {
	p := arrow(withLibs("web"), tea.KeyRight)
	p = arrow(arrow(p, tea.KeyDown), tea.KeyDown) // library, then range
	p = arrow(p, tea.KeyRight)
	if got := strings.Join(p.Argv(), " "); got != "install -w web workspace:core --range *" || !strings.Contains(screen(p), "whatever its") {
		t.Fatalf("argv %q\n%s", got, screen(p))
	}
	p = arrow(p, tea.KeyRight)
	if got := strings.Join(p.Argv(), " "); got != "install -w web workspace:core --range ^1.2.0" {
		t.Fatalf("argv %q", got)
	}
}

func TestNoLibraryMeansTheDialogSaysWhyAndWillNotGoAhead(t *testing.T) {
	p := arrow(withLibs("api"), tea.KeyRight)
	if p.Argv() != nil || !strings.Contains(screen(p), "No library of this workspace can be linked to api") {
		t.Fatalf("argv %v\n%s", p.Argv(), screen(p))
	}
	p, out := key(p, tea.KeyEnter)
	if out.Done || !strings.Contains(screen(p), "no library to link") && !strings.Contains(screen(p), "There is no library to link") {
		t.Fatalf("it must not go ahead: %+v\n%s", out, screen(p))
	}
}

func TestRemoveAndAnUnconfiguredDialogAskNoFirstQuestion(t *testing.T) {
	r := NewRemove("web", []string{"web"}).WithLibraries(libs).Place(100, 30)
	if strings.Contains(screen(r), "Add:") {
		t.Fatal("remove has no first question")
	}
	if strings.Contains(screen(NewInstall("web", []string{"web"}).Place(100, 30)), "Add:") {
		t.Fatal("a dialog given no libraries asks nothing")
	}
}
