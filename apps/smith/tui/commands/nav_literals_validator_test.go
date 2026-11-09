package commands

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// navLiteral matches a switch case that spells a movement key by hand.
var navLiteral = regexp.MustCompile(`^\s*case\b.*"(up|down|k|j|pgup|pgdown|home|end)"`)

// navLiteralAllowed are the files that may still spell a movement key in a
// case, with the reason. It is empty: every view switches on Nav. A new entry
// needs a reason, and removing one is the goal.
var navLiteralAllowed = map[string]string{}

// TestNoViewSpellsAMovementKeyInACase fails when a production file under tui
// outside this package has a `case "up"`, `case "k"` or `case "pgdown"` style
// line, so the key names stay in Nav and a new view cannot bind its own.
func TestNoViewSpellsAMovementKeyInACase(t *testing.T) {
	var found []string
	err := filepath.WalkDir("..", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		slash := filepath.ToSlash(path)
		switch {
		case d.IsDir() && (d.Name() == "commands" || d.Name() == "testdata") && slash != "..":
			return filepath.SkipDir
		case d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go"):
			return nil
		}
		if _, ok := navLiteralAllowed[strings.TrimPrefix(slash, "../")]; ok {
			return nil
		}
		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for i, line := range strings.Split(string(src), "\n") {
			if navLiteral.MatchString(line) {
				found = append(found, strings.TrimPrefix(slash, "../")+":"+strconv.Itoa(i+1)+": "+strings.TrimSpace(line))
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(found) > 0 {
		t.Fatalf("switch on commands.Nav instead of a movement key name:\n%s", strings.Join(found, "\n"))
	}
}
