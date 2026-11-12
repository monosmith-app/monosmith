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
