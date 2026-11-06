package button

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// handDrawn matches, inside a string, a clickable item drawn by hand: a label
// between square brackets, spaced as "[ Run  Enter ]" or ending in a key as
// "[Run (r)]". Index expressions and the checkbox marks "[ ]" and "[x]" do not
// match.
var handDrawn = regexp.MustCompile(`"[^"]*(\[ [A-Za-z+{][^\]\[]* \]|\[[A-Za-z+{][^\]\[]* \([^)]+\)\])[^"]*"`)

// TestNoScreenDrawsAClickableItemByHand fails when a Go source or a catalog
// entry under the TUI spells a bracketed label itself instead of calling Render.
func TestNoScreenDrawsAClickableItemByHand(t *testing.T) {
	root := filepath.Join("..")
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if filepath.Base(filepath.Dir(path)) == "button" || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		if !strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, ".json") {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for n, line := range strings.Split(string(raw), "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "//") || strings.Contains(line, "words:literal") {
				continue
			}
			if m := handDrawn.FindString(line); m != "" {
				t.Errorf("%s:%d draws %q by hand; use button.Render (add // words:literal if it is not a clickable item)", path, n+1, m)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
