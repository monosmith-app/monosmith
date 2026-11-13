package tui

import (
	"context"
	"errors"
	"io"
	"sync"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/depsrepair"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// repairFake answers the repair sources: an offer of node without the cache
// kind, and a record of each repair asked for.
type repairFake struct {
	mu   sync.Mutex
	ran  []protocol.RepairKind
	fail error
}

func repairOffer() depsrepair.Offer {
	return depsrepair.Offer{Root: "/w", Tools: []depsrepair.Tool{{
		Plugin:   "node",
		Projects: []protocol.Project{{Name: "web", Root: "apps/web"}},
		Plans: []protocol.RepairPlan{
			{Kind: protocol.RepairPrune, Summary: "remove what no package.json declares", Commands: []protocol.RepairCommand{{Argv: []string{"npm", "prune"}}}},
			{Kind: protocol.RepairReinstall, Summary: "delete every node_modules, then npm install", Delete: []string{"node_modules"}, Install: true},
			{Kind: protocol.RepairLockfile, Summary: "delete node_modules and package-lock.json", Delete: []string{"node_modules", "package-lock.json"}, Install: true, ChangesVersions: true},
		},
	}}}
}

func (f *repairFake) install(l *lab) {
	sh := l.sh()
	sh.sources.RepairOffer = func() (depsrepair.Offer, error) { return repairOffer(), nil }
	sh.sources.RepairDeps = func(_ context.Context, kind protocol.RepairKind, out io.Writer) error {
		f.mu.Lock()
		f.ran = append(f.ran, kind)
		err := f.fail
		f.mu.Unlock()
		for _, st := range depsrepair.Steps(repairOffer(), kind) {
			_, _ = io.WriteString(out, "… "+st.Name+"\n")
		}
		return err
	}
	l.m = sh
}

func (f *repairFake) kinds() []protocol.RepairKind {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]protocol.RepairKind(nil), f.ran...)
}

// H on the Deps tab opens the walk: every kind with what it does, none marked
// as suggested, a kind no tool has disabled; then the plan, then the repair.
func TestHOnTheDepsTabWalksTheDependencyRepair(t *testing.T) {
	l := newHomeLab(t, 160, 48)
	f := &repairFake{}
	f.install(l)
	l.openTab(commands.TabDeps)
	l.has("Repair dependencies")
	l.keys("H")
	l.has("Which repair?", "1) Prune: remove the installed packages", "3) Rebuild the lockfile", "Clear the download cache", "no tool of this workspace has it")
	l.hasNot("(suggested)")

	l.keys("3")
	l.has("node (1 project): delete node_modules and package-lock.json", "delete   package-lock.json", "The versions can change")
	if len(f.kinds()) != 0 {
		t.Fatal("nothing runs before the confirmation")
	}
	l.keys("esc")
	l.has("Which repair?")
	l.keys("2", "enter")
	if l.sh().fix != nil {
		t.Fatalf("the dialog closes after the repair:\n%s", l.view())
	}
	if got := f.kinds(); len(got) != 1 || got[0] != protocol.RepairReinstall {
		t.Fatalf("ran %v, want reinstall", got)
	}
	if l.sh().depsRepair != nil {
		t.Fatal("the walk is over")
	}
}

// A kind no tool has cannot be picked.
func TestADisabledRepairKindCannotBePicked(t *testing.T) {
	l := newHomeLab(t, 160, 48)
	f := &repairFake{}
	f.install(l)
	l.openTab(commands.TabDeps)
	l.keys("H", "4")
	l.has("Which repair?")
	if len(f.kinds()) != 0 {
		t.Fatal("a disabled kind ran")
	}
	l.keys("esc")
	if l.sh().fix != nil || l.sh().depsRepair != nil {
		t.Fatal("Esc on the first screen closes the walk")
	}
}

// A repair that fails shows the error and keeps the dialog.
func TestAFailedRepairShowsWhy(t *testing.T) {
	l := newHomeLab(t, 160, 48)
	f := &repairFake{fail: errors.New("npm ERR! EBUSY")}
	f.install(l)
	l.openTab(commands.TabDeps)
	l.keys("H", "1", "enter")
	l.has("EBUSY")
	if l.sh().fix == nil {
		t.Fatal("the dialog stays to show the failure")
	}
}
