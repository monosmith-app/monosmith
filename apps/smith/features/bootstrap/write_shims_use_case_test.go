package bootstrap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteShimsWritesOnceAndCheckWritesNothing(t *testing.T) {
	root := t.TempDir()
	changed, err := WriteShims(root, true)
	if err != nil || len(changed) != 2 {
		t.Fatalf("check: %v %v", changed, err)
	}
	if _, err := os.Stat(filepath.Join(root, "tools")); err == nil {
		t.Fatal("check wrote files")
	}
	if changed, err = WriteShims(root, false); err != nil || len(changed) != 2 {
		t.Fatalf("write: %v %v", changed, err)
	}
	if changed, _ = WriteShims(root, false); len(changed) != 0 {
		t.Errorf("second write changed %v", changed)
	}
	if err := os.WriteFile(filepath.Join(root, "tools", "smith", "smith"), []byte("edited"), 0o755); err != nil {
		t.Fatal(err)
	}
	if changed, _ = WriteShims(root, true); len(changed) != 1 || changed[0] != ShellPath {
		t.Errorf("an edited launcher was not seen: %v", changed)
	}
}

func TestShimsUseTheLineEndsOfTheirShell(t *testing.T) {
	sh, cmd := ShellShim(), CmdShim()
	if !strings.HasPrefix(sh, "#!/bin/sh\n") || strings.Contains(sh, "\r") {
		t.Error("the sh launcher needs a shebang and LF line ends")
	}
	if strings.Count(cmd, "\r\n") != strings.Count(cmd, "\n") {
		t.Error("the cmd launcher needs CRLF line ends")
	}
}
