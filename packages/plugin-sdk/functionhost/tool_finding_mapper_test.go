package functionhost

import (
	"strings"
	"testing"
)

func TestTheFindingIsAWarningWithTheWayToGetTheTool(t *testing.T) {
	for host, want := range map[string]string{Azure: "func", AWS: "sam", GCP: ""} {
		tool, ok := ToolOf(host)
		if !ok || tool.Binary != want {
			t.Fatalf("%s: %+v %v", host, tool, ok)
		}
		missing := FindingFor(tool, false)
		if missing.OK || !missing.Warning || missing.Remedy == "" {
			t.Errorf("%s: a missing tool is a warning that says how to get it: %+v", host, missing)
		}
		if !FindingFor(tool, true).OK {
			t.Errorf("%s: an installed tool passes", host)
		}
	}
	azure, _ := ToolOf(Azure)
	if !strings.Contains(azure.Install, "smith pipeline deploy-azure-function") {
		t.Error("the Azure text names the deploy pipeline smith has")
	}
	if _, ok := ToolOf("react"); ok {
		t.Error("only the three hosts have a tool")
	}
}

// TestFindingForProblem lists a missing tool by what is wrong, and a found
// tool by nothing.
func TestFindingForProblem(t *testing.T) {
	tool := Tool{Name: "SAM CLI"}
	if got := FindingFor(tool, true).Problem; got != "" {
		t.Errorf("found tool has Problem %q", got)
	}
	if got := FindingFor(tool, false).Problem; got != "SAM CLI is not found, so the function cannot run locally" {
		t.Errorf("Problem = %q", got)
	}
}

func TestFindingForIsAToolchainFinding(t *testing.T) {
	if f := FindingFor(Tool{Name: "func"}, false); f.Area != "toolchain" || f.Project != "" {
		t.Fatalf("%+v", f)
	}
}
