package adoptsplit

import "testing"

func TestSmithsOwnPackagesAreToolsOfTheWorkspace(t *testing.T) {
	for _, name := range []string{"@monosmith/eslint-config", "@monosmith/plugin-sdk"} {
		if !IsNodeTool(name) {
			t.Errorf("%s is not a tool: adopt would ask where it goes", name)
		}
	}
	if IsNodeTool("left-pad") {
		t.Error("an unknown package must still be asked about")
	}
}
