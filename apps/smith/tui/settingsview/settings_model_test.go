package settingsview

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// fake is a backend that records what it was asked and answers from a snapshot.
type fake struct {
	snap     Snapshot
	calls    []string
	setErr   error
	reviewed TrustReview
	saved    [2]string
}

func (f *fake) backend() Backend {
	return Backend{
		Load: func() Snapshot { return f.snap },
		SetPlugin: func(name string, enable bool) error {
			f.calls = append(f.calls, fmt.Sprintf("plugin %s %v", name, enable))
			for i := range f.snap.Plugins {
				if f.snap.Plugins[i].Name == name {
					f.snap.Plugins[i].Enabled = enable
				}
			}
			return f.setErr
		},
		SetArchitecture: func(name string) error {
			f.calls = append(f.calls, "architecture "+name)
			if f.setErr != nil {
				return f.setErr
			}
			f.snap.Architecture = name
			return nil
		},
		Review: func(name string) (TrustReview, error) { return f.reviewed, nil },
		Trust: func(name string, pin bool) (string, error) {
			f.calls = append(f.calls, fmt.Sprintf("trust %s pin=%v", name, pin))
			for i := range f.snap.External {
				f.snap.External[i].Trusted = true
			}
			return "Trusted " + name, nil
		},
		Untrust: func(name string) (string, error) {
			f.calls = append(f.calls, "untrust "+name)
			return "No longer trusting " + name, nil
		},
		SetSetting: func(path string, values []string) (string, error) {
			f.calls = append(f.calls, "set "+path+" "+strings.Join(values, "|"))
			if f.setErr != nil {
				return "", f.setErr
			}
			return path + " saved", nil
		},
		UnsetSetting: func(path string) (string, error) {
			f.calls = append(f.calls, "unset "+path)
			return path + " reset", nil
		},
		Style: StylePrefs{
			Path: "/home/me/.smith/tui.json", Dark: true,
			Save:   func(t, i string) error { f.saved = [2]string{t, i}; return nil },
			Getenv: func(string) string { return "" },
		},
	}
}

func sample() *fake {
	return &fake{
		reviewed: TrustReview{Text: "\nThe external plugin \"widget\" wants to run on this machine.\n  from      smith.toml\n  command   widget --serve\n"},
		snap: Snapshot{
			Architecture: "vertical-slices",
			Architectures: []ArchitectureChoice{
				{Name: "vertical-slices", Meaning: "code grouped by what it does", Writes: "adds [conventions.vertical-slices] to smith.toml"},
				{Name: "none", Meaning: "smith checks no code layout", Writes: "removes the lines that turn the other choices on"},
			},
			Plugins:    []Plugin{{"go", true}, {"node", true}, {"python", false}},
			Addons:     []string{"eslint", "jest"},
			External:   []ExternalPlugin{{Name: "widget", Source: "smith.toml", Command: "widget --serve", Detail: "never trusted here"}},
			Installed:  []InstalledPlugin{{Name: "kiwi", Version: "1.2.0", Dir: "/home/me/.smith/plugins/kiwi"}},
			SearchPath: []string{"/ws/.smith/plugins", "/home/me/.smith/plugins"},
			Settings: []Setting{
				{Path: "ci.devcontainer", Type: "boolean", Explanation: "Write a devcontainer.", Default: "false"},
				{Path: "ci.providers", Type: "array", Explanation: "Which CI systems.", Default: `["github"]`, Choices: []string{"github", "azure", "gitlab"}, Value: `["github"]`, Set: true, Current: []string{"github"}},
				{Path: "ci.branches", Type: "array", Explanation: "Branches.", Default: "the release branch"},
				{Path: "cache.max_size_mb", Type: "integer", Explanation: "Cache size.", Default: "2048", Value: "512", Set: true, Current: []string{"512"}},
				{Path: "release.branch", Type: "string", Explanation: "The release branch.", Default: "main"},
			},
			Recipes: []Recipe{{Path: "team.recipe.toml", Name: "team", Version: "1.0.0"}, {Path: "old.recipe.toml", Error: "no [recipe] table"}},
		},
	}
}

func plainTheme() theme.Theme {
	return theme.New(theme.Choice{Tokens: theme.Plain, Icons: theme.Unicode})
}

func keyMsg(name string) tea.KeyMsg {
	switch name {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "left":
		return tea.KeyMsg{Type: tea.KeyLeft}
	case "right":
		return tea.KeyMsg{Type: tea.KeyRight}
	case " ":
		return tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(name)}
}

// settle runs the background work of a result and gives each message back to the model.
func settle(m Model, res Result) (Model, []Result) {
	var results []Result
	var run func(cmd tea.Cmd)
	run = func(cmd tea.Cmd) {
		if cmd == nil {
			return
		}
		switch msg := cmd().(type) {
		case tea.BatchMsg:
			for _, c := range msg {
				run(c)
			}
		case Msg:
			var r Result
			m, r = m.Handle(msg)
			results = append(results, r)
			run(r.Cmd)
		}
	}
	run(res.Cmd)
	return m, results
}

// press sends keys and settles whatever they start.
func press(m Model, keys ...string) (Model, []Result) {
	var all []Result
	for _, k := range keys {
		var res Result
		m, res = m.Key(keyMsg(k))
		all = append(all, res)
		var more []Result
		m, more = settle(m, res)
		all = append(all, more...)
	}
	return m, all
}

func opened(t *testing.T, f *fake) Model {
	t.Helper()
	m := New(f.backend(), []Form{{Title: "Add a project", Hint: "app, lib or package"}})
	m, cmd := m.Open()
	m, _ = settle(m, Result{Cmd: cmd})
	if !m.loaded {
		t.Fatal("the workspace was not read when the tab opened")
	}
	return m
}

func text(m Model, w, h int) string {
	return pane.Strip(strings.Join(m.View(plainTheme(), w, h, -1, -1), "\n"))
}

func TestThePluginsSectionListsTheBuiltInTheExternalAndTheInstalledPlugins(t *testing.T) {
	m := opened(t, sample())
	out := text(m, 100, 28)
	for _, want := range []string{
		"Sections", "Plugins", "Architecture", "Style", "Workspace", "Recipes", "Forms",
		"Built-in plugins", "[x] go", "[x] node", "[ ] python",
		"External plugins this workspace uses", "widget", "not trusted",
		"Installed on the search path", "kiwi", "1.2.0",
		"+ Install a plugin", "smith plugin install <source>",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("the Plugins section lacks %q:\n%s", want, out)
		}
	}
}

func TestSpaceTurnsABuiltInPluginOffAndTheListShowsIt(t *testing.T) {
	f := sample()
	m := opened(t, f)
	m, _ = press(m, "down", " ") // node
	if got := strings.Join(f.calls, ";"); got != "plugin node false" {
		t.Fatalf("calls: %q", got)
	}
	if out := text(m, 100, 28); !strings.Contains(out, "[ ] node") || !strings.Contains(out, "node is off") {
		t.Fatalf("the row and the note show the change:\n%s", out)
	}
	press(m, "enter")
	if got := f.calls[len(f.calls)-1]; got != "plugin node true" {
		t.Fatalf("Enter toggles too: %q", got)
	}
}

func TestARefusedChangeStaysOnScreenWithItsReason(t *testing.T) {
	f := sample()
	f.setErr = errors.New("at least one plugin must stay enabled")
	m := opened(t, f)
	m, results := press(m, " ")
	if !strings.Contains(text(m, 100, 28), "at least one plugin must stay enabled") {
		t.Fatal("the reason is not shown")
	}
	failed := false
	for _, r := range results {
		failed = failed || r.Failed
	}
	if !failed {
		t.Fatal("the shell is told it failed, for the toast")
	}
}

func TestTrustingAnExternalPluginShowsTheReviewFirst(t *testing.T) {
	f := sample()
	m := opened(t, f)
	m, _ = press(m, "down", "down", "down", "enter") // go, node, python, widget
	if !m.Modal() {
		t.Fatal("the review is an editor: plain keys are its own")
	}
	out := text(m, 100, 28)
	for _, want := range []string{"Trust widget?", `wants to run on this machine`, "command   widget --serve", "y trust", "p trust and pin"} {
		if !strings.Contains(out, want) {
			t.Errorf("the review lacks %q:\n%s", want, out)
		}
	}
	if len(f.calls) != 0 {
		t.Fatalf("nothing is recorded before the person says yes: %v", f.calls)
	}
	m, _ = press(m, "esc")
	if m.Modal() || len(f.calls) != 0 {
		t.Fatal("Esc cancels the review without trusting")
	}
	m, _ = press(m, "t", "p")
	if got := strings.Join(f.calls, ";"); got != "trust widget pin=true" {
		t.Fatalf("calls: %q", got)
	}
	if m.Modal() || !strings.Contains(text(m, 100, 28), "trusted") {
		t.Fatalf("the review closes and the row says trusted:\n%s", text(m, 100, 28))
	}
	press(m, "f")
	if got := f.calls[len(f.calls)-1]; got != "untrust widget" {
		t.Fatalf("f forgets a trusted plugin: %q", got)
	}
}

func TestInstallAsksForASourceAndHandsTheCommandLineToTheShell(t *testing.T) {
	f := sample()
	m := opened(t, f)
	m, _ = press(m, "end", "enter")
	if !m.Modal() || !strings.Contains(text(m, 100, 28), "Install a plugin") {
		t.Fatalf("the row opens a field: %s", text(m, 100, 28))
	}
	if !strings.Contains(text(m, 100, 28), "smith plugin install <source>") {
		t.Fatalf("the command line is shown before anything is typed: %s", text(m, 100, 28))
	}
	m, _ = press(m, "k", "i", "w", "i")
	if !strings.Contains(text(m, 100, 28), "smith plugin install kiwi") {
		t.Fatalf("the command line follows the typing: %s", text(m, 100, 28))
	}
	m, results := press(m, "enter")
	var run []string
	for _, r := range results {
		if r.Run != nil {
			run = r.Run
		}
	}
	if strings.Join(run, " ") != "plugin install kiwi" || m.Modal() {
		t.Fatalf("Enter hands plugin install kiwi to the shell and closes the field: %v", run)
	}
	if len(f.calls) != 0 {
		t.Fatalf("the tab must not install by itself: %v", f.calls)
	}
}

func TestInstallTakesTheDigestAfterTheSourceAndRefusesAnEmptyField(t *testing.T) {
	f := sample()
	m := opened(t, f)
	digest := strings.Repeat("ab", 32)
	m, _ = press(m, "end", "enter", "enter")
	if !m.Modal() || !strings.Contains(text(m, 100, 28), "Say where the plugin comes from.") {
		t.Fatalf("an empty field is refused with the reason: %s", text(m, 100, 28))
	}
	for _, r := range "./kiwi.zip " + digest {
		m, _ = press(m, string(r))
	}
	if !strings.Contains(text(m, 120, 28), "--sha256") {
		t.Fatalf("the digest becomes the flag: %s", text(m, 120, 28))
	}
	_, results := press(m, "enter")
	var run []string
	for _, r := range results {
		if r.Run != nil {
			run = r.Run
		}
	}
	if strings.Join(run, " ") != "plugin install ./kiwi.zip --sha256 "+digest {
		t.Fatalf("the command line: %v", run)
	}
}

func TestABooleanSettingFlipsOnSpaceAndItsDefaultIsRespected(t *testing.T) {
	f := sample()
	m := opened(t, f)
	m, _ = press(m, "left", "down", "down", "down", "right") // Workspace section
	press(m, " ")                                            // ci.devcontainer, unset, default false
	if got := strings.Join(f.calls, ";"); got != "set ci.devcontainer true" {
		t.Fatalf("calls: %q", got)
	}
}

func TestAFixedListIsAPickerThatSavesWhatIsTicked(t *testing.T) {
	f := sample()
	m := opened(t, f)
	m = m.Show(Workspace)
	m, _ = press(m, "down", "enter") // ci.providers
	out := text(m, 100, 28)
	if !m.Modal() || !strings.Contains(out, "[x] github") || !strings.Contains(out, "[ ] azure") {
		t.Fatalf("the picker starts from the value:\n%s", out)
	}
	m, _ = press(m, "down", " ", "enter") // tick azure
	if got := f.calls[len(f.calls)-1]; got != "set ci.providers github|azure" {
		t.Fatalf("saved %q", got)
	}
	if m.Modal() {
		t.Fatal("saving closes the picker")
	}
}

func TestATypedSettingKeepsTheEditorOpenWithTheReasonWhenTheSchemaRefuses(t *testing.T) {
	f := sample()
	m := opened(t, f)
	m = m.Show(Workspace)
	m, _ = press(m, "down", "down", "down", "enter") // cache.max_size_mb, 512
	if !m.Modal() || !strings.Contains(text(m, 100, 28), "512") {
		t.Fatalf("the field starts from the value:\n%s", text(m, 100, 28))
	}
	f.setErr = errors.New("cache.max_size_mb is at least 0")
	m, _ = press(m, "enter")
	if !m.Modal() || !strings.Contains(text(m, 100, 28), "cache.max_size_mb is at least 0") {
		t.Fatalf("the reason shows in the editor:\n%s", text(m, 100, 28))
	}
	f.setErr = nil
	m, _ = press(m, "esc")
	if m.Modal() {
		t.Fatal("Esc closes the editor")
	}
}

func TestAListWithoutFixedChoicesIsTypedWithCommasAndSplitBeforeItIsSaved(t *testing.T) {
	f := sample()
	m := opened(t, f)
	m = m.Show(Workspace)
	m, _ = press(m, "down", "down", "enter") // ci.branches
	press(m, "m", "a", "i", "n", ",", " ", "d", "e", "v", "enter")
	if got := f.calls[len(f.calls)-1]; got != "set ci.branches main|dev" {
		t.Fatalf("saved %q", got)
	}
}

func TestUResetsASettingToItsDefaultAndOnlyWhenItIsSet(t *testing.T) {
	f := sample()
	m := opened(t, f)
	m = m.Show(Workspace)
	if reason := m.Reasons()["settings.unset"]; !strings.Contains(reason, "default") {
		t.Fatalf("an unset setting cannot be reset: %q", reason)
	}
	m, _ = press(m, "down", "u") // ci.providers is set
	if got := f.calls[len(f.calls)-1]; got != "unset ci.providers" {
		t.Fatalf("calls %v", f.calls)
	}
	if m.Reasons()["settings.unset"] != "" && !strings.Contains(m.Reasons()["settings.unset"], "default") {
		t.Fatal("the reason is about the default")
	}
}

func TestChoosingAThemeSavesItAndHandsTheShellTheNewTheme(t *testing.T) {
	f := sample()
	m := opened(t, f)
	m = m.Show(Style)
	out := text(m, 100, 28)
	for _, want := range []string{"Theme", "( ) dark", "Symbols", "Preview", "tui.json"} {
		if !strings.Contains(out, want) {
			t.Errorf("the Style section lacks %q:\n%s", want, out)
		}
	}
	m, results := press(m, "down", "down", "enter") // light
	if f.saved != [2]string{"light", ""} {
		t.Fatalf("saved %v", f.saved)
	}
	var got *theme.Choice
	for _, r := range results {
		if r.Theme != nil {
			got = r.Theme
		}
	}
	if got == nil || got.Tokens.Name != "light" {
		t.Fatalf("the shell is told to switch to light: %+v", got)
	}
	if !strings.Contains(text(m, 100, 28), "(●) light") {
		t.Fatalf("the radio follows:\n%s", text(m, 100, 28))
	}
	press(m, "down", "down", "down", "down", "enter", "down", "enter") // symbols: auto, unicode -> ascii
	if f.saved != [2]string{"light", "ascii"} {
		t.Fatalf("saved %v", f.saved)
	}
}

func TestTheEnvironmentWinsOverTheSavedChoiceAndTheSectionSaysSo(t *testing.T) {
	f := sample()
	b := f.backend()
	b.Style.Getenv = func(k string) string {
		if k == "SMITH_THEME" {
			return "plain"
		}
		return ""
	}
	m := New(b, nil).Show(Style)
	m, _ = press(m, "down", "down") // light
	if out := text(m, 110, 30); !strings.Contains(out, "SMITH_THEME is set, so it wins") {
		t.Fatalf("no note about the variable:\n%s", out)
	}
	_, results := press(m, "enter")
	for _, r := range results {
		if r.Theme != nil && r.Theme.Tokens.Name != "plain" {
			t.Fatalf("the variable still decides what is drawn: %s", r.Theme.Tokens.Name)
		}
	}
}

func TestThePreviewShowsTheHighlightedChoiceNotTheSavedOne(t *testing.T) {
	m := opened(t, sample()).Show(Style)
	if out := text(m, 110, 30); !strings.Contains(out, "theme dark, unicode symbols") {
		t.Fatalf("auto on a dark terminal previews dark:\n%s", out)
	}
	m, _ = press(m, "down", "down", "down") // plain
	if out := text(m, 110, 30); !strings.Contains(out, "theme plain") {
		t.Fatalf("the preview follows the cursor:\n%s", out)
	}
}

func TestAFormOpensFromTheFormsSection(t *testing.T) {
	m := opened(t, sample()).Show(Forms)
	_, results := press(m, "enter")
	if len(results) == 0 || results[0].Form != "Add a project" {
		t.Fatalf("%+v", results)
	}
}

func TestRecipesAreListedWithWhatIsWrongWithABrokenOne(t *testing.T) {
	m := opened(t, sample()).Show(Recipes)
	out := text(m, 100, 28)
	if !strings.Contains(out, "team.recipe.toml") || !strings.Contains(out, "old.recipe.toml") || !strings.Contains(out, "invalid") {
		t.Fatalf("%s", out)
	}
	m, _ = press(m, "down")
	if !strings.Contains(text(m, 100, 28), "no [recipe] table") {
		t.Fatal("the detail says why the recipe is invalid")
	}
	if text(New(sample().backend(), nil).Show(Recipes), 100, 20); false {
		t.Fatal()
	}
}

func TestEverySectionFillsExactlyTheAreaAtEverySize(t *testing.T) {
	f := sample()
	for _, sec := range Sections {
		for _, editing := range []string{"", "setting", "picker", "review"} {
			for _, sz := range [][2]int{{140, 40}, {100, 30}, {70, 8}, {70, 3}, {40, 12}, {20, 5}, {10, 3}, {5, 2}, {1, 1}} {
				m := opened(t, f).Show(sec)
				switch {
				case editing == "setting" && sec == Workspace:
					m, _ = press(m, "down", "down", "down", "enter")
				case editing == "picker" && sec == Workspace:
					m, _ = press(m, "down", "enter")
				case editing == "review" && sec == Plugins:
					m, _ = press(m, "down", "down", "down", "enter")
				}
				lines := m.View(plainTheme(), sz[0], sz[1], -1, -1)
				if len(lines) != sz[1] {
					t.Fatalf("%v %s/%s: %d lines", sz, sec.Title(), editing, len(lines))
				}
				for i, l := range lines {
					if pane.Width(l) != sz[0] {
						t.Fatalf("%v %s/%s: line %d is %d cells: %q", sz, sec.Title(), editing, i, pane.Width(l), l)
					}
				}
			}
		}
	}
}

func TestAClickSelectsASectionOrARowAndADoubleClickChangesIt(t *testing.T) {
	f := sample()
	m := opened(t, f)
	w, h := 100, 28
	m, _ = m.Click(w, h, 2, 4, false) // Workspace in the list
	if m.Section() != Workspace || m.onItems {
		t.Fatalf("section %v onItems=%v", m.Section(), m.onItems)
	}
	m, _ = m.Click(w, h, 2, 1, false) // Plugins
	row := 0
	for y := 1; y < h; y++ {
		if idx, ok := m.frame(plainTheme(), w, h, -1, -1).rowItem[y]; ok && idx == 1 {
			row = y
		}
	}
	m, _ = m.Click(w, h, 40, row, false)
	if !m.onItems || m.cursor[Plugins] != 1 {
		t.Fatalf("click did not select the row: %+v", m.cursor)
	}
	_, res := m.Click(w, h, 40, row, true)
	if _, _ = settle(m, res); len(f.calls) != 1 || f.calls[0] != "plugin node false" {
		t.Fatalf("calls %v", f.calls)
	}
}

func TestTheFooterKeysHaveReasonsWhereTheyDoNothing(t *testing.T) {
	m := opened(t, sample())
	r := m.Reasons()
	if r["settings.trust"] == "" || r["settings.unset"] == "" {
		t.Fatalf("on a built-in plugin trust and reset do nothing: %v", r)
	}
	if r["settings.change"] != "" {
		t.Fatalf("a built-in plugin can be changed: %v", r)
	}
	m, _ = press(m, "down", "down", "down")
	if r = m.Reasons(); r["settings.trust"] != "" || r["settings.forget"] == "" {
		t.Fatalf("on an untrusted plugin: %v", r)
	}
	m, _ = press(m, "down")
	if r = m.Reasons(); r["settings.change"] == "" {
		t.Fatalf("an installed plugin is only listed: %v", r)
	}
	m, _ = press(m, "enter")
	if m.Modal() {
		t.Fatal("no editor for a listed plugin")
	}
}

func TestTheTipIsWantedUntilTheTabHasBeenOpenedOnce(t *testing.T) {
	f := sample()
	marked := 0
	b := f.backend()
	b.MarkTipSeen = func() { marked++ }
	m := New(b, nil)
	if !m.TipWanted() {
		t.Fatal("a new user is pointed at the tab")
	}
	m, cmd := m.Open()
	settle(m, Result{Cmd: cmd})
	if m.TipWanted() || marked != 1 {
		t.Fatalf("wanted=%v marked=%d", m.TipWanted(), marked)
	}
	b.TipSeen = true
	if New(b, nil).TipWanted() {
		t.Fatal("a person who has seen it is not pointed again")
	}
}

func TestWithoutABackendTheTabSaysItIsNotAvailable(t *testing.T) {
	m := New(Backend{}, nil)
	if m.Available() || !strings.Contains(text(m, 100, 20), "not available here") {
		t.Fatalf("%s", text(m, 100, 20))
	}
	press(m, "enter", " ", "t", "u", "f")
}

func TestAWorkspaceThatCannotBeReadShowsWhy(t *testing.T) {
	f := sample()
	f.snap = Snapshot{Err: errors.New("smith.toml: unexpected key")}
	m := opened(t, f)
	if out := text(m, 100, 20); !strings.Contains(out, "unexpected key") {
		t.Fatalf("%s", out)
	}
}

func TestAnUnsetSettingShowsItsDefaultValueAndASetOneShowsItsOwn(t *testing.T) {
	m := opened(t, sample())
	m, _ = press(m, "left", "down", "down", "down", "right") // Workspace section
	var rows []string
	for _, l := range strings.Split(text(m, 120, 30), "\n") {
		rows = append(rows, strings.Join(strings.Fields(l), " "))
	}
	screenText := strings.Join(rows, "\n")
	for _, want := range []string{"ci.devcontainer false", "ci.branches the release branch", "cache.max_size_mb 512", "release.branch main"} {
		if !strings.Contains(screenText, want) {
			t.Fatalf("the list lacks %q:\n%s", want, screenText)
		}
	}
	if strings.Contains(screenText, "devcontainer default") {
		t.Fatalf("an unset setting is shown as the word default:\n%s", screenText)
	}
}

func TestTheArchitectureSectionOffersTheChoicesAndMarksTheOneInUse(t *testing.T) {
	m := opened(t, sample()).Show(Architecture)
	out := text(m, 100, 28)
	for _, want := range []string{
		"how the code of this workspace is laid out", "Code layout",
		"(●) vertical-slices", "( ) none", "smith checks no code layout",
		"is laid out", "smith architecture vertical-slices",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("the Architecture section lacks %q:\n%s", want, out)
		}
	}
}

func TestEnterChoosesAnArchitectureSavesItAndTheRadioFollows(t *testing.T) {
	f := sample()
	m := opened(t, f).Show(Architecture)
	m, results := press(m, "down", "enter")
	if got := strings.Join(f.calls, ";"); got != "architecture none" {
		t.Fatalf("calls: %q", got)
	}
	out := text(m, 100, 28)
	if !strings.Contains(out, "( ) vertical-slices") || !strings.Contains(out, "(●) none") || !strings.Contains(out, "Architecture: none") {
		t.Fatalf("the radio and the note follow:\n%s", out)
	}
	for _, r := range results {
		if r.Failed {
			t.Fatalf("failed: %+v", r)
		}
	}
}

func TestARefusedArchitectureStaysOnScreenWithItsReason(t *testing.T) {
	f := sample()
	f.setErr = errors.New("refusing to write an invalid smith.toml")
	m := opened(t, f).Show(Architecture)
	m, _ = press(m, "down", "enter")
	if out := text(m, 100, 28); !strings.Contains(out, "refusing to write") || !strings.Contains(out, "(●) vertical-slices") {
		t.Fatalf("%s", out)
	}
}

// fitting is the sample with a choice that fits some projects and one that
// would check none.
func fitting() *fake {
	f := sample()
	f.snap.Architectures = []ArchitectureChoice{
		{Name: "vertical-slices", Meaning: "code grouped by what it does", Writes: "adds [conventions.vertical-slices] to smith.toml",
			AppliesTo: "go, node", Here: "checks api (go); skips infra (bicep) (its language is not covered)"},
		{Name: "hexagonal", Meaning: "ports and adapters", Writes: "adds [conventions.hexagonal] to smith.toml",
			AppliesTo: "go", Warning: "hexagonal applies to go, and none of your projects is written in those languages (infra (bicep)), so it would check nothing."},
		{Name: "none", Meaning: "smith checks no code layout", Writes: "removes the lines that turn the other choices on"},
	}
	return f
}

func TestTheArchitectureSectionSaysWhatEachChoiceAppliesToAndDoesHere(t *testing.T) {
	m := opened(t, fitting()).Show(Architecture)
	out := text(m, 110, 30)
	for _, want := range []string{"Applies to: go, node", "In this workspace: checks api (go); skips infra (bicep)"} {
		if !strings.Contains(out, want) {
			t.Errorf("the Architecture section lacks %q:\n%s", want, out)
		}
	}
}

func TestAnArchitectureThatChecksNothingHereAsksForASecondEnter(t *testing.T) {
	f := fitting()
	m := opened(t, f).Show(Architecture)
	m, _ = press(m, "down", "enter")
	if len(f.calls) != 0 {
		t.Fatalf("applied before the warning was read: %v", f.calls)
	}
	if out := text(m, 110, 30); !strings.Contains(out, "hexagonal would check none of your projects. Enter again to choose it anyway.") {
		t.Fatalf("no warning:\n%s", out)
	}
	m, _ = press(m, "enter")
	if got := strings.Join(f.calls, ";"); got != "architecture hexagonal" {
		t.Fatalf("calls: %q", got)
	}
	_ = m
}

func TestMovingAwayForgetsAPendingWarning(t *testing.T) {
	f := fitting()
	m := opened(t, f).Show(Architecture)
	_, _ = press(m, "down", "enter", "down", "up", "enter")
	if len(f.calls) != 0 {
		t.Fatalf("a warning shown before moving away must be shown again: %v", f.calls)
	}
}

func TestAnArchitectureThatFitsIsChosenAtOnce(t *testing.T) {
	f := fitting()
	m := opened(t, f).Show(Architecture)
	_, _ = press(m, "down", "down", "enter")
	if got := strings.Join(f.calls, ";"); got != "architecture none" {
		t.Fatalf("calls: %q", got)
	}
}

func TestTheArchitectureSectionFitsASmallTerminal(t *testing.T) {
	m := opened(t, sample()).Show(Architecture)
	lines := m.View(plainTheme(), 80, 24, -1, -1)
	if len(lines) != 24 {
		t.Fatalf("%d lines", len(lines))
	}
	if out := text(m, 80, 24); !strings.Contains(out, "vertical-slices") || !strings.Contains(out, "none") {
		t.Fatalf("%s", out)
	}
}

func TestSearchAsksForWordsAndHandsThemToTheShell(t *testing.T) {
	f := sample()
	m := opened(t, f)
	m, _ = press(m, "end", "up", "enter")
	if !m.Modal() || !strings.Contains(text(m, 100, 28), "Find a plugin in the marketplace") {
		t.Fatalf("the row opens a field: %s", text(m, 100, 28))
	}
	m, _ = press(m, "d", "e", "p", "l", "o", "y", " ", "g", "o")
	m, results := press(m, "enter")
	var got *Result
	for i, r := range results {
		if r.Search {
			got = &results[i]
		}
	}
	if got == nil || got.Term != "deploy go" || m.Modal() {
		t.Fatalf("Enter hands the words to the shell and closes the field: %+v", results)
	}
	if len(f.calls) != 0 {
		t.Fatalf("the tab must not search by itself: %v", f.calls)
	}
}

func TestSearchWithNoWordsListsEveryPlugin(t *testing.T) {
	m := opened(t, sample())
	m, _ = press(m, "end", "up", "enter")
	_, results := press(m, "enter")
	for _, r := range results {
		if r.Search && r.Term == "" {
			return
		}
	}
	t.Fatalf("an empty field is a search for everything: %+v", results)
}
