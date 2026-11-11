package depsrepair

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// lab runs the command against a fake workspace: no session opens and no
// plugin starts.
type lab struct {
	t           *testing.T
	interactive bool
	answers     string
	offer       Offer
	opened      bool
	rec         *recorder
}

func newLab(t *testing.T) *lab {
	root := t.TempDir()
	offer := nodeOffer(root)
	offer.Tools[0].Plans = append(offer.Tools[0].Plans, protocol.RepairPlan{
		Kind: protocol.RepairCache, Summary: "clear the npm download cache, then npm install",
		Commands: []protocol.RepairCommand{{Argv: []string{"npm", "cache", "clean", "--force"}}}, Install: true, MachineWide: true,
	})
	return &lab{t: t, offer: offer, rec: &recorder{}}
}

func (l *lab) run(args ...string) (string, error) {
	l.t.Helper()
	env := environment{
		interactive: func() bool { return l.interactive },
		in:          strings.NewReader(l.answers),
		open: func(string, string, []string) (Offer, Ports, func(), error) {
			l.opened = true
			return l.offer, l.rec.ports(l.offer.Root), func() {}, nil
		},
	}
	cmd := newCommand("test", env)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	cmd.SetContext(context.Background())
	err := cmd.Execute()
	return out.String(), err
}

func TestWithoutATerminalItRefusesBeforeOpeningTheWorkspace(t *testing.T) {
	for _, args := range [][]string{nil, {"--kind", "reinstall"}, {"--yes"}} {
		l := newLab(t)
		_, err := l.run(args...)
		if err == nil || !strings.Contains(err.Error(), "without a terminal") || !strings.Contains(err.Error(), "→ run `smith deps repair --kind") {
			t.Fatalf("%v: want the refusal with its remedy, got %v", args, err)
		}
		if l.opened {
			t.Fatalf("%v: the workspace was opened before the refusal", args)
		}
	}
}

func TestDryRunPrintsEveryPlanAndChangesNothing(t *testing.T) {
	l := newLab(t)
	out, err := l.run("--dry-run")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(out, "══ Dependency repair ═") {
		t.Fatalf("the banner comes first:\n%s", out)
	}
	for _, want := range []string{"Prune:", "Rebuild the lockfile:", "Clear the download cache:", "delete   package-lock.json", "run      npm cache clean --force", "▲ This clears a cache that every workspace on this machine uses", "→ to repair this workspace only"} {
		if !strings.Contains(out, want) {
			t.Errorf("the dry run does not say %q:\n%s", want, out)
		}
	}
	if !strings.HasSuffix(out, "● Dry run: nothing was deleted, run or installed.\n") {
		t.Fatalf("the closing line says nothing changed:\n%s", out)
	}
	if len(l.rec.did) != 0 {
		t.Fatalf("a dry run did %v", l.rec.did)
	}
}

func TestDryRunJSONHasThePlansOfTheKind(t *testing.T) {
	l := newLab(t)
	out, err := l.run("--dry-run", "--kind", "lockfile", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var r Report
	if err := json.Unmarshal([]byte(out), &r); err != nil {
		t.Fatalf("%v:\n%s", err, out)
	}
	if len(r.Kinds) != 1 || r.Kinds[0].Kind != protocol.RepairLockfile || !r.Kinds[0].ChangesVersions || len(r.Kinds[0].Tools) != 1 {
		t.Fatalf("report: %+v", r)
	}
}

func TestInATerminalItAsksTheKindThenConfirmsThenRepairs(t *testing.T) {
	l := newLab(t)
	l.interactive, l.answers = true, "3\ny\n"
	out, err := l.run()
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	for _, want := range []string{"1) Prune:", "2) Clean reinstall:", "(no tool of this workspace has it)", "══ Dependency repair: Rebuild the lockfile ═", "Repair now? [y/N]", "▲ The versions can change"} {
		if !strings.Contains(out, want) {
			t.Errorf("the walk does not say %q:\n%s", want, out)
		}
	}
	if want := "✓ Repaired the dependencies (rebuild the lockfile) with node. Nothing is committed.\n"; !strings.HasSuffix(out, want) {
		t.Fatalf("closing line: want %q in\n%s", want, out)
	}
	if len(l.rec.did) != 4 {
		t.Fatalf("did %v", l.rec.did)
	}
}

func TestAKindNoToolHasIsRefusedWhenPicked(t *testing.T) {
	l := newLab(t)
	l.interactive, l.answers = true, "2\n"
	_, err := l.run()
	if err == nil || !strings.Contains(err.Error(), "not a repair this workspace has") {
		t.Fatalf("got %v", err)
	}
	if len(l.rec.did) != 0 {
		t.Fatalf("did %v", l.rec.did)
	}
}

func TestNoAtTheConfirmationChangesNothing(t *testing.T) {
	l := newLab(t)
	l.interactive, l.answers = true, "n\n"
	_, err := l.run("--kind", "prune")
	if err == nil || err.Error() != "repair cancelled" || len(l.rec.did) != 0 {
		t.Fatalf("got %v, did %v", err, l.rec.did)
	}
}

func TestKindAndYesRunWithoutATerminal(t *testing.T) {
	l := newLab(t)
	out, err := l.run("--kind", "prune", "--yes")
	if err != nil || len(l.rec.did) != 1 || !strings.Contains(out, "✓ Repaired the dependencies (prune)") {
		t.Fatalf("%v, did %v:\n%s", err, l.rec.did, out)
	}
}

func TestAFailedStepEndsWithItsRemedy(t *testing.T) {
	l := newLab(t)
	l.rec.fail = "run npm prune"
	_, err := l.run("--kind", "prune", "--yes")
	if err == nil || !strings.Contains(err.Error(), "the dependency repair stopped") || !strings.Contains(err.Error(), "→ fix what the tool reported, then run `smith deps repair --kind prune` again") {
		t.Fatalf("got %v", err)
	}
}

func TestAnUnknownKindIsRefused(t *testing.T) {
	l := newLab(t)
	_, err := l.run("--kind", "nuke", "--dry-run")
	if err == nil || !strings.Contains(err.Error(), "prune|reinstall|lockfile|cache") || l.opened {
		t.Fatalf("got %v (opened %v)", err, l.opened)
	}
}

func TestNoToolSaysSoWithTheReason(t *testing.T) {
	l := newLab(t)
	l.offer = Offer{Root: l.offer.Root, Notes: []string{"python could not say how it repairs dependencies: crashed"}}
	_, err := l.run("--dry-run")
	if err == nil || !strings.Contains(err.Error(), "no tool of this workspace offers a dependency repair") || !strings.Contains(err.Error(), "→ python could not say") {
		t.Fatalf("got %v", err)
	}
}

