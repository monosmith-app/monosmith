package tui

import (
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/choiceselect"
	"github.com/monosmith-app/monosmith/apps/smith/compose"
	"github.com/monosmith-app/monosmith/apps/smith/tui/composer"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// composeLab is a shell with the Compose a stack screen open on a small catalogue.
func composeLab(t *testing.T) *lab {
	t.Helper()
	l := newLab(t, true, 140, 40)
	cat := compose.Build(compose.Source{
		Languages: []compose.Language{{Name: "node", Title: "TypeScript / JavaScript"}},
		Offers: []choiceselect.Offer{
			{Plugin: "eslint", Choice: protocol.Choice{Axis: "lint", Name: "eslint", Title: "ESLint", Language: "node", Roles: []protocol.Role{protocol.RoleToolchain}}},
		},
	})
	sh := l.sh()
	m := composer.New(cat).Place(sh.w, sh.h)
	sh.compose = &m
	l.m = sh
	return l
}

func TestComposeAStackSitsUnderTheRuleLikeEveryDialog(t *testing.T) {
	for _, size := range [][2]int{{140, 40}, {100, 16}, {80, 24}} {
		l := newLab(t, true, size[0], size[1])
		sh := l.sh()
		m := composer.New(compose.Build(compose.Source{})).Within(sh.dialogMargins()).Place(sh.w, sh.h)
		first := sh.mainInner().Y
		// Compose needs 12 rows: on a lower screen it may reach over the footer,
		// never over the rule.
		if r := m.Rect(); r.Y < first || (r.H <= sh.lay().Footer.Y-first && r.Y+r.H > sh.lay().Footer.Y) {
			t.Errorf("%dx%d: Compose spans rows %d-%d, want inside %d-%d", size[0], size[1], r.Y, r.Y+r.H-1, first, sh.lay().Footer.Y-1)
		}
	}
}

func TestAClickInComposeAStackTicksTheBoxUnderIt(t *testing.T) {
	l := composeLab(t)
	l.clickText("[ ] ESLint")
	if got := l.sh().compose.Ticked(); len(got) != 1 || got[0] != "eslint" {
		t.Fatalf("the click ticked %v", got)
	}
	l.has("[x] ESLint")
	l.clickText("[x] ESLint")
	if got := l.sh().compose.Ticked(); len(got) != 0 {
		t.Fatalf("a second click unticks it, ticked %v", got)
	}
}
