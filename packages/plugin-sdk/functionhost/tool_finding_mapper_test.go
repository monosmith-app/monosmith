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
