package depsrepair

import (
	"context"
	"errors"
	"io"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// recorder is ports that write down what they were asked to do.
type recorder struct {
	did  []string
	fail string
}

func (r *recorder) ports(root string) Ports {
	rel := func(p string) string { s, _ := filepath.Rel(root, p); return filepath.ToSlash(s) }
	act := func(what string) error {
		r.did = append(r.did, what)
		if what == r.fail {
			return errors.New("boom")
		}
		return nil
	}
	return Ports{
		Delete: func(_ context.Context, path string) error { return act("delete " + rel(path)) },
		Run: func(_ context.Context, _ string, step protocol.FixStep, out io.Writer) error {
			_, _ = io.WriteString(out,"… "+step.Description+"\n")
			if err := act("run " + strings.Join(step.Argv, " ")); err != nil {
				return &doctorfix.StepError{Number: 1, Total: 1, Step: step, Err: err, Output: "npm ERR! boom"}
			}
			return nil
		},
		Install: func(_ context.Context, tool Tool, _ io.Writer) error { return act("install " + tool.Plugin) },
	}
}

func nodeOffer(root string) Offer {
	return Offer{Root: root, ProjectRoots: []string{"apps/web"}, Tools: []Tool{{
		Plugin:   "node",
		Projects: []protocol.Project{{Name: "web", Root: "apps/web"}},
		Plans: []protocol.RepairPlan{
			{Kind: protocol.RepairPrune, Summary: "prune", Commands: []protocol.RepairCommand{{Argv: []string{"npm", "prune"}}}},
			{Kind: protocol.RepairLockfile, Summary: "lock", Delete: []string{"node_modules", "package-lock.json"}, Commands: []protocol.RepairCommand{{Argv: []string{"npm", "cache", "verify"}}}, Install: true, ChangesVersions: true},
		},
	}}}
}

func TestApplyDeletesThenRunsThenInstalls(t *testing.T) {
	root := t.TempDir()
	r := &recorder{}
	var progress strings.Builder
	if err := Apply(context.Background(), nodeOffer(root), protocol.RepairLockfile, r.ports(root), &progress); err != nil {
		t.Fatal(err)
	}
	want := []string{"delete node_modules", "delete package-lock.json", "run npm cache verify", "install node"}
	if !reflect.DeepEqual(r.did, want) {
		t.Fatalf("did %v, want %v", r.did, want)
	}
	if n := strings.Count(progress.String(), "… "); n != 4 {
		t.Fatalf("each step says it starts once, got %d in\n%s", n, progress.String())
	}
}

func TestApplyStopsAtTheFirstFailureAndNamesTheStep(t *testing.T) {
	root := t.TempDir()
	r := &recorder{fail: "run npm cache verify"}
	err := Apply(context.Background(), nodeOffer(root), protocol.RepairLockfile, r.ports(root), io.Discard)
	var step *doctorfix.StepError
	if !errors.As(err, &step) {
		t.Fatalf("want a step error, got %v", err)
	}
	if step.Number != 3 || step.Total != 4 || step.Output != "npm ERR! boom" {
		t.Fatalf("step %d of %d, output %q", step.Number, step.Total, step.Output)
	}
	if len(r.did) != 3 {
		t.Fatalf("the install after the failure ran: %v", r.did)
	}
}

func TestApplyRefusesAKindNoToolHas(t *testing.T) {
	root := t.TempDir()
	r := &recorder{}
	if err := Apply(context.Background(), nodeOffer(root), protocol.RepairCache, r.ports(root), io.Discard); err == nil || len(r.did) != 0 {
		t.Fatalf("want a refusal and nothing done, got %v, %v", err, r.did)
	}
}

// The disk can change between the plan and the run: each path is checked again.
func TestApplyChecksEachPathAgainBeforeItDeletes(t *testing.T) {
	root := t.TempDir()
	offer := nodeOffer(root)
	offer.Tools[0].Plans[1].Delete = []string{"apps/web"}
	r := &recorder{}
	err := Apply(context.Background(), offer, protocol.RepairLockfile, r.ports(root), io.Discard)
	if err == nil || len(r.did) != 0 {
		t.Fatalf("a project's root is never deleted: %v, %v", err, r.did)
	}
}

func TestStepsNameEachStepByToolAndAction(t *testing.T) {
	var names []string
	for _, s := range Steps(nodeOffer("/w"), protocol.RepairLockfile) {
		names = append(names, s.Name)
	}
	want := []string{"node: delete node_modules", "node: delete package-lock.json", "node: run npm cache verify", "node: install 1 project again"}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("got %v, want %v", names, want)
	}
}
