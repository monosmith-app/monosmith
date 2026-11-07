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

// handDrawnBox matches, inside a Go string, a tick box spelled by hand: "[ ]",
// "[x]" or "[-]".
var handDrawnBox = regexp.MustCompile(`"[^"]*\[[ x-]\]`)

// boxGuardExempt are the directories that may spell a box: choicerow draws it,
// theme owns the status icons (which are not tick boxes).
var boxGuardExempt = map[string]bool{"choicerow": true, "theme": true}

// TestNoScreenDrawsATickBoxByHand fails when a Go source under the TUI spells
// "[ ]", "[x]" or "[-]" itself instead of calling choicerow.
func TestNoScreenDrawsATickBoxByHand(t *testing.T) {
	err := filepath.WalkDir("..", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != ".." && boxGuardExempt[d.Name()] && filepath.Dir(path) == ".." {
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
			if m := handDrawnBox.FindString(line); m != "" {
				t.Errorf("%s:%d draws a tick box in %q by hand; use choicerow.Box", path, n+1, m)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// fixedState matches a button drawn in one state for good, as in
// button.Render(th, label, key, button.Rest): such a button never shows hover or
// active, so it does not behave like a button.
var fixedState = regexp.MustCompile(`button\.Render\([^\n]*button\.(Rest|Hover|Active)\)`)

// TestNoButtonIsDrawnInAFixedState fails when a Go source under the TUI calls
// Render with a literal state. A screen computes the state of each button (rest,
// hover under the pointer, active while pressed) and passes that.
func TestNoButtonIsDrawnInAFixedState(t *testing.T) {
	err := filepath.WalkDir("..", func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if filepath.Base(filepath.Dir(path)) == "button" || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for n, line := range strings.Split(string(raw), "\n") {
			if m := fixedState.FindString(line); m != "" && !strings.HasPrefix(strings.TrimSpace(line), "//") {
				t.Errorf("%s:%d draws %q in a fixed state; pass the state of the button", path, n+1, m)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// TestTheDetailPanelDrawsEveryActionWithTheButtonComponent fails when the panel
// goes back to drawing action rows by hand: the rows must come from Render, and a
// key must not be pushed to the right edge by padding.
func TestTheDetailPanelDrawsEveryActionWithTheButtonComponent(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "detail", "panel_view_mapper.go"))
	if err != nil {
		t.Fatal(err)
	}
	src := string(raw)
	if !strings.Contains(src, "button.Render(") {
		t.Error("the detail panel must draw its actions with button.Render")
	}
	for _, banned := range []string{"th.Selected(text)", "th.Bold(text)", "strings.Repeat(\" \", gap)"} {
		if strings.Contains(src, banned) {
			t.Errorf("the detail panel draws an action row by hand again (%s)", banned)
		}
	}
}
