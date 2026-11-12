package toolinstall

import (
	"bufio"
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// dartOffPath is the Dart SDK unpacked by winget under a fake LOCALAPPDATA, and
// its status: missing from PATH, with the folder found.
func dartOffPath(t *testing.T) (local, dir string, st Status) {
	t.Helper()
	local = t.TempDir()
	dir = putFile(t, WingetPackagesDir(local), "Google.DartSDK_Microsoft.Winget.Source_8wekyb3d8bbwe", "dart-sdk", "bin", "dart.exe")
	need := wingetNeed("dart", "Google.DartSDK")
	need.Title = "the Dart SDK"
	return local, dir, Status{Need: need, State: Missing, OffPath: dir}
}

func values(plan ToolPlan) string {
	var v []string
	for _, o := range plan.Options {
		v = append(v, o.Value)
	}
	return strings.Join(v, " ")
}

func TestBuildPlanOffersTheFolderChoiceForAToolOffPath(t *testing.T) {
	_, dir, st := dartOffPath(t)
	managers := DetectManagers("windows", onPath("winget"))
	plan := BuildPlan(st, managers, Env{OS: "windows", ToolFolders: "tool-folders.json"})
	if got := values(plan); got != "install:smith-path install:user-path myself skip" {
		t.Fatalf("options %q: a new winget install would not help, so it is not offered", got)
	}
	if plan.Default != "install:smith-path" {
		t.Fatalf("default %q", plan.Default)
	}
	use, user := plan.Options[0], plan.Options[1]
	if strings.Join(use.Command, " ") != "smith tools use dart" || !use.ChangesPath() || !strings.Contains(use.Hint, dir) || !strings.Contains(use.Hint, "tool-folders.json") {
		t.Fatalf("use option %+v", use)
	}
	if user.Command[0] != "powershell" || !strings.Contains(user.Command[3], "'"+dir+"'") || user.Admin || !user.ChangesPath() {
		t.Fatalf("user PATH option %+v", user)
	}
	if !strings.Contains(st.Problem(), dir) || !strings.Contains(st.Problem(), "not on PATH") {
		t.Fatalf("problem %q", st.Problem())
	}
	if other := BuildPlan(st, nil, Env{OS: "linux"}); values(other) != "install:smith-path myself skip" {
		t.Fatalf("the user PATH command is for Windows only: %q", values(other))
	}
}

func TestUserPathScriptQuotesTheFolder(t *testing.T) {
	got := UserPathScript(`C:\Users\o'neil\bin`)
	if !strings.HasPrefix(got, `$d='C:\Users\o''neil\bin'; `) || !strings.Contains(got, "DoNotExpandEnvironmentNames") || !strings.Contains(got, "ExpandString") {
		t.Fatalf("script %q", got)
	}
}

func folderFlow(t *testing.T, input string, local string) (Flow, *fakeRunner, *bytes.Buffer, string) {
	t.Helper()
	t.Setenv("PATH", t.TempDir())
	r := &fakeRunner{}
	var out bytes.Buffer
	file := filepath.Join(t.TempDir(), "tool-folders.json")
	return Flow{
		In: bufio.NewReader(strings.NewReader(input)), Out: &out, Interactive: true, Runner: r,
		ToolFolders: file, LocalAppData: local,
	}, r, &out, file
}

func TestFlowUsesTheFolderWithoutAChangeToPath(t *testing.T) {
	local, dir, st := dartOffPath(t)
	f, r, out, file := folderFlow(t, "\n", local)
	got, err := f.Run(context.Background(), []ToolPlan{BuildPlan(st, nil, Env{OS: "windows"})})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.ran) != 0 {
		t.Fatalf("nothing runs for this answer: %v", r.ran)
	}
	recorded, _ := toolrun.ReadToolFolders(file)
	if recorded["dart"] != dir || got[0].State != Found {
		t.Fatalf("recorded %v, outcome %+v", recorded, got[0])
	}
	if !strings.Contains(out.String(), "smith now uses the Dart SDK from "+dir) || !strings.Contains(out.String(), "Your PATH did not change") {
		t.Fatalf("output:\n%s", out.String())
	}
}

func TestFlowAddsTheFolderToTheUserPathAfterAConfirmation(t *testing.T) {
	local, dir, st := dartOffPath(t)
	f, r, out, _ := folderFlow(t, "2\n1\n", local) // the user PATH answer, then "Add it to my PATH"
	got, err := f.Run(context.Background(), []ToolPlan{BuildPlan(st, nil, Env{OS: "windows"})})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.ran) != 1 || !strings.HasPrefix(r.ran[0], "powershell -NoProfile -Command $d='"+dir+"'") {
		t.Fatalf("ran %v", r.ran)
	}
	if got[0].State != Found || !strings.Contains(out.String(), "Open a new terminal") {
		t.Fatalf("outcome %+v\n%s", got[0], out.String())
	}
}

func TestFlowChangesNoPathWhenThePersonSaysNo(t *testing.T) {
	local, dir, st := dartOffPath(t)
	f, r, out, _ := folderFlow(t, "2\n2\n", local)
	if _, err := f.Run(context.Background(), []ToolPlan{BuildPlan(st, nil, Env{OS: "windows"})}); err != nil {
		t.Fatal(err)
	}
	if len(r.ran) != 0 || !strings.Contains(out.String(), dir) || !strings.Contains(out.String(), "smith tools use dart") {
		t.Fatalf("ran %v\n%s", r.ran, out.String())
	}
}

func TestFlowShowsTheFolderAndChangesNothing(t *testing.T) {
	local, dir, st := dartOffPath(t)
	f, r, out, file := folderFlow(t, "3\n", local)
	if _, err := f.Run(context.Background(), []ToolPlan{BuildPlan(st, nil, Env{OS: "windows"})}); err != nil {
		t.Fatal(err)
	}
	recorded, _ := toolrun.ReadToolFolders(file)
	if len(r.ran) != 0 || len(recorded) != 0 || !strings.Contains(out.String(), "the Dart SDK is in "+dir+", which is not on PATH") {
		t.Fatalf("ran %v, recorded %v\n%s", r.ran, recorded, out.String())
	}
}

// exitCode is an error with an exit code, as *exec.ExitError is.
type exitCode int

func (e exitCode) Error() string { return "exit status" }
func (e exitCode) ExitCode() int { return int(e) }

// offPathAfter answers, once the runner ran, that the program is in a folder off PATH.
type offPathAfter struct {
	runner *fakeRunner
	dir    string
}

func (a offPathAfter) Detect(context.Context, Need) Probe {
	if len(a.runner.ran) > 0 {
		return Probe{Folder: a.dir, Version: "3.13.7"}
	}
	return Probe{}
}

func TestFlowTakesWingetsAlreadyInstalledAndOffersTheFolder(t *testing.T) {
	local, dir, st := dartOffPath(t)
	st.OffPath = "" // the first detection did not look: the winget install comes first
	f, r, out, file := folderFlow(t, "\n\n", local)
	r.err = exitCode(0x8A15002B)
	f.Detector = offPathAfter{runner: r, dir: dir}
	plan := BuildPlan(st, DetectManagers("windows", onPath("winget")), Env{OS: "windows"})
	plan.Options[0].Admin = false // run winget without the administrator question
	got, err := f.Run(context.Background(), []ToolPlan{plan})
	if err != nil {
		t.Fatal(err)
	}
	text := out.String()
	if strings.Contains(text, "did not finish") || !strings.Contains(text, "already installed") || !strings.Contains(text, "winget put the Dart SDK in "+dir) {
		t.Fatalf("output:\n%s", text)
	}
	recorded, _ := toolrun.ReadToolFolders(file)
	if recorded["dart"] != dir || got[0].State != Found || !got[0].Ran {
		t.Fatalf("recorded %v, outcome %+v", recorded, got[0])
	}
}

func TestFlowStillReportsARealWingetFailure(t *testing.T) {
	local, _, st := dartOffPath(t)
	st.OffPath = ""
	f, r, out, _ := folderFlow(t, "\n", local)
	r.err = exitCode(0x8A150107) // no network
	plan := BuildPlan(st, DetectManagers("windows", onPath("winget")), Env{OS: "windows"})
	plan.Options[0].Admin = false
	if _, err := f.Run(context.Background(), []ToolPlan{plan}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "The install did not finish") {
		t.Fatalf("output:\n%s", out.String())
	}
}
