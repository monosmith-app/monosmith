package commands

import "testing"

func TestNothingStartsAWorkspaceOverAnUnreadableSmithToml(t *testing.T) {
	broken := State{View: TabGuide, BrokenConfig: true}
	for _, key := range []string{"n", "a"} {
		if c, _, ok := ByKey(key, broken); ok && (c.ID == "file.new" || c.ID == "file.adopt") {
			t.Errorf("%q must not offer %s over a smith.toml that cannot be read", key, c.ID)
		}
	}
}
