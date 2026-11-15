package doctorfix

import (
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestAStepShowsTheEnvironmentItAddsAndPassesItOn(t *testing.T) {
	step := protocol.FixStep{Argv: []string{"go", "install", "x@v1"}, Env: []string{"GOBIN=/w/.smith/bin"}}
	if got := StepCommand(step); got != "GOBIN=/w/.smith/bin go install x@v1" {
		t.Fatalf("got %q", got)
	}
	if env := stepEnv(step); env["GOBIN"] != "/w/.smith/bin" || len(env) != 1 {
		t.Fatalf("got %v", env)
	}
	if stepEnv(protocol.FixStep{Argv: []string{"npm"}}) != nil {
		t.Fatal("a step without Env adds none")
	}
}

func TestDisplayCommand(t *testing.T) {
	for _, tc := range []struct {
		argv []string
		want string
	}{
		{[]string{`C:\Users\me\AppData\smith.exe`, "adopt", "--init"}, "smith adopt --init"},
		{[]string{"/usr/local/bin/smith", "agent", "install"}, "smith agent install"},
		{[]string{"npm", "install"}, "npm install"},
		{nil, ""},
	} {
		if got := displayCommand(tc.argv); got != tc.want {
			t.Errorf("displayCommand(%v) = %q, want %q", tc.argv, got, tc.want)
		}
	}
}
