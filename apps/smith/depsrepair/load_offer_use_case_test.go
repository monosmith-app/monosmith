package depsrepair

import (
	"errors"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func answering(plugin string, installs bool, plans ...protocol.RepairPlan) asker {
	return asker{plugin: plugin, installs: installs, projects: []protocol.Project{{Name: plugin + "-app", Root: "apps/" + plugin}},
		ask: func(protocol.RepairParams) (protocol.RepairResult, error) { return protocol.RepairResult{Plans: plans}, nil }}
}

func TestGatherKeepsGoodPlansInKindOrder(t *testing.T) {
	offer := gather(t.TempDir(), nil, []asker{answering("node", true,
		protocol.RepairPlan{Kind: protocol.RepairCache, Summary: "c", Commands: []protocol.RepairCommand{{Argv: []string{"npm", "cache", "clean", "--force"}}}, Install: true, MachineWide: true},
		protocol.RepairPlan{Kind: protocol.RepairPrune, Summary: "p", Commands: []protocol.RepairCommand{{Argv: []string{"npm", "prune"}}}},
	)})
	if len(offer.Tools) != 1 || len(offer.Notes) != 0 {
		t.Fatalf("one tool, no note: %+v", offer)
	}
	kinds := offer.Kinds()
	if len(kinds) != 2 || kinds[0] != protocol.RepairPrune || kinds[1] != protocol.RepairCache {
		t.Fatalf("lightest first: %v", kinds)
	}
	if !offer.MachineWide(protocol.RepairCache) || offer.MachineWide(protocol.RepairPrune) {
		t.Fatal("only the cache kind is machine-wide")
	}
}

func TestGatherRefusesBadPlansWithANote(t *testing.T) {
	offer := gather(t.TempDir(), []string{"apps/node"}, []asker{answering("node", false,
		protocol.RepairPlan{Kind: protocol.RepairPrune, Summary: "nothing"},
		protocol.RepairPlan{Kind: protocol.RepairReinstall, Summary: "r", Delete: []string{"../outside"}},
		protocol.RepairPlan{Kind: protocol.RepairLockfile, Summary: "l", Delete: []string{"package-lock.json"}, Install: true},
		protocol.RepairPlan{Kind: protocol.RepairCache, Summary: "c", Delete: []string{"apps/node"}},
		protocol.RepairPlan{Kind: "nuke", Summary: "n", Delete: []string{"x"}},
	)})
	if len(offer.Tools) != 0 {
		t.Fatalf("every plan is refused: %+v", offer.Tools)
	}
	text := strings.Join(offer.Notes, "\n")
	for _, want := range []string{"deletes nothing", "not a path inside the workspace", "cannot install", "never source", `"nuke"`} {
		if !strings.Contains(text, want) {
			t.Errorf("no note says %q:\n%s", want, text)
		}
	}
}

func TestGatherKeepsTheOtherToolsWhenOneFails(t *testing.T) {
	broken := asker{plugin: "python", ask: func(protocol.RepairParams) (protocol.RepairResult, error) { return protocol.RepairResult{}, errors.New("crashed") }}
	offer := gather(t.TempDir(), nil, []asker{broken, answering("node", true, protocol.RepairPlan{Kind: protocol.RepairPrune, Summary: "p", Commands: []protocol.RepairCommand{{Argv: []string{"npm", "prune"}}}})})
	if len(offer.Tools) != 1 || offer.Tools[0].Plugin != "node" {
		t.Fatalf("node is still offered: %+v", offer.Tools)
	}
	if len(offer.Notes) != 1 || !strings.Contains(offer.Notes[0], "python") {
		t.Fatalf("a note names python: %v", offer.Notes)
	}
}

func TestGatherUsesTheFirstPlanOfAKind(t *testing.T) {
	offer := gather(t.TempDir(), nil, []asker{answering("node", true,
		protocol.RepairPlan{Kind: protocol.RepairPrune, Summary: "first", Commands: []protocol.RepairCommand{{Argv: []string{"npm", "prune"}}}},
		protocol.RepairPlan{Kind: protocol.RepairPrune, Summary: "second", Commands: []protocol.RepairCommand{{Argv: []string{"npm", "prune"}}}},
	)})
	if p, _ := offer.Tools[0].Plan(protocol.RepairPrune); p.Summary != "first" || len(offer.Notes) != 1 {
		t.Fatalf("the first plan wins, with a note: %+v %v", p, offer.Notes)
	}
}
