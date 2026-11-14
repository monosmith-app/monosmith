package selfupdate

import (
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/updatenotice"
)

func TestDescribeNamesEachBlockingProcess(t *testing.T) {
	got := describe([]updatenotice.Blocker{{PID: 12, Command: "smith.exe mcp"}, {PID: 40, Command: "smith.exe serve"}})
	if got != "smith.exe mcp (PID 12), smith.exe serve (PID 40)" {
		t.Errorf("describe = %q", got)
	}
}

func TestCommandHasTheTwoFlagsAndTakesNoArgument(t *testing.T) {
	cmd := Command("1.0.0")
	for _, name := range []string{"check", "yes"} {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("no --%s flag", name)
		}
	}
	if err := cmd.Args(cmd, []string{"x"}); err == nil {
		t.Error("an argument was accepted")
	}
}
