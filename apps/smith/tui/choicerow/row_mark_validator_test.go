package choicerow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// rowMarkLiterals are the old hand-spelled cursor marks of a row.
var rowMarkLiterals = []string{`"> "`, `"▸ "`}

// markGuardExempt are the directories that may spell a mark: choicerow draws
// the row mark and theme owns the icons.
var markGuardExempt = map[string]bool{"choicerow": true, "theme": true}

// TestNoScreenSpellsARowMarkByHand fails when a non-test Go source under the
// TUI uses the literal "> " or "▸ " instead of choicerow.Cursor.
func TestNoScreenSpellsARowMarkByHand(t *testing.T) {
	err := filepath.WalkDir("..", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != ".." && markGuardExempt[d.Name()] && filepath.Dir(path) == ".." {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for n, line := range strings.Split(string(raw), "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "//") {
				continue
			}
			for _, lit := range rowMarkLiterals {
				if strings.Contains(line, lit) {
					t.Errorf("%s:%d spells the row mark %s by hand; use choicerow.Cursor", path, n+1, lit)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCursorBlankHasTheWidthOfTheMark(t *testing.T) {
	th := theme.Default()
	if got, want := len([]rune(Cursor(th, false))), len([]rune(Cursor(th, true))); got != want {
		t.Fatalf("blank %d cells, mark %d", got, want)
	}
}
