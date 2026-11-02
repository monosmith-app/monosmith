package doctor

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/toolinstall"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolcatalog"
)

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
