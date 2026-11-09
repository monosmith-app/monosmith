package operation

import "testing"

func TestTheCommitAdviceFitsTheInterface(t *testing.T) {
	t.Setenv(InterfaceEnv, "")
	if got := CommitNextTime("add --commit"); got != "add --commit" {
		t.Errorf("at a command line: %q", got)
	}
	t.Setenv(InterfaceEnv, InterfaceTUI)
	if got := CommitNextTime("add --commit"); got != `tick "Commit this change"` {
		t.Errorf("in the terminal interface: %q", got)
	}
}
