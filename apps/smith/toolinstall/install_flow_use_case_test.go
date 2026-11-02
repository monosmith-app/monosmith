package toolinstall

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
)

// fakeRunner records the commands it was asked to run and runs none.
type fakeRunner struct {
	ran []string
	err error
}

func (f *fakeRunner) Run(_ context.Context, argv []string, out io.Writer) error {
	f.ran = append(f.ran, strings.Join(argv, " "))
	_, _ = io.WriteString(out, "fake output\n")
	return f.err
}

// afterInstall answers "found" for a tool once the runner ran.
type afterInstall struct {
	runner *fakeRunner
	probe  Probe
}

func (a afterInstall) Detect(context.Context, Need) Probe {
	if len(a.runner.ran) > 0 {
		return a.probe
	}
	return Probe{}
}

func flow(input string, interactive bool) (Flow, *fakeRunner, *bytes.Buffer) {
	r := &fakeRunner{}
	var out bytes.Buffer
	return Flow{
		In: bufio.NewReader(strings.NewReader(input)), Out: &out, Interactive: interactive, Runner: r,
		Detector: afterInstall{runner: r, probe: Probe{Program: "python3", Version: "3.13.1"}},
	}, r, &out
}

func pythonPlan(goos string, have ...string) ToolPlan {
	return BuildPlan(missing("python"), DetectManagers(goos, onPath(have...)), Env{OS: goos})
}

func TestFlowRefusesWithoutATerminalAndNamesTheFlags(t *testing.T) {
	f, r, _ := flow("", false)
	_, err := f.Run(context.Background(), []ToolPlan{pythonPlan("windows", "winget")})
	if err == nil || !strings.Contains(err.Error(), "--yes") || !strings.Contains(err.Error(), "--manager") || !strings.Contains(err.Error(), "--dry-run") {
		t.Fatalf("want a refusal that names the flags, got %v", err)
	}
	if len(r.ran) != 0 {
		t.Fatalf("nothing may run: %v", r.ran)
	}
}

func TestFlowRunsTheChosenCommandAfterPrintingIt(t *testing.T) {
	f, r, out := flow("\n", true) // Enter takes the marked default
	got, err := f.Run(context.Background(), []ToolPlan{pythonPlan("darwin", "brew")})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.ran) != 1 || r.ran[0] != "brew install python@3.13" {
		t.Fatalf("ran %v", r.ran)
	}
	text := out.String()
	if strings.Index(text, "Running: brew install python@3.13") > strings.Index(text, "fake output") {
		t.Fatalf("the command must print before its output:\n%s", text)
	}
	if got[0].State != Found || !strings.Contains(text, "Python is installed (python3 3.13.1)") {
		t.Fatalf("outcome %+v\n%s", got[0], text)
	}
}

func TestFlowSaysWhenANewTerminalIsNeeded(t *testing.T) {
	f, r, out := flow("\n", true)
	f.Detector = afterInstall{runner: r} // still not found after the install
	got, _ := f.Run(context.Background(), []ToolPlan{pythonPlan("darwin", "brew")})
	if got[0].State != Missing || !strings.Contains(out.String(), "Open a new terminal") {
		t.Fatalf("outcome %+v\n%s", got[0], out.String())
	}
}

func TestFlowAReportedFailureIsAnOutcomeNotACrash(t *testing.T) {
	f, r, out := flow("\n", true)
	r.err = errors.New("exit status 1")
	got, err := f.Run(context.Background(), []ToolPlan{pythonPlan("darwin", "brew")})
	if err != nil || !got[0].Ran || got[0].State != Missing || !strings.Contains(out.String(), "did not finish") {
		t.Fatalf("outcome %+v err %v\n%s", got, err, out.String())
	}
}

func TestFlowNeverRunsAdminCommandsOnYes(t *testing.T) {
	f, r, out := flow("", false)
	f.Yes = true
	plan := pythonPlan("linux", "apt-get")
	got, err := f.Run(context.Background(), []ToolPlan{plan})
	if err != nil || len(r.ran) != 0 || got[0].Ran {
		t.Fatalf("--yes must not run sudo: ran %v err %v", r.ran, err)
	}
	if !strings.Contains(out.String(), "sudo apt-get install python3") || !strings.Contains(out.String(), "--with-admin") {
		t.Fatalf("the command and the flag must show:\n%s", out.String())
	}
}

func TestFlowAsksBeforeAnAdminCommandAndDefaultsToRunItMyself(t *testing.T) {
	f, r, out := flow("\n\n", true) // the default install, then the default of the admin question
	_, _ = f.Run(context.Background(), []ToolPlan{pythonPlan("windows", "winget")})
	if len(r.ran) != 0 || !strings.Contains(out.String(), "administrator rights") || !strings.Contains(out.String(), "Run this, then open a new terminal") {
		t.Fatalf("ran %v\n%s", r.ran, out.String())
	}
	f, r, _ = flow("\n2\n", true) // let smith run it
	_, _ = f.Run(context.Background(), []ToolPlan{pythonPlan("windows", "winget")})
	if len(r.ran) != 1 || !strings.HasPrefix(r.ran[0], "winget install --id Python.Python.3.13") {
		t.Fatalf("ran %v", r.ran)
	}
}

func TestFlowWithAdminRunsWithoutAskingAgain(t *testing.T) {
	f, r, _ := flow("", false)
	f.Manager, f.WithAdmin = "apt", true
	_, _ = f.Run(context.Background(), []ToolPlan{pythonPlan("linux", "apt-get")})
	if len(r.ran) != 1 || !strings.HasPrefix(r.ran[0], "sudo apt-get install") {
		t.Fatalf("ran %v", r.ran)
	}
}

func TestFlowManagerFlagMustFitTheTool(t *testing.T) {
	f, r, _ := flow("", false)
	f.Manager = "brew"
	_, err := f.Run(context.Background(), []ToolPlan{pythonPlan("windows", "winget")})
	if err == nil || !strings.Contains(err.Error(), "winget") || len(r.ran) != 0 {
		t.Fatalf("want an error that names winget, got %v", err)
	}
}

func TestFlowPageMyselfAndSkipRunNothing(t *testing.T) {
	for choice, want := range map[string]string{"page": "python.org/downloads", "myself": "Ways to get Python", "skip": "fail for the python projects"} {
		f, r, out := flow(choice+"\n", true)
		got, _ := f.Run(context.Background(), []ToolPlan{pythonPlan("linux", "brew")})
		if len(r.ran) != 0 || got[0].Ran || !strings.Contains(out.String(), want) {
			t.Errorf("%s: ran %v\n%s", choice, r.ran, out.String())
		}
	}
}

func TestRenderListAndPlans(t *testing.T) {
	list := RenderList([]Status{{Need: Need{ToolNeed: missing("git").ToolNeed}, State: Found, Program: "git", Version: "2.45.1"}, missing("python")})
	for _, want := range []string{"✔ Git: git 2.45.1", "✖ Python: not installed", "If you skip it:", "smith tools install"} {
		if !strings.Contains(list, want) {
			t.Errorf("list lacks %q:\n%s", want, list)
		}
	}
	plans := RenderPlans([]ToolPlan{pythonPlan("windows", "winget")})
	for _, want := range []string{"1) Install with winget (default)", "winget install --id Python.Python.3.13 -e", "administrator", "2) Use the official page", "4) Skip it"} {
		if !strings.Contains(plans, want) {
			t.Errorf("plans lack %q:\n%s", want, plans)
		}
	}
}
