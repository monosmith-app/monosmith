package completioncommand

import (
	"bytes"
	"github.com/monosmith-app/monosmith/apps/smith/shellcompletion"
	"github.com/monosmith-app/monosmith/apps/smith/shellenv"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// run executes `smith completion <args>` on a small command tree.
func run(t *testing.T, d Dependencies, typed string, args ...string) (string, error) {
	t.Helper()
	root := &cobra.Command{Use: "smith", SilenceUsage: true, SilenceErrors: true}
	root.AddCommand(&cobra.Command{Use: "build", Short: "build things", Run: func(*cobra.Command, []string) {}}, NewCommand(d))
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetIn(strings.NewReader(typed))
	root.SetArgs(append([]string{"completion"}, args...))
	err := root.Execute()
	return out.String(), err
}

func deps(t *testing.T, shellParent string, interactive bool) (Dependencies, *tempEnv, *answerStore) {
	env := newTempEnv(t, "linux")
	env.parent = shellParent
	mem := &answerStore{}
	return Dependencies{Env: env, Memory: mem, Interactive: interactive}, env, mem
}

func TestInstallYesWritesTheFileWithoutAsking(t *testing.T) {
	d, env, mem := deps(t, "fish", false)
	out, err := run(t, d, "", "install", "--yes")
	if err != nil {
		t.Fatal(err)
	}
	path := shellcompletion.Targets(env, shellenv.Fish, "")[0].Path
	if !shellcompletion.HasBlock(env.read(path)) || !strings.Contains(out, "Wrote "+path) || mem.a.Choice != ChoiceInstalled {
		t.Fatalf("%q\n%s", mem.a.Choice, out)
	}
}

func TestInstallWithoutYesAsksAndNoWritesNothing(t *testing.T) {
	d, env, _ := deps(t, "fish", true)
	out, err := run(t, d, "2\n", "install")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Write these files?") || !strings.Contains(out, "Show me the script") || !strings.Contains(out, "Nothing was written.") {
		t.Fatalf("the question or the answer text is missing:\n%s", out)
	}
	if env.Exists(shellcompletion.Targets(env, shellenv.Fish, "")[0].Path) {
		t.Fatal("wrote after No")
	}
}

func TestInstallWithoutAnAnswerOrATerminalFailsAndWritesNothing(t *testing.T) {
	d, env, _ := deps(t, "fish", false)
	if _, err := run(t, d, "", "install"); err == nil || !strings.Contains(err.Error(), "--yes") {
		t.Fatalf("%v", err)
	}
	if env.Exists(shellcompletion.Targets(env, shellenv.Fish, "")[0].Path) {
		t.Fatal("wrote without an answer")
	}
}

func TestInstallDryRunShowsThePlanAndWritesNothing(t *testing.T) {
	d, env, _ := deps(t, "bash", false)
	out, err := run(t, d, "", "install", "--yes", "--dry-run")
	if err != nil || !strings.Contains(out, "smith.bash") || !strings.Contains(out, "Dry run: nothing was changed.") {
		t.Fatalf("%v\n%s", err, out)
	}
	if env.Exists(shellcompletion.Targets(env, shellenv.Bash, "")[0].Path) {
		t.Fatal("a dry run wrote a file")
	}
}

func TestInstallTwiceSaysSoAndKeepsOneBlock(t *testing.T) {
	d, env, _ := deps(t, "bash", false)
	if _, err := run(t, d, "", "install", "--yes"); err != nil {
		t.Fatal(err)
	}
	out, err := run(t, d, "", "install", "--yes")
	if err != nil || !strings.Contains(out, "already set up") {
		t.Fatalf("%v\n%s", err, out)
	}
	if n := strings.Count(env.read(shellcompletion.Targets(env, shellenv.Bash, "")[0].Path), "# >>> smith completion >>>"); n != 1 {
		t.Fatalf("%d blocks", n)
	}
}

func TestInstallShellFlagOverridesDetection(t *testing.T) {
	d, env, _ := deps(t, "bash", false)
	if _, err := run(t, d, "", "install", "--yes", "--shell", "powershell"); err != nil {
		t.Fatal(err)
	}
	if !shellcompletion.HasBlock(env.read(shellcompletion.Targets(env, shellenv.PowerShell, "")[0].Path)) {
		t.Fatal("no PowerShell profile block")
	}
	if _, err := run(t, d, "", "install", "--yes", "--shell", "tcsh"); err == nil {
		t.Fatal("an unknown shell was accepted")
	}
}

func TestInstallSaysSoWhenThePackageShipsTheFile(t *testing.T) {
	d, env, _ := deps(t, "zsh", false)
	env.vars["HOMEBREW_PREFIX"] = env.home
	shipped := env.put(t, "#compdef smith\n", "share", "zsh", "site-functions", "_smith")
	out, err := run(t, d, "", "install", "--yes")
	if err != nil || !strings.Contains(out, shipped) || !strings.Contains(out, "changes nothing") {
		t.Fatalf("%v\n%s", err, out)
	}
	if env.Exists(shellcompletion.Targets(env, shellenv.Zsh, "")[0].Path) {
		t.Fatal("wrote beside a packaged file")
	}
}

func TestUninstallAsksThenRemovesWhatInstallAddedInEveryShell(t *testing.T) {
	d, env, mem := deps(t, "bash", true)
	for _, s := range []string{"bash", "fish"} {
		if _, err := run(t, d, "", "install", "--yes", "--shell", s); err != nil {
			t.Fatal(err)
		}
	}
	if out, _ := run(t, d, "2\n", "uninstall"); !strings.Contains(out, "Nothing was removed.") || !env.Exists(shellcompletion.Targets(env, shellenv.Bash, "")[0].Path) {
		t.Fatalf("No removed something:\n%s", out)
	}
	if _, err := run(t, d, "1\n", "uninstall"); err != nil {
		t.Fatal(err)
	}
	if env.Exists(shellcompletion.Targets(env, shellenv.Bash, "")[0].Path) || env.Exists(shellcompletion.Targets(env, shellenv.Fish, "")[0].Path) {
		t.Fatal("files are left")
	}
	if mem.a.Choice != "" {
		t.Fatalf("the answer was not reset: %+v", mem.a)
	}
	if out, _ := run(t, d, "", "uninstall", "--yes"); !strings.Contains(out, "nothing to remove") {
		t.Fatalf("%s", out)
	}
}

func TestStatusListsEveryShell(t *testing.T) {
	d, _, _ := deps(t, "bash", false)
	if _, err := run(t, d, "", "install", "--yes"); err != nil {
		t.Fatal(err)
	}
	out, err := run(t, d, "", "status")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Bash", "installed (", "Zsh", "not installed", "your shell", "Smith installed the completion script."} {
		if !strings.Contains(out, want) {
			t.Errorf("status lacks %q:\n%s", want, out)
		}
	}
}

func TestStatusJSONTellsAnEditorWhetherToOffer(t *testing.T) {
	d, _, _ := deps(t, "bash", false)
	out, err := run(t, d, "", "status", "--json")
	if err != nil || strings.TrimSpace(out) != `{"shell":"bash","wanted":true,"answer":"","stale":false}` {
		t.Fatalf("%v\n%s", err, out)
	}
	if _, err := run(t, d, "", "install", "--yes"); err != nil {
		t.Fatal(err)
	}
	out, _ = run(t, d, "", "status", "--json")
	if strings.TrimSpace(out) != `{"shell":"bash","wanted":false,"answer":"installed","stale":false}` {
		t.Fatalf("%s", out)
	}
}

func TestTheScriptSubcommandsStillPrintScripts(t *testing.T) {
	d, _, _ := deps(t, "bash", false)
	for _, s := range shellenv.Shells {
		out, err := run(t, d, "", string(s))
		if err != nil || !strings.Contains(out, "smith") || len(out) < 200 {
			t.Errorf("%s: %v (%d bytes)", s, err, len(out))
		}
	}
}
