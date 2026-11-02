package resultdialog

import (
	"reflect"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

var missingDotnet = []string{
	"Creating app api-cs in apps/api-cs: a console app",
	"Wrote 12 file(s):",
	"  apps/api-cs/Program.cs",
	"Verifying: smith run build test -p api-cs",
	"dotnet is not installed or not on PATH. Install it from https://dot.net/download, or run smith tools install to choose how, then run the command again.",
	"api-cs was created and its files are kept, but a tool it needs is missing.",
}

func TestAnAddThatCouldNotCheckBecauseAToolIsMissingIsRecognised(t *testing.T) {
	got, ok := ParseMissingTool([]string{"add", "app", "api-cs", "--lang", "dotnet"}, missingDotnet)
	want := MissingTool{Project: "api-cs", Dir: "apps/api-cs", Files: 12, Program: "dotnet", Title: "the .NET SDK", URL: "https://dot.net/download", Check: []string{"run", "build", "test", "-p", "api-cs"}}
	if !ok || !reflect.DeepEqual(got, want) {
		t.Fatalf("ok=%v\n got %+v\nwant %+v", ok, got, want)
	}
}

func TestOtherFailuresAreNotAMissingTool(t *testing.T) {
	for name, c := range map[string]struct {
		argv, lines []string
	}{
		"not an add":       {[]string{"run", "test"}, missingDotnet},
		"the test failed":  {[]string{"add", "app", "web"}, []string{"Wrote 3 file(s):", "the new project web fails its first build"}},
		"a flag, no name":  {[]string{"add", "app", "--help"}, missingDotnet},
		"too few words":    {[]string{"add"}, missingDotnet},
		"no output at all": {[]string{"add", "app", "web"}, nil},
	} {
		if _, ok := ParseMissingTool(c.argv, c.lines); ok {
			t.Errorf("%s: taken for a missing tool", name)
		}
	}
}

func TestTheFailureDialogSaysTheProjectExistsAndOffersTheWaysOn(t *testing.T) {
	tool, _ := ParseMissingTool([]string{"add", "app", "api-cs"}, missingDotnet)
	m := NewMissingTool(tool).Place(100, 30)
	view := pane.Strip(strings.Join(m.Lines(theme.Default()), "\n"))
	for _, want := range []string{"api-cs was created in apps/api-cs (12 files).", "It could not be checked yet: the .NET SDK is not installed.",
		"Open the install page", "dot.net/download (copies the link)", "Install it with smith", "Check again", "Continue without checking", "Doctor keeps reminding you", "Nothing is lost: your project stays on disk."} {
		if !strings.Contains(view, want) {
			t.Errorf("the failure dialog lacks %q:\n%s", want, view)
		}
	}
	steps := map[string]Outcome{}
	for i, label := range m.Labels() {
		next := m
		next.cursor = i
		_, out := next.Key("enter")
		steps[label] = out
	}
	if steps["Open the install page"].Open != "https://dot.net/download" {
		t.Errorf("the install page: %+v", steps["Open the install page"])
	}
	if !reflect.DeepEqual(steps["Install it with smith"].Install, []string{"tools", "install"}) {
		t.Errorf("install: %+v", steps["Install it with smith"])
	}
	if !reflect.DeepEqual(steps["Check again"].Check, []string{"run", "build", "test", "-p", "api-cs"}) {
		t.Errorf("check: %+v", steps["Check again"])
	}
	if out := steps["Continue without checking"]; !out.Done || !strings.Contains(out.Note, "Doctor will keep reminding you") {
		t.Errorf("continue: %+v", out)
	}
	if _, out := m.Key("esc"); !out.Done || out.Install != nil || out.Check != nil {
		t.Errorf("esc must only close: %+v", out)
	}
}

func TestWithoutAnAddressOrACheckLineThoseChoicesAreLeftOut(t *testing.T) {
	m := NewMissingTool(MissingTool{Project: "x", Title: "Foo"})
	if got := m.Labels(); !reflect.DeepEqual(got, []string{"Install it with smith", "Continue without checking"}) {
		t.Fatalf("labels %v", got)
	}
}

func TestACommittableChangeStartsOnLeavingItUncommitted(t *testing.T) {
	argv := []string{"add", "lib", "core", "--lang", "go", "--yes"}
	m := NewAfterChange(argv, []string{"Creating lib core in libs/core: go", "Wrote 4 file(s):"}).Place(100, 30)
	if m.Cursor() != 0 || m.Labels()[0] != "Leave them uncommitted" || m.Labels()[1] != "Commit these changes now" {
		t.Fatalf("smith never commits unasked: the default is to leave: %v at %d", m.Labels(), m.Cursor())
	}
	if _, out := m.Key("enter"); !out.Done || out.Commit != nil {
		t.Fatalf("enter on the default commits nothing: %+v", out)
	}
	m, _ = m.Key("down")
	if _, out := m.Key("enter"); !reflect.DeepEqual(out.Commit, []string{"commit", "add", "lib", "core"}) {
		t.Fatalf("the commit is named from the operation: %+v", out)
	}
	view := pane.Strip(strings.Join(m.Lines(theme.Default()), "\n"))
	for _, want := range []string{"Added lib core", "core was created in libs/core (4 files).", "Use it from an app: press + on the app and pick core", "smith install -w <app>", "chore: add lib core"} {
		if !strings.Contains(view, want) {
			t.Errorf("the result lacks %q:\n%s", want, view)
		}
	}
}

func TestAnAppResultDoesNotSendYouToAnApp(t *testing.T) {
	m := NewAfterChange([]string{"add", "app", "web"}, []string{"Wrote 4 file(s):"}).Place(100, 30)
	if view := pane.Strip(strings.Join(m.Lines(theme.Default()), "\n")); strings.Contains(view, "Use it from an app") {
		t.Fatalf("only a library is used from an app:\n%s", view)
	}
}

func TestOperationTextStopsAtTheFirstFlag(t *testing.T) {
	if got := OperationText([]string{"add", "app", "web", "--lang", "go", "--yes"}); got != "add app web" {
		t.Fatalf("%q", got)
	}
	if got := OperationText(nil); got != "" {
		t.Fatalf("%q", got)
	}
}
