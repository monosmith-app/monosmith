package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/depsview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
)

func TestTheProjectsViewShowsStackNeedsAndWhatIsAffected(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.has("4 projects", "1 affected", "Stack", "node-vitest", "Language")
}

func TestTheProjectsViewRunsATaskAndGraphsWhatItNeeds(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("ctrl+b")
	if got := l.lastStarted(); !strings.Contains(got, "build") {
		t.Fatalf("ctrl+b started %q", got)
	}
	l.keys("alt+1")
	l.clickText("Graph: needs and used by")
	if got := l.lastStarted(); got != "graph --focus orders" {
		t.Fatalf("the graph action started %q", got)
	}
	l.has("graph --focus orders") // the command line is printed before it runs
}

func TestThePluginsViewListsBuiltInAndInstalledPluginsWithTheirUse(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("alt+7")
	l.has("Plugins", "4 built-in · 1 installed · 4 on", "go", "node", "python", "vitest", "acme-lint", "language", "orders, kit", "Used by")
	l.has("from     built-in")
}

func TestThePluginsViewTurnsAPluginOnOnlyAfterConfirmingTheCommand(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("alt+7", "down", "down") // python, which is off
	l.has("e runs smith plugin enable python")
	l.keys("e")
	l.has("smith plugin enable python") // the dialog shows the command line
	if len(l.started) != 0 {
		t.Fatalf("started %v before the confirmation", l.started)
	}
	l.keys("y")
	if got := l.lastStarted(); got != "plugin enable python" {
		t.Fatalf("started %q", got)
	}
}

func TestThePluginsViewDisablesAnEnabledPluginAndLeavesAddonsAlone(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("alt+7", "e", "y")
	if got := l.lastStarted(); got != "plugin disable go" {
		t.Fatalf("started %q", got)
	}
	l = newLab(t, true, 140, 34)
	l.keys("alt+7", "down", "down", "down", "e") // vitest is not switchable
	if len(l.started) != 0 || l.sh().confirm != nil {
		t.Fatal("an add-on plugin was switched")
	}
}

func TestTheDepsViewListsNewerReleasesAndThePolicy(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("alt+8")
	l.has("3 with newer releases: 1 patch · 1 minor · 1 major", "axios", "1.6.8", "zod", "react", "web, api",
		"One version", "every production dependency has one version", "axios (node) should be ^1.6.8 everywhere")
}

func TestTheDepsViewUpdatesTheMarkedSelectionWithDepsUp(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("alt+8", "space", "down", "down", "space") // axios and react
	l.has("[x] axios", "[ ] zod", "[x] react", "2 marked")
	l.keys("u")
	l.has("smith deps up axios react")
	if len(l.started) != 0 {
		t.Fatalf("started %v before the confirmation", l.started)
	}
	l.keys("y")
	if got := l.lastStarted(); got != "deps up axios react" {
		t.Fatalf("started %q", got)
	}
}

func TestTheDepsViewUpdatesTheRowUnderTheCursorWhenNothingIsMarked(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("alt+8", "down", "enter", "y")
	if got := l.lastStarted(); got != "deps up zod" {
		t.Fatalf("started %q", got)
	}
}

func TestTheDepsViewMarksAllAndAlignsVersions(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("alt+8", "a")
	l.has("[x] axios", "[x] zod", "[x] react", "3 marked")
	l.keys("a")
	l.has("0 marked")
	l.keys("v", "y")
	if got := l.lastStarted(); got != "deps up --align" {
		t.Fatalf("started %q", got)
	}
}

func TestTheDepsViewSaysWhenThereIsNothingToUpdateAndWhenThePolicyIsOff(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.m = func() Shell {
		sh := l.sh()
		sh.sources.Deps = func() depsview.Result { return depsview.Result{} }
		return sh
	}()
	l.keys("alt+8")
	l.has("Every dependency is on its latest release.", "not required")
	l.keys("u", "v")
	if len(l.started) != 0 || l.sh().confirm != nil {
		t.Fatal("a command was offered with nothing to update")
	}
}

func TestTheDepsViewShowsAFailedLookup(t *testing.T) {
	l := newLab(t, true, 100, 30)
	l.m = func() Shell {
		sh := l.sh()
		sh.sources.Deps = func() depsview.Result { return depsview.Result{Err: errString("the registry is unreachable")} }
		return sh
	}()
	l.keys("alt+8")
	l.has("the registry is unreachable")
}

func TestTheDoctorTabGroupsByProjectAndAppliesOneSafeFixAfterConfirming(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("alt+3")
	l.has("workspace", "web  1 failed", "orders  all passed", "Fix this…")
	l.keys("f") // the cursor is on go.work, which has no fix
	if l.sh().confirm != nil || len(l.started) != 0 {
		t.Fatal("a finding without a safe fix offered a command")
	}
	l.has("no safe fix")
	l.keys("down", "down", "f") // git, then web's Docker finding
	l.has("Apply one doctor fix", "$ smith doctor --fix --only 'Docker registries of web agree' --yes")
	if len(l.started) != 0 {
		t.Fatal("nothing runs before the answer")
	}
	l.keys("y")
	want := "doctor --fix --only Docker registries of web agree --yes"
	if got := l.lastStarted(); got != want {
		t.Fatalf("started %q, want %q", got, want)
	}
}

func TestTheDoctorTabListsFlatWhenGroupingIsOff(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("alt+3", "g")
	l.hasNot("orders  all passed", "web  1 failed")
	l.has("Docker registries of web agree")
}

func TestTheReleaseTabDrillsDownToTheCommitsAndPrintsTheNotesCommand(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("alt+4")
	l.has("web", "1.0.0", "1.1.0", "kit", "0.2.1", "Notes")
	l.hasNot("a1b2c3d")
	l.keys("enter")
	l.has("a1b2c3d feat(ui): dark mode", "9f8e7d6 fix!: cookie parsing", "BREAKING: drops the old cookie format")
	l.keys("down", "enter")
	l.has("dependency web releases (no commit of its own releases it)")
	l.keys("n")
	if got := l.lastStarted(); got != "release --dry-run --notes" {
		t.Fatalf("the notes started %q", got)
	}
	l.has("$ smith release --dry-run --notes")
}

// recipeWorkspace writes a workspace on disk whose apps are made from a team recipe.
func recipeWorkspace(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	recipe := "name = \"team\"\nversion = \"1.2.0\"\n\n[choices.language]\nask = \"Language\"\noptions = [\"python\", \"csharp\"]\n\n" +
		"[use]\nstructure = \"vertical-slices\"\n\n[when.language.python.use]\nlanguage = \"python\"\n\n[when.language.csharp.use]\nlanguage = \"dotnet\"\n"
	for name, content := range map[string]string{
		"smith.toml":               "schema = 1\n\n[workspace]\nname = \"acme\"\n\n[recipes]\napp = \"recipes/team.recipe.toml\"\n",
		"recipes/team.recipe.toml": recipe,
	} {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func TestTheRecipesTabBuildsASmithAddFromTheRecipesChoicesAndConfirmsFirst(t *testing.T) {
	l := newLab(t, true, 140, 40)
	l.keys("alt+9")
	l.has("Recipes", "new project", "made from recipes/team.recipe.toml", "Recipe team 1.2.0", "Roles depend on language")
	l.keys("c")
	l.has("type the project's name")
	if len(l.started) != 0 || l.sh().confirm != nil {
		t.Fatal("a command was offered without a name")
	}
	l.keys("down", "down", "enter")
	if sh := l.sh(); !sh.editing() {
		t.Fatal("Enter on the name edits it")
	}
	l.typeText("web").keys("enter", "down", "right")
	l.has("It puts, for these answers:", "structure", "vertical-slices", "smith add app web --choice language=python")
	l.keys("c")
	l.has("Add a project", "The command that will run:", "$ smith add app web --choice language=python")
	if len(l.started) != 0 {
		t.Fatal("nothing runs before the answer")
	}
	l.keys("y")
	if got := l.lastStarted(); got != "add app web --choice language=python" {
		t.Fatalf("started %q", got)
	}
}

func TestTypingInTheRecipesFormIsNotAShortcut(t *testing.T) {
	l := newLab(t, true, 140, 40)
	l.keys("alt+9", "down", "down", "enter")
	l.typeText("c?r")
	if l.sh().help != nil || len(l.started) != 0 {
		t.Fatal("a typed letter acted as a shortcut")
	}
	l.keys("enter")
	l.has("Name      c?r")
}

func TestTheRecipeWizardLeavesTheShellForSmithRecipeNew(t *testing.T) {
	l := newLab(t, true, 140, 40)
	l.keys("alt+9", "left")
	l.has("new recipe", "The recipe wizard", "smith recipe new team.recipe.toml", "in a terminal of its own")
	l.keys("c")
	if !l.quit {
		t.Fatal("the wizard asks on a terminal, so the shell steps aside")
	}
	if got := strings.Join(l.sh().Argv, " "); got != "recipe new team.recipe.toml" {
		t.Fatalf("the shell handed over %q", got)
	}
}

func TestTheRecipesTabIsOnTheMouseAndTheViewMenuToo(t *testing.T) {
	l := newLab(t, true, 140, 40)
	l.clickText("Recipes")
	l.has("Create", "new project")
	x, y := l.find("Kind")
	l.send(click(x, y))
	l.send(click(x, y)) // the selected row acts: the kind moves on to lib
	l.has("lib", "no recipe: made by the stack or the language plugin")
}

type errString string

func (e errString) Error() string { return string(e) }

func TestRReloadsTheViewsThatLoadFromTheWorkspace(t *testing.T) {
	l := newLab(t, true, 140, 34)
	calls := 0
	l.m = func() Shell {
		sh := l.sh()
		sh.sources.Deps = func() depsview.Result { calls++; return sampleDeps }
		return sh
	}()
	l.keys("alt+8")
	l.keys("r")
	if calls != 2 {
		t.Fatalf("the dependencies were looked up %d times, want 2", calls)
	}
	l.keys("alt+3", "r")
	l.has("go.work lists every module")
}

func TestQuestionMarkOpensTheHelpOfEveryTab(t *testing.T) {
	cases := []struct {
		key  string
		want string
	}{
		{"alt+1", "Show or hide this help"},
		{"alt+7", "Show or hide this help"},
		{"alt+8", "Show or hide this help"},
		{"alt+2", "Reload the view"},
		{"alt+3", "Reload the view"},
		{"alt+4", "Show or hide this help"},
		{"alt+5", "Show or hide this help"},
	}
	for _, c := range cases {
		l := newLab(t, true, 100, 40)
		l.keys(c.key, "?")
		l.keys("end") // the keys every view shares are last
		l.has(c.want, "Esc", "Mouse")
		l.keys("?")
		l.hasNot("Keyboard shortcuts")
	}
}

func TestQuestionMarkIsTextWhileASettingIsBeingEdited(t *testing.T) {
	l := newLab(t, true, 100, 30)
	l.keys("alt+5", "enter")
	sh := l.sh()
	if !sh.editing() {
		t.Skip("the first setting is not text")
	}
	l.keys("?")
	if l.sh().help != nil {
		t.Fatal("the help opened while typing")
	}
}

func TestEveryTabFitsEightyColumns(t *testing.T) {
	l := newLab(t, true, 80, 26)
	// Ten tabs do not fit 56 cells with their full names: the long ones are shortened.
	l.has("Proj", "Run", "Doctor", "Rel", "Plug", "Deps", "Recipes", "Sett", "Guide")
	for _, key := range []string{"alt+1", "alt+2", "alt+3", "alt+4", "alt+7", "alt+8", "alt+9", "alt+5", "alt+6"} {
		l.keys(key)
		for i, line := range strings.Split(l.sh().View(), "\n") {
			if w := pane.Width(line); w != 80 {
				t.Fatalf("%s: line %d is %d cells wide: %q", key, i, w, line)
			}
		}
		if got := len(strings.Split(l.sh().View(), "\n")); got != 26 {
			t.Fatalf("%s: %d lines", key, got)
		}
	}
}

func TestTheTabsStayReachableBelowEightyColumns(t *testing.T) {
	l := newLab(t, true, 79, 24)
	l.has("Workspace", "Plug", "Deps", "Guide")
	l.clickText("Deps")
	l.has("3 with newer releases")
	l.clickText("Plug")
	l.has("built-in")
}

func TestClickingAPluginRowSelectsItAndAMarkBoxMarksADependency(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("alt+7")
	l.clickText("acme-lint")
	l.has("from     /home/me/.smith/plugins")
	l.keys("alt+8")
	l.clickText("zod")
	l.clickText("zod") // a second click on the row marks it
	l.has("[x] zod")
}

func TestTheWheelMovesThroughThePluginsAndDeps(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("alt+7")
	x, y := l.find("vitest")
	l.send(wheel(x+5, y, 1))
	l.has("e runs smith plugin disable node")
}

func TestTheHelpOfPluginsAndDepsListsTheirOwnKeysFirst(t *testing.T) {
	for key, want := range map[string][]string{
		"alt+1": {"Projects view", "Run the selected task"},
		"alt+7": {"Plugins view", "smith plugin enable|disable"},
		"alt+8": {"Deps view", "smith deps up", "--align"},
	} {
		l := newLab(t, true, 100, 40)
		l.keys(key, "?")
		l.has(want...)
	}
}

func TestAboutShowsTheLogoWhatSmithIsForAndTheDocsLink(t *testing.T) {
	l := newLab(t, true, 100, 40)
	l.keys("alt+h")
	l.clickText("About")
	l.has("About MonoSmith", "▄████", "scaffolds projects", "https://monosmith-app.github.io/monosmith/README.html", "[ OK ]")
	l.keys("esc")
	l.hasNot("About MonoSmith")
}

func TestAboutClosesOnTheOKButton(t *testing.T) {
	l := newLab(t, true, 100, 40)
	l.keys("alt+h")
	l.clickText("About")
	l.clickText("[ OK ]")
	l.hasNot("About MonoSmith")
}

func TestTheMenuBarStartsWithTheAppName(t *testing.T) {
	l := newLab(t, true, 100, 40)
	if first := l.lines()[0]; !strings.HasPrefix(first, " MonoSmith │ ") {
		t.Fatalf("bar %q", first)
	}
}
