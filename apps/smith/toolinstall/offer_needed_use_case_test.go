package toolinstall

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolcatalog"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// lintStatus is golangci-lint at a version, as the Go plugin declares it (CI pins 2.15.0).
func lintStatus(state State, version string) Status {
	t, _ := toolcatalog.Find("golangci-lint")
	return Status{Need: Need{ToolNeed: t}, State: state, Version: version, Program: "golangci-lint",
		Wanted: Requirement{Prefix: "2.15.0", Min: "2.15.0"}}
}

func TestExactToolDiffersWhenOlderOrNewer(t *testing.T) {
	tool, _ := toolcatalog.Find("golangci-lint")
	need := Need{ToolNeed: tool}
	for version, want := range map[string]State{"2.15.0": Found, "2.13.0": Outdated, "2.16.0": Outdated, "1.64.8": Outdated} {
		got := DetectAll(context.Background(), fakeDetector{"golangci-lint": {Program: "golangci-lint", Version: version}}, []Need{need})[0]
		if got.State != want {
			t.Errorf("%s: state %s, want %s", version, got.State, want)
		}
	}
}

func TestProblemWordsAVersionDifference(t *testing.T) {
	if got := lintStatus(Outdated, "2.13.0").Problem(); got != "is 2.13.0, not the 2.15.0 that CI pins" {
		t.Fatalf("got %q", got)
	}
}

func TestPlanForADifferentVersionInstallsThePinAndOffersToKeepMine(t *testing.T) {
	plan := BuildPlan(lintStatus(Outdated, "2.13.0"), DetectManagers("linux", onPath("go")), Env{OS: "linux"})
	opt, ok := plan.Find("install:go")
	if !ok || strings.Join(opt.Command, " ") != "go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0" || opt.Admin {
		t.Fatalf("the install is go install of the pinned version, no admin: %+v", opt)
	}
	keep, ok := plan.Find("skip")
	if !ok || keep.Label != "Keep mine" || !strings.Contains(keep.Hint, "2.13.0") {
		t.Fatalf("skip reads as keep mine: %+v", keep)
	}
}

func TestOfferTextNamesEveryToolAndTheOneCommand(t *testing.T) {
	got := OfferText([]Status{missing("python"), lintStatus(Outdated, "2.13.0"), {Need: missing("git").Need, State: Found}})
	for _, want := range []string{"! Python is not installed", "! golangci-lint is 2.13.0, not the 2.15.0 that CI pins", "→ Run smith tools install python golangci-lint to choose how to get them."} {
		if !strings.Contains(got, want) {
			t.Errorf("lacks %q:\n%s", want, got)
		}
	}
	if OfferText([]Status{{Need: missing("git").Need, State: Found}}) != "" {
		t.Fatal("nothing lacking, nothing said")
	}
}

func TestOfferWithoutAQuestionInstallsNothingAndPrintsTheCommand(t *testing.T) {
	r := &fakeRunner{}
	var out bytes.Buffer
	got := offerFor(context.Background(), []Status{lintStatus(Outdated, "2.13.0")}, OfferOptions{Out: &out, Ask: false}, r, afterInstall{runner: r})
	if len(r.ran) != 0 || len(got) != 1 {
		t.Fatalf("nothing runs without a person: ran %v", r.ran)
	}
	if !strings.Contains(out.String(), "smith tools install golangci-lint") {
		t.Fatalf("the output names the command:\n%s", out.String())
	}
}

func TestOfferWithAQuestionPresentsTheChoiceAndRunsOnlyWhatWasChosen(t *testing.T) {
	r := &fakeRunner{}
	var out bytes.Buffer
	// "skip" is the last option: choosing it must run nothing.
	offerFor(context.Background(), []Status{lintStatus(Outdated, "2.13.0")}, OfferOptions{Out: &out, In: strings.NewReader("skip\n"), Ask: true}, r, afterInstall{runner: r})
	if len(r.ran) != 0 {
		t.Fatalf("keep mine runs nothing: %v", r.ran)
	}
	for _, want := range []string{"Keep mine", "Nothing is installed before you choose"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("the choice lacks %q:\n%s", want, out.String())
		}
	}
}

// envRunner is a fake runner that also takes environment entries, and records them.
type envRunner struct {
	fakeRunner
	env []string
}

func (e *envRunner) RunEnv(ctx context.Context, argv, env []string, out io.Writer) error {
	e.env = env
	return e.Run(ctx, argv, out)
}

func TestPlanInstallsThePinIntoSmithsOwnFolderFirstAndLeavesTheOtherCopyAlone(t *testing.T) {
	root := t.TempDir()
	st := lintStatus(Outdated, "2.16.0")
	st.Dirs = []string{root}
	plan := BuildPlan(st, DetectManagers("linux", onPath("go")), Env{OS: "linux"})
	if plan.Default != "install:go-managed" || plan.Options[0].Value != "install:go-managed" {
		t.Fatalf("smith's own folder is the recommended first answer: %v default %s", optionValues(plan), plan.Default)
	}
	managed := plan.Options[0]
	if managed.Env[0] != "GOBIN="+toolrun.ManagedBinDir(root) || !strings.HasSuffix(strings.Join(managed.Command, " "), "@v2.15.0") || managed.Admin {
		t.Fatalf("go install of the pin with GOBIN in .smith/bin: %+v", managed)
	}
	global, _ := plan.Find("install:go")
	if len(global.Env) != 0 || !strings.Contains(global.Hint, "may stay behind another copy on PATH") {
		t.Fatalf("the plain go install stays, with its warning: %+v", global)
	}
	if keep, _ := plan.Find("skip"); keep.Label != "Keep mine" {
		t.Fatalf("keep mine stays: %+v", keep)
	}
}

func TestChoosingSmithsOwnFolderRunsGoInstallWithGOBINAndNothingElse(t *testing.T) {
	root := t.TempDir()
	st := lintStatus(Outdated, "2.16.0")
	st.Dirs = []string{root}
	plan := BuildPlan(st, DetectManagers("linux", onPath("go")), Env{OS: "linux"})
	r := &envRunner{}
	var out bytes.Buffer
	f := Flow{In: bufio.NewReader(strings.NewReader("")), Out: &out, Interactive: true, Yes: true, Runner: r, Detector: fakeDetector{"golangci-lint": {Program: "golangci-lint", Version: "2.15.0"}}}
	outcomes, err := f.Run(context.Background(), []ToolPlan{plan})
	if err != nil || len(outcomes) != 1 || outcomes[0].State != Found {
		t.Fatalf("outcomes %+v, err %v\n%s", outcomes, err, out.String())
	}
	if len(r.ran) != 1 || !strings.HasPrefix(r.ran[0], "go install ") || len(r.env) != 1 || r.env[0] != "GOBIN="+toolrun.ManagedBinDir(root) {
		t.Fatalf("ran %v with %v", r.ran, r.env)
	}
	if !strings.Contains(out.String(), "Running: GOBIN=") {
		t.Fatalf("the output shows the environment of the command:\n%s", out.String())
	}
}

func TestOfferSaysNothingWhenEveryToolIsThere(t *testing.T) {
	var out bytes.Buffer
	if got := offerFor(context.Background(), []Status{lintStatus(Found, "2.15.0")}, OfferOptions{Out: &out, Ask: true}, &fakeRunner{}, nil); got != nil || out.Len() != 0 {
		t.Fatalf("got %v, %q", got, out.String())
	}
}
