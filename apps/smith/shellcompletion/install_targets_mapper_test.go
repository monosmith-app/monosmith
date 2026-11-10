package shellcompletion

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/shellenv"
)

func TestTheScriptLivesInSmithsOwnFolderAndTheProfileGetsOneLine(t *testing.T) {
	env := newTempEnv(t, "linux")
	h := env.home
	dir := filepath.Join(h, ".smith", "completions")

	bash := Targets(env, shellenv.Bash, "s")
	if len(bash) != 2 || bash[0].Path != filepath.Join(dir, "smith.bash") || !bash[0].Whole || bash[1].Path != filepath.Join(h, ".bashrc") || bash[1].Whole {
		t.Fatalf("bash: %+v", bash)
	}
	if want := `[ -f "` + filepath.ToSlash(bash[0].Path) + `" ] && source "` + filepath.ToSlash(bash[0].Path) + `"`; bash[1].Body != want {
		t.Errorf("bash line %q, want %q", bash[1].Body, want)
	}

	zsh := Targets(env, shellenv.Zsh, "s")
	if len(zsh) != 2 || zsh[0].Path != filepath.Join(dir, "smith.zsh") || zsh[1].Path != filepath.Join(h, ".zshrc") {
		t.Fatalf("zsh: %+v", zsh)
	}
	if !strings.HasPrefix(zsh[1].Body, "[ -f ") || !strings.Contains(zsh[1].Body, "compinit") || strings.Contains(zsh[1].Body, "\n") {
		t.Errorf("zsh line %q", zsh[1].Body)
	}

	ps := Targets(env, shellenv.PowerShell, "s")
	if len(ps) != 2 || ps[0].Path != filepath.Join(dir, "smith.ps1") || ps[1].Path != filepath.Join(h, ".config", "powershell", "Microsoft.PowerShell_profile.ps1") {
		t.Fatalf("pwsh on Linux: %+v", ps)
	}
	if want := "if (Test-Path '" + ps[0].Path + "') { . '" + ps[0].Path + "' }"; ps[1].Body != want {
		t.Errorf("PowerShell line %q, want %q", ps[1].Body, want)
	}

	fish := Targets(env, shellenv.Fish, "s")
	if len(fish) != 1 || fish[0].Path != filepath.Join(h, ".config", "fish", "completions", "smith.fish") || !fish[0].Whole {
		t.Fatalf("fish needs no profile line: %+v", fish)
	}
}

func TestNoScriptIsEverPutIntoAProfile(t *testing.T) {
	env := newTempEnv(t, "windows")
	for _, s := range shellenv.Shells {
		for _, tg := range Targets(env, s, "# the completion script\nline two\nline three\nline four") {
			if !tg.Whole && strings.Contains(tg.Body, "\n") {
				t.Errorf("%s: the profile %s gets more than one line:\n%s", s, tg.Path, tg.Body)
			}
		}
	}
}

func TestTargetsHonourSmithHomeAndTheXDGAndZDOTDIRVariables(t *testing.T) {
	env := newTempEnv(t, "linux")
	env.vars["SMITH_HOME"], env.vars["XDG_CONFIG_HOME"], env.vars["ZDOTDIR"] = "/sh", "/conf", "/zdot"
	if got := Targets(env, shellenv.Bash, "s")[0].Path; got != filepath.Join("/sh", "completions", "smith.bash") {
		t.Error(got)
	}
	if got := Targets(env, shellenv.Fish, "s")[0].Path; got != filepath.Join("/conf", "fish", "completions", "smith.fish") {
		t.Error(got)
	}
	if got := Targets(env, shellenv.Zsh, "s")[1].Path; got != filepath.Join("/zdot", ".zshrc") {
		t.Error(got)
	}
}

func TestPowerShellOnWindowsHasOneProfileForEachEditionUnlessTheParentSaysWhich(t *testing.T) {
	env := newTempEnv(t, "windows")
	docs := filepath.Join(env.home, "Documents")
	v5 := filepath.Join(docs, "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1")
	v7 := filepath.Join(docs, "PowerShell", "Microsoft.PowerShell_profile.ps1")
	for parent, want := range map[string][]string{"": {v5, v7}, "cmd.exe": {v5, v7}, "powershell.exe": {v5}, "pwsh.exe": {v7}} {
		env.parent = parent
		var paths []string
		for _, tg := range Targets(env, shellenv.PowerShell, "s")[1:] {
			paths = append(paths, tg.Path)
		}
		if len(paths) != len(want) || paths[0] != want[0] || (len(want) == 2 && paths[1] != want[1]) {
			t.Errorf("parent %q: got %v, want %v", parent, paths, want)
		}
	}
}

func TestPowerShellFollowsADocumentsFolderThatOneDriveMoved(t *testing.T) {
	env := newTempEnv(t, "windows")
	env.put(t, "", "OneDrive", "Documents", "keep")
	env.parent = "pwsh.exe"
	want := filepath.Join(env.home, "OneDrive", "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1")
	if got := Targets(env, shellenv.PowerShell, "s"); len(got) != 2 || got[1].Path != want {
		t.Fatalf("%+v", got)
	}
}

func TestWithoutAHomeThereAreNoTargets(t *testing.T) {
	env := newTempEnv(t, "linux")
	env.home = ""
	if got := Targets(env, shellenv.Bash, "s"); got != nil {
		t.Fatal(got)
	}
}
