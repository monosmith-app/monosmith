package doctor

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/toolinstall"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolcatalog"
)

func TestAToolTwoProjectsAskForIsListedOnce(t *testing.T) {
	core := func(project string) Finding {
		return Finding{Source: "dotnet", Finding: protocol.Finding{Check: "Azure Functions Core Tools is installed", Detail: project + " needs func"}}
	}
	other := Finding{Source: "smith", Finding: protocol.Finding{Check: "go.work exists", OK: true}}
	got := withoutRepeatedMachineChecks([]Finding{core("api-cs"), other, core("api-cs2")})
	if len(got) != 2 || got[0].Detail != "api-cs needs func" || got[1].Check != "go.work exists" {
		t.Fatalf("got %+v", got)
	}
	ok := core("x")
	ok.OK = true
	if got := withoutRepeatedMachineChecks([]Finding{core("a"), ok}); len(got) != 2 {
		t.Errorf("a failed and a passed check of one tool are different findings: %+v", got)
	}
}

func TestCheckNameIgnoresCaseAndArticles(t *testing.T) {
	for _, pair := range [][2]string{{"a JDK is installed", "JDK is installed"}, {"The .NET SDK is installed", "the .net sdk is installed"}} {
		if checkName(pair[0]) != checkName(pair[1]) {
			t.Errorf("%q and %q must match", pair[0], pair[1])
		}
	}
}

func planFor(goos string, managers ...string) toolinstall.ToolPlan {
	dart, _ := toolcatalog.Find("dart")
	var ms []toolinstall.Manager
	for _, m := range managers {
		ms = append(ms, toolinstall.Manager{Name: m})
	}
	return toolinstall.BuildPlan(toolinstall.Status{Need: toolinstall.Need{ToolNeed: dart}, State: toolinstall.Missing}, ms, toolinstall.Env{OS: goos})
}

func TestToolFixPutsTheDefaultFirstAndNeverHidesAnAdminCommand(t *testing.T) {
	fix := ToolFix(planFor("windows", "winget"))
	if fix.Safe || len(fix.Steps) != 0 || len(fix.Files) != 0 || len(fix.Choices) != 1 {
		t.Fatalf("a tool fix is a question and runs nothing by itself: %+v", fix)
	}
	opts := fix.Choices[0].Options
	if !strings.HasPrefix(opts[0].Label, "I will run it myself") || len(opts[0].Fix.Steps) != 0 {
		t.Fatalf("an administrator command defaults to run it myself: %+v", opts[0])
	}
	if !strings.HasPrefix(opts[1].Label, "Let smith run it") || len(opts[1].Fix.Steps) != 1 {
		t.Fatalf("the second answer lets smith run it: %+v", opts[1])
	}
	last := opts[len(opts)-1]
	if !strings.HasPrefix(last.Label, "Skip it") || len(last.Fix.Steps) != 0 {
		t.Fatalf("the last answer skips: %+v", last)
	}
}

func TestToolFixWithoutAPackageManagerOffersThePageAndSkip(t *testing.T) {
	opts := ToolFix(planFor("linux")).Choices[0].Options
	if len(opts) != 3 || !strings.Contains(opts[0].Label, "https://dart.dev/get-dart") && !strings.Contains(opts[1].Label, "https://dart.dev/get-dart") {
		t.Fatalf("options %+v", opts)
	}
}

func TestToolFixForAToolOffPathOffersTheFolderChoice(t *testing.T) {
	dart, _ := toolcatalog.Find("dart")
	dir := `C:\Users\o\AppData\Local\Microsoft\WinGet\Packages\Google.DartSDK_x\dart-sdk\bin`
	st := toolinstall.Status{Need: toolinstall.Need{ToolNeed: dart}, State: toolinstall.Missing, OffPath: dir}
	fix := ToolFix(toolinstall.BuildPlan(st, []toolinstall.Manager{{Name: "winget"}}, toolinstall.Env{OS: "windows"}))
	opts := fix.Choices[0].Options
	if len(opts) != 4 {
		t.Fatalf("use it, add to PATH, show, skip: %+v", opts)
	}
	if steps := opts[0].Fix.Steps; !strings.HasPrefix(opts[0].Label, "Let smith use it") || len(steps) != 1 || strings.Join(steps[0].Argv, " ") != "smith tools use dart" {
		t.Fatalf("the default lets smith use the folder: %+v", opts[0])
	}
	if steps := opts[1].Fix.Steps; !strings.HasPrefix(opts[1].Label, "Add the folder to your user PATH") || len(steps) != 1 || steps[0].Argv[0] != "powershell" {
		t.Fatalf("the second answer changes the user PATH: %+v", opts[1])
	}
	if !strings.HasPrefix(opts[2].Label, "Show me the folder") || len(opts[2].Fix.Steps) != 0 || !strings.Contains(opts[2].Fix.Description, dir) {
		t.Fatalf("the third answer shows the folder and runs nothing: %+v", opts[2])
	}
}
