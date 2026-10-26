package settingssource

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/configedit"
	"github.com/monosmith-app/monosmith/apps/smith/pluginhost"
	"github.com/monosmith-app/monosmith/apps/smith/tui/settingsview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/userprefs"
	"github.com/spf13/cobra"
)

const baseToml = `schema = 1

# The name of this workspace.
[workspace]
name = "acme"   # stays
scope = "@acme"

[ci]
providers = ["github"]   # which CI
devcontainer = false

[release]
# the branch releases come from
branch = "main"
`

// twoWorkspaces makes two identical workspaces, one for the tab and one for the CLI.
func twoWorkspaces(t *testing.T) (tab, cli string) {
	t.Helper()
	// A private smith home, so no trust decision of the machine is read or written.
	t.Setenv("SMITH_HOME", t.TempDir())
	t.Setenv("CI", "")
	// smith sets it for its own tasks (smith run test); the commands would follow it to the real workspace.
	t.Setenv("SMITH_WORKSPACE_ROOT", "")
	base := t.TempDir()
	for _, name := range []string{"tab", "cli"} {
		dir := filepath.Join(base, name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		write(t, dir, baseToml)
	}
	return filepath.Join(base, "tab"), filepath.Join(base, "cli")
}

func write(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "smith.toml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func read(t *testing.T, dir string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(dir, "smith.toml"))
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

// cli runs a smith command in dir, as the terminal would.
func cli(t *testing.T, dir string, cmd *cobra.Command, args ...string) {
	t.Helper()
	t.Chdir(dir)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("smith %s: %v\n%s", strings.Join(args, " "), err, out.String())
	}
}

func same(t *testing.T, tab, cli string) {
	t.Helper()
	if a, b := read(t, tab), read(t, cli); a != b {
		t.Fatalf("the tab and the command wrote different files.\ntab:\n%s\ncommand:\n%s", a, b)
	}
}

func TestTurningAPluginOffWritesTheBytesPluginDisableWrites(t *testing.T) {
	tab, cliDir := twoWorkspaces(t)
	b := New(tab, t.TempDir(), func(string) string { return "" })
	if err := b.SetPlugin("python", false); err != nil {
		t.Fatal(err)
	}
	cli(t, cliDir, pluginhost.Command("test"), "disable", "python")
	same(t, tab, cliDir)
	if !strings.Contains(read(t, tab), `enabled = [`) || strings.Contains(read(t, tab), `"python"`) {
		t.Fatalf("python is off:\n%s", read(t, tab))
	}
	// And back on.
	if err := b.SetPlugin("python", true); err != nil {
		t.Fatal(err)
	}
	cli(t, cliDir, pluginhost.Command("test"), "enable", "python")
	same(t, tab, cliDir)
	snap := b.Load()
	for _, p := range snap.Plugins {
		if !p.Enabled {
			t.Fatalf("%s should be on: %+v", p.Name, snap.Plugins)
		}
	}
}

func TestTheLastPluginCannotBeSwitchedOffEitherWay(t *testing.T) {
	tab, _ := twoWorkspaces(t)
	b := New(tab, t.TempDir(), func(string) string { return "" })
	var names []string
	for _, p := range b.Load().Plugins {
		names = append(names, p.Name)
	}
	for _, n := range names[:len(names)-1] {
		if err := b.SetPlugin(n, false); err != nil {
			t.Fatal(err)
		}
	}
	err := b.SetPlugin(names[len(names)-1], false)
	if err == nil || !strings.Contains(err.Error(), "at least one plugin must stay enabled") {
		t.Fatalf("%v", err)
	}
}

func TestChangingAndResettingASettingWritesTheBytesConfigSetAndUnsetWrite(t *testing.T) {
	tab, cliDir := twoWorkspaces(t)
	b := New(tab, t.TempDir(), func(string) string { return "" })
	steps := []struct {
		path   string
		values []string
	}{
		{"ci.devcontainer", []string{"true"}},
		{"ci.providers", []string{"github", "azure"}},
		{"release.branch", []string{"trunk"}},
		{"cache.max_size_mb", []string{"512"}},
		{"ci.branches", []string{"main", "dev"}},
	}
	for _, s := range steps {
		if _, err := b.SetSetting(s.path, s.values); err != nil {
			t.Fatal(err)
		}
		cli(t, cliDir, configedit.Command(), append([]string{"set", s.path}, s.values...)...)
		same(t, tab, cliDir)
	}
	for _, path := range []string{"release.branch", "cache.max_size_mb"} {
		if _, err := b.UnsetSetting(path); err != nil {
			t.Fatal(err)
		}
		cli(t, cliDir, configedit.Command(), "unset", path)
		same(t, tab, cliDir)
	}
	got := read(t, tab)
	for _, comment := range []string{"# The name of this workspace.", "# stays", "# which CI", "# the branch releases come from"} {
		if !strings.Contains(got, comment) {
			t.Errorf("the comment %q was lost:\n%s", comment, got)
		}
	}
}

func TestASettingTheSchemaRefusesIsRefusedWithTheCommandsReasonAndTheFileIsUntouched(t *testing.T) {
	tab, _ := twoWorkspaces(t)
	b := New(tab, t.TempDir(), func(string) string { return "" })
	before := read(t, tab)
	if _, err := b.SetSetting("ci.providers", []string{"bitbucket"}); err == nil || !strings.Contains(err.Error(), "github or azure or gitlab") {
		t.Fatalf("%v", err)
	}
	if _, err := b.SetSetting("cache.max_size_mb", []string{"lots"}); err == nil || !strings.Contains(err.Error(), "whole number") {
		t.Fatalf("%v", err)
	}
	if read(t, tab) != before {
		t.Fatal("a refused change wrote the file")
	}
}

func TestTrustingAnExternalPluginRecordsAndPinsAsPluginTrustDoes(t *testing.T) {
	tab, cliDir := twoWorkspaces(t)
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	withPlugin := baseToml + "\n[plugins.external.widget]\ncommand = ['" + filepath.ToSlash(exe) + "']\n"
	write(t, tab, withPlugin)
	write(t, cliDir, withPlugin)
	b := New(tab, t.TempDir(), func(string) string { return "" })

	snap := b.Load()
	if len(snap.External) != 1 || snap.External[0].Trusted {
		t.Fatalf("widget is declared and not yet trusted: %+v", snap.External)
	}
	review, err := b.Review("widget")
	if err != nil || review.Trusted {
		t.Fatalf("%+v %v", review, err)
	}
	for _, want := range []string{`The external plugin "widget" wants to run on this machine`, "command", "sha256"} {
		if !strings.Contains(review.Text, want) {
			t.Errorf("the review lacks %q:\n%s", want, review.Text)
		}
	}
	if _, err := b.Trust("widget", true); err != nil {
		t.Fatal(err)
	}
	cli(t, cliDir, pluginhost.Command("test"), "trust", "widget", "--yes", "--pin")
	// Each workspace has its own decision in the one smith home; the pinned
	// digest in smith.toml is the same file in both.
	same(t, tab, cliDir)
	if !strings.Contains(read(t, tab), "sha256 =") {
		t.Fatalf("not pinned:\n%s", read(t, tab))
	}
	if snap = b.Load(); !snap.External[0].Trusted || !snap.External[0].Pinned {
		t.Fatalf("trusted and pinned now: %+v", snap.External[0])
	}
	if again, _ := b.Review("widget"); !again.Trusted {
		t.Fatal("the review knows it is trusted")
	}
	if note, err := b.Untrust("widget"); err != nil || !strings.Contains(note, "No longer trusting") {
		t.Fatalf("%q %v", note, err)
	}
	if note, _ := b.Untrust("widget"); !strings.Contains(note, "was not trusted") {
		t.Fatalf("%q", note)
	}
}

func TestABuiltInPluginHasNothingToTrust(t *testing.T) {
	tab, _ := twoWorkspaces(t)
	b := New(tab, t.TempDir(), func(string) string { return "" })
	if _, err := b.Review("node"); err == nil || !strings.Contains(err.Error(), "built-in") {
		t.Fatalf("%v", err)
	}
}

func TestTheSnapshotListsEverySettingConfigListKnowsExceptTheOnePluginsEdits(t *testing.T) {
	tab, _ := twoWorkspaces(t)
	write(t, tab, baseToml+"\n[default_stacks]\nnode = \"node-jest\"\n")
	snap := New(tab, t.TempDir(), func(string) string { return "" }).Load()
	if snap.Err != nil {
		t.Fatal(snap.Err)
	}
	fields, _ := configedit.Fields()
	got := map[string]settingsview.Setting{}
	for _, s := range snap.Settings {
		got[s.Path] = s
	}
	for _, f := range fields {
		if _, ok := got[f.Path]; !ok && f.Path != "plugins.enabled" {
			t.Errorf("%s is missing", f.Path)
		}
	}
	if _, ok := got["plugins.enabled"]; ok {
		t.Error("plugins.enabled belongs to the Plugins section")
	}
	if s := got["default_stacks.node"]; !s.Set || s.Current[0] != "node-jest" {
		t.Fatalf("a default stack that is set is listed: %+v", s)
	}
	if s := got["ci.providers"]; !s.Set || len(s.Current) != 1 || s.Current[0] != "github" || len(s.Choices) != 3 {
		t.Fatalf("an array reads as its items: %+v", s)
	}
	if s := got["release.github_releases"]; s.Set || s.Default != "false" || s.Explanation == "" || strings.Contains(s.Explanation, "Default:") {
		t.Fatalf("a boolean that is not set: %+v", s)
	}
}

func TestAnUnreadableWorkspaceIsReportedNotHidden(t *testing.T) {
	dir := t.TempDir()
	if snap := New(dir, t.TempDir(), func(string) string { return "" }).Load(); snap.Err == nil {
		t.Fatal("no smith.toml here")
	}
}

func TestTheStyleIsKeptInTheSmithHomeAndTheTipOnlyOnce(t *testing.T) {
	tab, _ := twoWorkspaces(t)
	home := t.TempDir()
	b := New(tab, home, func(string) string { return "" })
	if b.TipSeen {
		t.Fatal("a fresh home has not seen the tip")
	}
	if err := b.Style.Save("light", "ascii"); err != nil {
		t.Fatal(err)
	}
	b.MarkTipSeen()
	p, err := userprefs.Load(home)
	if err != nil || p.Theme != "light" || p.Icons != "ascii" || !p.SettingsTipSeen {
		t.Fatalf("%+v %v", p, err)
	}
	again := New(tab, home, func(string) string { return "" })
	if again.Style.Theme != "light" || !again.TipSeen {
		t.Fatalf("a new session reads it back: %+v", again.Style)
	}
	if err := again.Style.Save("", ""); err != nil {
		t.Fatal(err)
	}
	if p, _ = userprefs.Load(home); !p.SettingsTipSeen {
		t.Fatal("saving the style must not forget the tip")
	}
	if read(t, tab) != baseToml {
		t.Fatal("the style never goes into smith.toml")
	}
}
