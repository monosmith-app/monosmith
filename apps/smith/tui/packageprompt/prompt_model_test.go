package packageprompt

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/dialog"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

func typeText(p Prompt, text string) Prompt {
	for _, r := range text {
		p, _ = p.Key(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	return p
}

func key(p Prompt, t tea.KeyType) (Prompt, dialog.Outcome) { return p.Key(tea.KeyMsg{Type: t}) }

func plain() theme.Theme { return theme.New(theme.Choice{Tokens: theme.Plain, Icons: theme.ASCII}) }

func screen(p Prompt) string { return pane.Strip(strings.Join(p.Lines(plain()), "\n")) }

func TestTypedPackagesEndAsTheInstallCommandOfTheProject(t *testing.T) {
	p := typeText(NewInstall("web", []string{"web"}).Place(100, 30), "zod @types/node")
	if got := strings.Join(p.Argv(), " "); got != "install -w web zod @types/node" {
		t.Fatalf("argv %q", got)
	}
	if !strings.Contains(screen(p), "$ smith install -w web zod @types/node") {
		t.Fatalf("the dialog shows the command that will run:\n%s", screen(p))
	}
	_, out := key(p, tea.KeyEnter)
	if !out.Done || !out.Confirmed {
		t.Fatalf("Enter on the packages goes ahead: %+v", out)
	}
}

func TestTheDevelopmentBoxAddsDAndRemoveHasNoBox(t *testing.T) {
	p := typeText(NewInstall("api", []string{"api"}).Place(100, 30), "pytest-cov")
	p, _ = key(p, tea.KeyTab) // to the box
	p, _ = p.Key(tea.KeyMsg{Type: tea.KeySpace})
	if got := strings.Join(p.Argv(), " "); got != "install -w api -D pytest-cov" || !strings.Contains(screen(p), "[x] a development") {
		t.Fatalf("argv %q\n%s", got, screen(p))
	}
	r := typeText(NewRemove("api", []string{"api"}).Place(100, 30), "left-pad")
	if got := strings.Join(r.Argv(), " "); got != "uninstall -w api left-pad" || strings.Contains(screen(r), "development") {
		t.Fatalf("argv %q\n%s", got, screen(r))
	}
	if r.Title() != "Remove packages" || !strings.Contains(screen(r), "Project: api") {
		t.Fatal(r.Title())
	}
}

func TestNothingTypedOrAFlagKeepsThePromptOpenAndSaysWhy(t *testing.T) {
	p := NewInstall("web", []string{"web"}).Place(100, 30)
	p, out := key(p, tea.KeyEnter)
	if out.Done || !strings.Contains(screen(p), "type at least one package") {
		t.Fatalf("an empty prompt went ahead or did not say why: %+v\n%s", out, screen(p))
	}
	p = typeText(p, "--registry=evil")
	p, out = key(p, tea.KeyEnter)
	if out.Done || p.Argv() != nil || !strings.Contains(screen(p), "is an option, not a package") {
		t.Fatalf("a flag went ahead: %+v\n%s", out, screen(p))
	}
}

func TestEscAndTheCancelButtonAndAClickOutsideCancel(t *testing.T) {
	p := typeText(NewInstall("web", []string{"web"}).Place(100, 30), "zod")
	if _, out := key(p, tea.KeyEsc); !out.Done || out.Confirmed {
		t.Fatalf("Esc: %+v", out)
	}
	r := p.Rect()
	if _, out := p.Click(r.X-1, r.Y); !out.Done || out.Confirmed {
		t.Fatalf("a click outside: %+v", out)
	}
	spans := p.buttonsRow(r.W - 4)
	if _, out := p.Click(r.X+spans[1][0]+1, r.Y+r.H-2); !out.Done || out.Confirmed {
		t.Fatalf("the Cancel button: %+v", out)
	}
	if _, out := p.Click(r.X+spans[0][0]+1, r.Y+r.H-2); !out.Done || !out.Confirmed {
		t.Fatalf("the Install button: %+v", out)
	}
}

func TestClickingTheBoxFlipsItAndTheLabelIsClickableToo(t *testing.T) {
	p := typeText(NewInstall("web", []string{"web"}).Place(100, 30), "zod")
	r := p.Rect()
	row := -1
	for i, l := range strings.Split(screen(p), "\n") {
		if strings.Contains(l, "[ ] a development") {
			row = i
		}
	}
	if row < 0 {
		t.Fatalf("no box:\n%s", screen(p))
	}
	p, out := p.Click(r.X+20, r.Y+row)
	if out.Done || !p.Dev() {
		t.Fatalf("clicking the line did not tick the box: %+v dev=%v", out, p.Dev())
	}
}

func TestTheProjectIsPreselectedAndLeftRightChooseAnother(t *testing.T) {
	p := NewInstall("orders", []string{"web", "orders", "kit"}).Place(100, 30)
	if p.Project() != "orders" || !strings.Contains(screen(p), "Project: ‹ orders ›") {
		t.Fatalf("the project on screen is preselected: %s", screen(p))
	}
	p, _ = p.Key(tea.KeyMsg{Type: tea.KeyShiftTab}) // from the packages to the project
	p, _ = key(p, tea.KeyRight)
	if p.Project() != "kit" {
		t.Fatalf("Right chose %q", p.Project())
	}
	p, _ = key(p, tea.KeyRight)
	p, _ = key(p, tea.KeyLeft)
	p, _ = key(p, tea.KeyEnter) // from the project to the packages
	p = typeText(p, "zod")
	if got := strings.Join(p.Argv(), " "); got != "install -w kit zod" {
		t.Fatalf("argv %q", got)
	}
}
