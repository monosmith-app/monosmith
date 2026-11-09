package shellcompletion

import (
	"github.com/monosmith-app/monosmith/apps/smith/shellenv"
	"path/filepath"
	"strings"
	"testing"
)

const testScript = "# completion script\nline two\nline three\nline four\n"

func TestFishNeedsOnlyItsOwnFileAndUninstallDeletesIt(t *testing.T) {
	env := newTempEnv(t, "linux")
	plan := PlanFor(env, shellenv.Fish, testScript)
	changes, err := Install(env, plan)
	if err != nil || len(changes) != 1 || changes[0].Action != "wrote" || !strings.Contains(env.read(plan.Targets[0].Path), "line four") {
		t.Fatalf("%+v, %v", changes, err)
	}
	if _, err := Uninstall(env, shellenv.Fish); err != nil || env.Exists(plan.Targets[0].Path) {
		t.Fatalf("file left behind (%v)", err)
	}
}

func TestBashGetsItsScriptInSmithsFolderAndOneLineInBashrc(t *testing.T) {
	env := newTempEnv(t, "linux")
	original := "export A=1\n"
	rc := env.put(t, original, ".bashrc")
	plan := PlanFor(env, shellenv.Bash, testScript)
	if _, err := Install(env, plan); err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(env.home, ".smith", "completions", "smith.bash")
	if !strings.Contains(env.read(script), "line four") {
		t.Fatal("the script is not in the folder of smith")
	}
	if got := env.read(rc); strings.Contains(got, "line four") || !strings.Contains(got, bashLine(script)) || strings.Count(got, "\n") > 8 {
		t.Fatalf("the script, or more than one line, went into .bashrc:\n%s", got)
	}
	if _, err := Uninstall(env, shellenv.Bash); err != nil || env.read(rc) != original || env.Exists(script) {
		t.Fatalf("not removed exactly (%v): %q", err, env.read(rc))
	}
}

func TestAStaleScriptIsRewrittenAloneAndTheProfileIsNotTouched(t *testing.T) {
	env := newTempEnv(t, "linux")
	if _, err := Install(env, PlanFor(env, shellenv.Zsh, "script of another smith")); err != nil {
		t.Fatal(err)
	}
	rc := filepath.Join(env.home, ".zshrc")
	before := env.read(rc)
	plan := PlanFor(env, shellenv.Zsh, testScript)
	if plan.Targets[0].State != Outdated || plan.Targets[1].State != Current || plan.Installed() {
		t.Fatalf("states %v %v", plan.Targets[0].State, plan.Targets[1].State)
	}
	changes, err := Install(env, plan)
	if err != nil || len(changes) != 2 || changes[0].Action != "updated" || changes[1].Action != "unchanged" {
		t.Fatalf("%+v, %v", changes, err)
	}
	if env.read(rc) != before || !strings.Contains(env.read(plan.Targets[0].Path), "line four") {
		t.Fatal("the profile changed, or the script did not")
	}
}

func TestASecondInstallChangesNothingAndAddsNoSecondBlock(t *testing.T) {
	env := newTempEnv(t, "linux")
	rc := env.put(t, "export A=1\n", ".zshrc")
	first, err := Install(env, PlanFor(env, shellenv.Zsh, testScript))
	if err != nil {
		t.Fatal(err)
	}
	after := env.read(rc)
	plan := PlanFor(env, shellenv.Zsh, testScript)
	if !plan.Installed() {
		t.Fatalf("plan after install says not installed: %+v", plan.Targets)
	}
	second, err := Install(env, plan)
	if err != nil || env.read(rc) != after || strings.Count(after, beginMarker) != 1 {
		t.Fatalf("second run changed the file or doubled the block: %v\n%s", err, after)
	}
	for _, c := range second {
		if c.Action != "unchanged" {
			t.Errorf("second run: %+v", c)
		}
	}
	if len(first) != 2 || first[0].Action != "wrote" {
		t.Errorf("first run: %+v", first)
	}
}

func TestUninstallRestoresTheFileOfThePersonExactly(t *testing.T) {
	env := newTempEnv(t, "linux")
	original := "export A=1\nalias ll='ls -l'\n"
	rc := env.put(t, original, ".zshrc")
	if _, err := Install(env, PlanFor(env, shellenv.Zsh, testScript)); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(env.read(rc), zshLine(filepath.Join(env.home, ".smith", "completions", "smith.zsh"))) {
		t.Fatalf("the zsh line is not in .zshrc:\n%s", env.read(rc))
	}
	changes, err := Uninstall(env, shellenv.Zsh)
	if err != nil || len(changes) != 2 {
		t.Fatalf("%+v, %v", changes, err)
	}
	if got := env.read(rc); got != original {
		t.Fatalf(".zshrc not restored: %q", got)
	}
	if env.Exists(filepath.Join(env.home, ".smith", "completions", "smith.zsh")) {
		t.Fatal("the script file is still there")
	}
}

func TestUninstallRemovesAProfileThatInstallCreated(t *testing.T) {
	env := newTempEnv(t, "linux")
	plan := PlanFor(env, shellenv.PowerShell, testScript)
	if _, err := Install(env, plan); err != nil {
		t.Fatal(err)
	}
	if _, err := Uninstall(env, shellenv.PowerShell); err != nil || env.Exists(plan.Targets[0].Path) {
		t.Fatalf("profile left behind: %v", err)
	}
}

func TestInstallLeavesAFileThatSmithDidNotWriteAlone(t *testing.T) {
	env := newTempEnv(t, "linux")
	path := env.put(t, "# somebody else's completion\n", ".smith", "completions", "smith.bash")
	plan := PlanFor(env, shellenv.Bash, testScript)
	if _, blocked := plan.Blocked(); !blocked {
		t.Fatal("the file is not reported as foreign")
	}
	if _, err := Install(env, plan); err == nil {
		t.Fatal("install overwrote a file it did not write")
	}
	if env.read(path) != "# somebody else's completion\n" {
		t.Fatal("the foreign file changed")
	}
	if changes, _ := Uninstall(env, shellenv.Bash); len(changes) != 0 || !env.Exists(path) {
		t.Fatal("uninstall touched a foreign file")
	}
}

func TestAnOutdatedBlockIsUpdatedInPlace(t *testing.T) {
	env := newTempEnv(t, "linux")
	if _, err := Install(env, PlanFor(env, shellenv.Bash, "old script")); err != nil {
		t.Fatal(err)
	}
	plan := PlanFor(env, shellenv.Bash, "new script")
	if plan.Targets[0].State != Outdated || plan.Installed() {
		t.Fatalf("state %v", plan.Targets[0].State)
	}
	changes, err := Install(env, plan)
	if err != nil || changes[0].Action != "updated" || strings.Contains(env.read(plan.Targets[0].Path), "old script") {
		t.Fatalf("%+v, %v", changes, err)
	}
}

func TestPowerShellOnWindowsWritesOneLineToBothProfilesAndTheScriptOnce(t *testing.T) {
	env := newTempEnv(t, "windows")
	env.put(t, "Set-Alias g git\r\n", "Documents", "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1")
	plan := PlanFor(env, shellenv.PowerShell, testScript)
	if len(plan.Targets) != 3 {
		t.Fatalf("%+v", plan.Targets)
	}
	if _, err := Install(env, plan); err != nil {
		t.Fatal(err)
	}
	for _, tg := range plan.Targets {
		if !HasBlock(env.read(tg.Path)) {
			t.Errorf("no block in %s", tg.Path)
		}
	}
	if got := env.read(plan.Targets[1].Path); !strings.HasPrefix(got, "Set-Alias g git\r\n") {
		t.Errorf("the profile of the person changed: %q", got)
	}
}

func TestInstalledListsTheShellsThatHoldABlock(t *testing.T) {
	env := newTempEnv(t, "linux")
	if _, err := Install(env, PlanFor(env, shellenv.Fish, testScript)); err != nil {
		t.Fatal(err)
	}
	if got := Installed(env); len(got) != 1 || got[0] != shellenv.Fish {
		t.Fatalf("%v", got)
	}
}

func TestRenderPlanListsEachFileAndTheExactLines(t *testing.T) {
	env := newTempEnv(t, "linux")
	text := RenderPlan(PlanFor(env, shellenv.Zsh, testScript))
	for _, want := range []string{filepath.Join(env.home, ".smith", "completions", "smith.zsh"), filepath.Join(env.home, ".zshrc"), zshLine(filepath.Join(env.home, ".smith", "completions", "smith.zsh")), beginMarker, endMarker, "the completion script, 4 lines", "smith completion uninstall"} {
		if !strings.Contains(text, want) {
			t.Errorf("plan lacks %q:\n%s", want, text)
		}
	}
}
