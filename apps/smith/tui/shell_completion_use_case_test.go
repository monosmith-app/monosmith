package tui

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/sources"
)

const completionPlan = "Tab completion for Zsh. Smith writes these files and changes nothing else:\n\n  /home/ada/.zshrc  (smith adds a marked block to your file)\n    # >>> smith completion >>>\n    fpath=(~/.zfunc $fpath); autoload -Uz compinit; compinit\n    # <<< smith completion <<<\n\nTo remove it again, run: smith completion uninstall\nDry run: nothing was changed."

// completionLab is Home of a healthy workspace where tab completion is not set
// up. declined counts how often "Not now" reached the preferences.
func completionLab(t *testing.T) (l *lab, declined *int) {
	t.Helper()
	l = healthyShell(t, false)
	n := 0
	sh := l.sh()
	sh.sources.CompletionWanted = func() bool { return true }
	sh.sources.DeclineCompletion = func() { n++ }
	sh.completionWanted = true
	l.m = sh
	l.answers(map[string]sources.Captured{"completion install --yes --dry-run": {OK: true, Text: completionPlan}})
	return l, &n
}

func TestHomeOffersTabCompletionAsSomethingThatWaits(t *testing.T) {
	l, _ := completionLab(t)
	l.has("All good", "Tab completion is not set up", "[Set up]", "[Not now]")
	if first := l.lines(); !strings.Contains(strings.Join(first, "\n"), "[Build all (Enter)]") {
		t.Fatal("the offer took Enter away from the first item")
	}
}

func TestHomeHasNoOfferWhenTheSourceSaysNo(t *testing.T) {
	l := healthyShell(t, false)
	l.hasNot("Tab completion is not set up")
}

func TestSetUpShowsTheFileAndTheLineAndWritesOnlyAfterYes(t *testing.T) {
	l, _ := completionLab(t)
	l.clickText("[Set up]")
	l.has("Set up tab completion", "/home/ada/.zshrc", "fpath=(~/.zfunc $fpath)", "# >>> smith completion >>>", "smith completion install --yes")
	if len(l.started) != 0 {
		t.Fatalf("nothing may run before the answer: %v", l.started)
	}
	l.keys("esc")
	if len(l.started) != 0 {
		t.Fatalf("Esc must write nothing: %v", l.started)
	}
	l.clickText("[Set up]")
	l.keys("enter")
	if got := l.lastStarted(); got != "completion install --yes" {
		t.Fatalf("started %q", got)
	}
	if l.sh().completionWanted {
		t.Fatal("Home still offers completion after a good setup")
	}
}

func TestNotNowIsRememberedAndTheItemGoes(t *testing.T) {
	l, declined := completionLab(t)
	l.clickText("[Not now]")
	if *declined != 1 || l.sh().completionWanted {
		t.Fatalf("declined %d times, still wanted %v", *declined, l.sh().completionWanted)
	}
	l.hasNot("Tab completion is not set up")
	if len(l.started) != 0 || len(l.captured) != 0 {
		t.Fatalf("Not now must run nothing: %v %v", l.started, l.captured)
	}
}

func TestSetUpTabCompletionIsInThePalette(t *testing.T) {
	l, _ := completionLab(t)
	l.keys("ctrl+p").typeText("Set up tab").has("Set up tab completion")
	l.keys("enter")
	l.has("smith completion install --yes", "/home/ada/.zshrc")
}
