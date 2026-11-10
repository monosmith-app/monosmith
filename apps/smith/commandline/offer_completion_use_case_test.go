package commandline

import (
	"bytes"
	"strings"
	"testing"
)

func TestOnlyADoctorRunInATerminalOffersCompletion(t *testing.T) {
	for args, want := range map[string]bool{
		"doctor": true, "doctor --fix": true, "doctor --json": false, "build": false, "": false, "completion install": false,
		"doctor --fix --yes": false, "doctor --fix -y": false, "doctor --fix --only x": false, "doctor --fix --only x --choose 1": false,
	} {
		if got := offersCompletionAfter(strings.Fields(args)); got != want {
			t.Errorf("%q: got %v", args, got)
		}
	}
}

// The script of a shell calls `smith __complete` on every Tab, so the command
// tree must answer it, with the new subcommands in the answer.
func TestTheShellScriptCanCompleteTheCompletionCommandsThemselves(t *testing.T) {
	root := Root("0.0.0")
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"__complete", "completion", "ins"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "install") {
		t.Fatalf("Tab after `smith completion ins` does not offer install:\n%s", out.String())
	}
}

// Cobra's own `completion <shell>` commands still print scripts.
func TestSmithCompletionBashStillPrintsTheScript(t *testing.T) {
	root := Root("0.0.0")
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"completion", "bash"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "bash completion V2 for smith") {
		t.Fatalf("not the bash script: %.80s", out.String())
	}
}
