package shellcompletion

import (
	"github.com/monosmith-app/monosmith/apps/smith/shellenv"
	"path/filepath"
	"testing"
)

func TestAFileThatThePackageShipsIsReportedAndSmithsOwnIsNot(t *testing.T) {
	env := newTempEnv(t, "linux")
	prefix := filepath.Join(env.home, "brew")
	env.vars["HOMEBREW_PREFIX"] = prefix
	shipped := env.put(t, "# shipped by the package\n", "brew", "share", "zsh", "site-functions", "_smith")
	if got := PackagedPath(env, shellenv.Zsh); got != shipped {
		t.Fatalf("got %q, want %q", got, shipped)
	}
	if plan := PlanFor(env, shellenv.Zsh, testScript); plan.PackagedAt != shipped || !plan.Installed() {
		t.Fatalf("plan: %+v", plan)
	}
	if got := PackagedPath(env, shellenv.Fish); got != "" {
		t.Fatalf("fish: %q", got)
	}
	env.put(t, WithBlock("", "x"), "brew", "share", "zsh", "site-functions", "_smith")
	if got := PackagedPath(env, shellenv.Zsh); got != "" {
		t.Fatalf("a file with the markers is smith's own, not the package's: %q", got)
	}
}

func TestWindowsHasNoPackagedCompletion(t *testing.T) {
	env := newTempEnv(t, "windows")
	for _, s := range shellenv.Shells {
		if got := PackagedPath(env, s); got != "" {
			t.Errorf("%s: %q", s, got)
		}
	}
}
