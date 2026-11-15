package cipipeline

import "testing"

func TestVerifyIsTheCIVerifyPhaseWithAFlagToSkipTheCache(t *testing.T) {
	cmd := VerifyCommand("0.0.0")
	if cmd.Use != "verify" {
		t.Fatalf("command is %q", cmd.Use)
	}
	if f := cmd.Flags().Lookup("no-cache"); f == nil || f.DefValue != "false" {
		t.Fatalf("--no-cache must exist and be off by default: %+v", f)
	}
	if err := cmd.Args(cmd, []string{"extra"}); err == nil {
		t.Fatal("verify takes no arguments")
	}
}
