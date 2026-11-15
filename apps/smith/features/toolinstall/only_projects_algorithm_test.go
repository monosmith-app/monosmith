package toolinstall

import (
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestAPythonChangeDoesNotWarnAboutTheDotnetSdk(t *testing.T) {
	need := func(name string, projects ...string) Need {
		return Need{ToolNeed: protocol.ToolNeed{Name: name}, Projects: projects}
	}
	all := []Need{need("git"), need("python", "api-py"), need("dotnet", "api-cs"), need("func", "api-cs")}
	got := OnlyProjects(all, []string{"api-py"})
	if len(got) != 2 || got[0].Name != "git" || got[1].Name != "python" {
		t.Fatalf("got %+v", got)
	}
	if got := OnlyProjects(all, nil); len(got) != len(all) {
		t.Errorf("nil means every project: %+v", got)
	}
}
