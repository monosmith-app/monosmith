package adoption

import (
	"strings"
	"testing"
)

func TestExistingWorkspaceItemSaysWhenSmithTomlCannotBeRead(t *testing.T) {
	broken := existingWorkspaceItem([]byte("schema = 1\n[workspace\nname = \"x\"\n"))
	for _, want := range []string{"smith.toml cannot be read", "smith.toml:2:"} {
		if !strings.Contains(broken.Message, want) {
			t.Errorf("message %q lacks %q", broken.Message, want)
		}
	}
	if strings.Contains(broken.Message, "already") || !strings.Contains(broken.Step, "fix that line") || broken.Blocks != BlocksAdoption {
		t.Errorf("item = %+v", broken)
	}
}

func TestExistingWorkspaceItemKeepsTheOldWordsForAReadableFile(t *testing.T) {
	ok := existingWorkspaceItem([]byte("schema = 1\n\n[workspace]\nname = \"x\"\nscope = \"@x\"\n"))
	if ok.Message != "this is already a smith workspace" {
		t.Errorf("message = %q", ok.Message)
	}
}
