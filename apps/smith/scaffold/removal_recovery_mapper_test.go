package scaffold

import (
	"strings"
	"testing"
)

func TestARemovalSaysHowToGetTheProjectBack(t *testing.T) {
	plain := strings.Join(RecoveryLines(Removal{Project: "web", Root: "apps/web"}), "\n")
	for _, want := range []string{"git restore --source=HEAD --staged --worktree -- apps/web", "smith undo", "git diff"} {
		if !strings.Contains(plain, want) {
			t.Errorf("lacks %q:\n%s", want, plain)
		}
	}
	if strings.Contains(plain, ".smith/removed") {
		t.Errorf("no copy was kept, so none is named:\n%s", plain)
	}
	kept := strings.Join(RecoveryLines(Removal{Project: "web", Root: "apps/web", KeptAt: ".smith/removed/web-1"}), "\n")
	if !strings.Contains(kept, "move .smith/removed/web-1 back to apps/web") {
		t.Errorf("the kept copy is not named:\n%s", kept)
	}
}
