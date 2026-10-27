package tui

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/configedit"
	"github.com/monosmith-app/monosmith/apps/smith/pluginhost"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/guide"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/settingssource"
	"github.com/monosmith-app/monosmith/apps/smith/tui/settingsview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/userprefs"
	"github.com/spf13/cobra"
)

const settingsToml = `schema = 1

# The name of this workspace.
[workspace]
name = "acme"

[ci]
providers = ["github"]   # which CI
`

// settingsWorkspace is a real workspace on disk and a private smith home, with
// the tab wired to the real backend: what it changes is what the CLI changes.
type settingsWorkspace struct{ dir, home string }

func newSettingsLab(t *testing.T, w, h int) (*lab, settingsWorkspace) {
	t.Helper()
	t.Setenv("SMITH_HOME", t.TempDir())
	t.Setenv("CI", "")
	t.Setenv("SMITH_WORKSPACE_ROOT", "")
	ws := settingsWorkspace{dir: t.TempDir(), home: t.TempDir()}
	writeToml(t, ws.dir)
	l := newLab(t, true, w, h)
	sh := l.sh()
	sh.sources.Settings = settingssource.New(ws.dir, ws.home, "test", func(string) string { return "" })
	sh.settings = settingsview.New(sh.sources.Settings, settingsForms(guide.Actions([]string{"go", "node"}, []string{"go", "node", "python"}, []string{"go", "node"})))
	// Read the workspace now (Reload does not count as opening the tab), so
	// no test depends on how fast a message comes back.
	var read tea.Cmd
	sh.settings, read = sh.settings.Reload()
	l.m = sh
	l.patiently(read)
	return l, ws
}

// patiently is drive with all the time a slow machine needs: the tab's
// background work is real file work, and a message the lab gave up on is lost.
func (l *lab) patiently(cmd tea.Cmd) {
	if cmd == nil {
		return
	}
	done := make(chan tea.Msg, 1)
	go func() { done <- cmd() }()
	select {
	case msg := <-done:
		switch msg := msg.(type) {
		case nil, tickMsg:
		case tea.BatchMsg:
			for _, c := range msg {
				l.patiently(c)
			}
		case tea.QuitMsg:
			l.quit = true
		default:
			l.sendPatiently(msg)
		}
	case <-time.After(10 * time.Second):
		l.t.Fatal("background work of the tab did not finish")
	}
}

func (l *lab) sendPatiently(msg tea.Msg) *lab {
	var cmd tea.Cmd
	l.m, cmd = l.m.Update(msg)
	l.patiently(cmd)
	return l
}

// press sends key names and waits for what they start.
func (l *lab) press(names ...string) *lab {
	for _, n := range names {
		l.sendPatiently(keyMsg(n))
	}
	return l
}

// typed sends each rune as a key and waits.
func (l *lab) typed(text string) *lab {
	for _, r := range text {
		l.sendPatiently(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	return l
}

func (l *lab) doubleClickPatiently(x, y int) {
	l.sendPatiently(click(x, y))
	l.advance(120 * time.Millisecond)
	l.sendPatiently(click(x, y))
}

func writeToml(t *testing.T, dir string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "smith.toml"), []byte(settingsToml), 0o644); err != nil {
		t.Fatal(err)
	}
}

func tomlOf(t *testing.T, dir string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(dir, "smith.toml"))
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

// cliTwin is a second copy of the workspace, changed by the command line.
func cliTwin(t *testing.T, cmd *cobra.Command, args ...string) string {
	t.Helper()
	dir := t.TempDir()
	writeToml(t, dir)
	t.Chdir(dir)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("smith %s: %v\n%s", strings.Join(args, " "), err, out.String())
	}
	return tomlOf(t, dir)
}

func TestTheSettingsTabIsTheLastOfTheStripAndKeySevenOpensIt(t *testing.T) {
	l, _ := newSettingsLab(t, 140, 34)
	l.has("Settings") // in the strip
	l.press("7")
	if l.sh().tab != commands.TabSettings {
		t.Fatalf("on %q", l.sh().tab)
	}
	l.has("Sections", "Plugins", "Style", "Workspace", "Recipes", "Forms", "Built-in plugins", "[x] go", "[x] node", "+ Install a plugin", "smith plugin install <name|path>")
	footer := footerOf(l)
	for _, want := range []string{"↑↓ move", "Space change", "Tab sections", "r reload", "1-7 views"} {
		if !strings.Contains(footer, want) {
			t.Errorf("the footer lacks %q: %q", want, footer)
		}
	}
	if strings.Contains(footer, "t trust") || strings.Contains(footer, "u reset") {
		t.Fatalf("a built-in plugin cannot be trusted or reset: %q", footer)
	}
}

func TestTurningAPluginOffInTheTabWritesWhatPluginDisableWrites(t *testing.T) {
	l, ws := newSettingsLab(t, 140, 34)
	l.press("7", "space") // the first built-in plugin
	first := firstBuiltin(t)
	l.has("[ ] "+first, first+" is off")
	want := cliTwin(t, pluginhost.Command("test"), "disable", first)
	if got := tomlOf(t, ws.dir); got != want {
		t.Fatalf("the tab and smith plugin disable wrote different files.\ntab:\n%s\ncommand:\n%s", got, want)
	}
	l.press("space")
	l.has("[x] " + first)
}

func firstBuiltin(t *testing.T) string {
	t.Helper()
	return pluginhost.BuiltinNames()[0]
}

func TestChangingASettingInTheTabWritesWhatConfigSetWrites(t *testing.T) {
	l, ws := newSettingsLab(t, 140, 34)
	l.press("7", "left", "down", "down", "down", "enter") // Workspace, into its rows
	l.has("workspace.name", `"acme"`, "default")
	l.walkTo(t, "ci.devcontainer")
	l.press("space")
	l.has("ci.devcontainer = true")
	want := cliTwin(t, configedit.Command(), "set", "ci.devcontainer", "true")
	if got := tomlOf(t, ws.dir); got != want {
		t.Fatalf("the tab and smith config set wrote different files.\ntab:\n%s\ncommand:\n%s", got, want)
	}
	if !strings.Contains(tomlOf(t, ws.dir), "# which CI") || !strings.Contains(tomlOf(t, ws.dir), "# The name of this workspace.") {
		t.Fatalf("comments were lost:\n%s", tomlOf(t, ws.dir))
	}
	l.press("u")
	if got := tomlOf(t, ws.dir); got != settingsToml {
		t.Fatalf("u resets the setting to its default and leaves the file as it was:\n%s", got)
	}
}

// walkTo moves down the rows until the highlighted one holds text.
func (l *lab) walkTo(t *testing.T, text string) {
	t.Helper()
	for range 80 {
		if strings.Contains(l.selectedLine(), text) {
			return
		}
		l.press("down")
	}
	t.Fatalf("no row %q was ever selected:\n%s", text, l.view())
}

// selectedLine is the row the pointer of the right pane is on.
func (l *lab) selectedLine() string {
	for _, line := range l.lines() {
		// "│ left │ right │": the right pane is the third part.
		if parts := strings.SplitN(line, "│", 4); len(parts) == 4 && strings.HasPrefix(parts[2], " › ") {
			return parts[2]
		}
	}
	return ""
}

func TestATypedSettingIsAnEditorThatTakesEveryKeyIncludingTheOnesThatAreShortcuts(t *testing.T) {
	l, ws := newSettingsLab(t, 140, 34)
	l.press("7", "left", "down", "down", "down", "enter")
	l.walkTo(t, "release.branch")
	l.press("enter")
	if !l.sh().editing() {
		t.Fatal("the field has the keyboard")
	}
	l.typed("q7yt").keys("esc")
	if l.quit || l.sh().tab != commands.TabSettings {
		t.Fatal("q and 7 are text inside a field")
	}
	if got := tomlOf(t, ws.dir); got != settingsToml {
		t.Fatalf("Esc cancels without writing:\n%s", got)
	}
	l.press("enter").typeText("trunk").keys("enter")
	want := cliTwin(t, configedit.Command(), "set", "release.branch", "trunk")
	if got := tomlOf(t, ws.dir); got != want {
		t.Fatalf("tab:\n%s\ncommand:\n%s", got, want)
	}
}

func TestAnInvalidValueIsRefusedInTheEditorWithTheSchemasReason(t *testing.T) {
	l, ws := newSettingsLab(t, 140, 34)
	l.press("7", "left", "down", "down", "down", "enter")
	l.walkTo(t, "cache.max_size_mb")
	l.press("enter").typeText("lots").keys("enter")
	l.has("cache.max_size_mb is a whole number")
	if !l.sh().editing() || tomlOf(t, ws.dir) != settingsToml {
		t.Fatal("the editor stays open and the file is untouched")
	}
}

func TestChoosingAThemeInTheStyleSectionRedrawsAtOnceAndKeepsItForYouAlone(t *testing.T) {
	l, ws := newSettingsLab(t, 140, 34)
	l.press("7", "left", "down", "down", "right") // Style
	l.has("Theme", "Symbols", "Preview")
	l.press("down", "down", "enter") // light
	if got := l.sh().th.Tokens.Name; got != "light" {
		t.Fatalf("the interface still draws in %q", got)
	}
	p, err := userprefs.Load(ws.home)
	if err != nil || p.Theme != "light" {
		t.Fatalf("saved %+v, %v", p, err)
	}
	l.press("down", "down", "down", "down", "enter") // symbols: ascii
	if l.sh().th.Icons.Passed != "[v]" {
		t.Fatalf("ascii symbols are drawn now: %q", l.sh().th.Icons.Passed)
	}
	if tomlOf(t, ws.dir) != settingsToml {
		t.Fatal("a personal preference never goes into smith.toml")
	}
}

func TestHomePointsANewUserAtTheSettingsTabUntilTheyHaveOpenedIt(t *testing.T) {
	l, ws := newSettingsLab(t, 140, 34)
	l.openTab(commands.TabHome)
	l.has("New here?", "Settings tab: press 7")
	l.press("7")
	l.press("1")
	l.hasNot("New here?")
	if p, _ := userprefs.Load(ws.home); !p.SettingsTipSeen {
		t.Fatal("the next run knows the tip was seen")
	}
}

func TestTheSettingsTabFillsExactlyTheTerminalInEverySectionAndEditor(t *testing.T) {
	sizes := [][2]int{{140, 40}, {100, 30}, {80, 24}, {72, 9}, {72, 12}, {79, 20}, {60, 15}, {40, 10}, {20, 5}, {10, 3}, {1, 1}}
	steps := map[string][]string{
		"plugins":      {},
		"review":       {},
		"architecture": {"left", "down", "right"},
		"style":        {"left", "down", "down", "right"},
		"workspace":    {"left", "down", "down", "down", "right", "down", "down"},
		"editor":       {"left", "down", "down", "down", "right", "down", "down", "enter"},
		"recipes":      {"left", "down", "down", "down", "down", "right"},
		"forms":        {"left", "down", "down", "down", "down", "down", "right"},
	}
	for name, keys := range steps {
		for _, sz := range sizes {
			l, _ := newSettingsLab(t, sz[0], sz[1])
			l.press("7")
			l.press(keys...)
			lines := strings.Split(l.sh().View(), "\n")
			if len(lines) != sz[1] {
				t.Fatalf("%s at %v: %d lines", name, sz, len(lines))
			}
			for i, line := range lines {
				if w := pane.Width(line); w != sz[0] {
					t.Fatalf("%s at %v: line %d is %d cells wide: %q", name, sz, i, w, line)
				}
			}
		}
	}
}

func TestTheTabStripKeepsEveryTabReachableAtTheNarrowestScreen(t *testing.T) {
	for _, w := range []int{72, 80, 100} {
		l, _ := newSettingsLab(t, w, 24)
		strip := strings.Join(l.lines()[:4], "\n")
		if !strings.Contains(strip, "Sett") {
			t.Errorf("width %d: the Settings tab is not in the strip:\n%s", w, strip)
		}
	}
}

func TestTheSettingsTabCarriesNoColourInThePlainTheme(t *testing.T) {
	l, _ := newSettingsLab(t, 120, 30)
	l.press("7")
	raw := l.sh().View()
	if strings.Contains(raw, "\x1b[38;") || strings.Contains(raw, "\x1b[48;") {
		t.Fatal("colour in the plain theme")
	}
	// State is carried by words and symbols, not by colour alone.
	l.has("[x] go", "on")
	l.press("down", "down", "down", "down", "down", "down", "down", "down", "down", "down", "down", "down", "down", "down")
	l.has("Install")
}

func TestTheMouseOpensTheTabPicksASectionAndADoubleClickTurnsAPluginOff(t *testing.T) {
	l, ws := newSettingsLab(t, 140, 34)
	x, y := l.find("Settings")
	l.sendPatiently(click(x+1, y))
	if l.sh().tab != commands.TabSettings {
		t.Fatalf("a click on the tab opens it: %q", l.sh().tab)
	}
	x, y = l.find("Style")
	l.sendPatiently(click(x+1, y))
	l.has("Theme", "Preview")
	x, y = l.find("Plugins")
	l.sendPatiently(click(x+1, y))
	first := firstBuiltin(t)
	x, y = l.find("[x] " + first)
	l.doubleClickPatiently(x+1, y)
	l.has("[ ] " + first)
	if !strings.Contains(tomlOf(t, ws.dir), "enabled = [") {
		t.Fatal("the double click wrote smith.toml")
	}
	l.send(tea.MouseMsg(tea.MouseEvent{X: x + 1, Y: y, Action: tea.MouseActionPress, Button: tea.MouseButtonWheelDown}))
}

func TestAKeyThatDoesNothingOnTheRowSaysWhy(t *testing.T) {
	l, _ := newSettingsLab(t, 140, 34)
	l.press("7", "t")
	l.has("is unavailable: select an external plugin to trust")
	l.press("u")
	l.has("is unavailable: select a setting to reset to its default")
	if l.sh().settings.Modal() {
		t.Fatal("nothing opened")
	}
}

func TestTheHelpListsTheKeysOfTheSettingsTab(t *testing.T) {
	l, _ := newSettingsLab(t, 140, 40)
	l.press("7", "?")
	l.has("Settings view", "Reset a setting to its default", "Review an external plugin and trust it", "Forget that an external plugin is trusted")
}

func TestAFormOfTheFormsSectionOpensTheGuide(t *testing.T) {
	l, _ := newSettingsLab(t, 140, 34)
	l.press("7", "left", "down", "down", "down", "down", "down", "right")
	l.has("Guided forms")
	l.press("enter")
	if l.sh().tab != commands.TabGuide {
		t.Fatalf("on %q", l.sh().tab)
	}
}
