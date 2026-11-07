package tui

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/depsview"
)

// Until the registries answer, only the keys that can work stay in the footer.
func TestTheDepsFooterHidesTheKeysThatCannotWorkBeforeTheRegistriesAnswer(t *testing.T) {
	l := newLab(t, true, 140, 34)
	sh := l.sh()
	sh.tab, sh.deps = commands.TabDeps, depsview.Model{}
	l.m = sh
	footer := l.lines()[len(l.lines())-2]
	for _, banned := range []string{"update (u", "align (v", "mark (", "remove ("} {
		if strings.Contains(footer, banned) {
			t.Fatalf("the footer promises %q before the answer: %q", banned, footer)
		}
	}
	if !strings.Contains(footer, "reload view (r)") {
		t.Fatalf("reload still works: %q", footer)
	}
	l.keys("u")
	l.has("the registries have not answered yet")
}
