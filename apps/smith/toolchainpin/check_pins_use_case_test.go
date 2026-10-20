package toolchainpin

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestAMismatchFailsWithTheFixAndAMatchPasses(t *testing.T) {
	projects := []protocol.Project{{Name: "web", Toolchain: "node 20.11.0"}, {Name: "api", Toolchain: "go 1.23.4"}}
	got := CheckPins(map[string]string{"node": "22", "go": "1.23", "python": "3.12"}, projects, "")
	if len(got) != 2 {
		t.Fatalf("python is not reported by any project, so no finding: %+v", got)
	}
	if !got[0].OK || got[0].Check != "go matches the pinned version 1.23" {
		t.Errorf("go matches: %+v", got[0])
	}
	node := got[1]
	if node.OK || !strings.Contains(node.Detail, "20.11.0") || !strings.Contains(node.Remedy, "mise use node@22") {
		t.Errorf("node mismatch names the version and the remedy: %+v", node)
	}
	if node.Fix == nil || !node.Fix.Safe || node.Fix.Files[0].Path != ".tool-versions" || !strings.Contains(node.Fix.Files[0].Content, "node 22") {
		t.Errorf("the fix writes .tool-versions: %+v", node.Fix)
	}
}

func TestTheFixIsNotSafeOverAnExistingFile(t *testing.T) {
	got := CheckPins(map[string]string{"node": "22"}, []protocol.Project{{Toolchain: "node 20.0.0"}}, "node 18.0.0\n")
	if got[0].Fix.Safe {
		t.Error("an existing .tool-versions is changed, not created: not safe")
	}
}
