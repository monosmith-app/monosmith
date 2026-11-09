package cipipeline

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAFileThatIsGoneIsReportedRemovedNotWritten(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "kept.yml"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := changeVerb(root, "kept.yml"); got != "wrote" {
		t.Errorf("an existing file: %q", got)
	}
	if got := changeVerb(root, ".devcontainer/devcontainer.json"); got != "removed" {
		t.Errorf("a missing file: %q", got)
	}
}

func TestCiAloneRunsTheSetupPhaseAndRefusesStrayWords(t *testing.T) {
	cmd := Command("test")
	if cmd.RunE == nil {
		t.Fatal("smith ci alone does nothing: it must install what a clean checkout needs")
	}
	cmd.SetArgs([]string{"nonsense"})
	cmd.SilenceUsage, cmd.SilenceErrors = true, true
	if err := cmd.Execute(); err == nil {
		t.Fatal("smith ci nonsense must fail, not install")
	}
}
