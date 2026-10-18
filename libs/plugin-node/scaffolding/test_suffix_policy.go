package scaffolding

import (
	"errors"
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"
)

// DefaultTestSuffix is the suffix of a test file in a workspace that has
// none yet.
const DefaultTestSuffix = "spec"

var testFile = regexp.MustCompile(`\.(test|spec)\.(?:ts|tsx|mts|cts|js|jsx|mjs|cjs)$`)

// skippedTestDirs hold no test of the workspace's own.
var skippedTestDirs = map[string]bool{"node_modules": true, "dist": true, "coverage": true, "build": true}

// errEnoughTests ends the walk once the answer cannot change.
var errEnoughTests = errors.New("enough test files")

// TestSuffix is the suffix the workspace's Node tests already use, "test"
// (*.test.ts) or "spec" (*.spec.ts), so a new project never mixes the two.
// The more common one wins, a tie keeps DefaultTestSuffix, and a workspace
// with no tests yet gets DefaultTestSuffix.
func TestSuffix(workspaceRoot string) string {
	counts := map[string]int{}
	_ = filepath.WalkDir(workspaceRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if path != workspaceRoot && (skippedTestDirs[d.Name()] || strings.HasPrefix(d.Name(), ".")) {
				return filepath.SkipDir
			}
			return nil
		}
		if m := testFile.FindStringSubmatch(d.Name()); m != nil {
			counts[m[1]]++
			if counts[m[1]] >= 50 {
				return errEnoughTests
			}
		}
		return nil
	})
	if counts["test"] > counts["spec"] {
		return "test"
	}
	return DefaultTestSuffix
}
