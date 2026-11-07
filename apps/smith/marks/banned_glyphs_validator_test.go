package marks

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// banned lists the glyphs that overlap the next character in common terminal
// fonts, or that have an emoji form. The marks constants replace them.
var banned = []rune{
	0x2714, 0x2716, 0x2717, 0x26A0, 0x21B7, 0x26A1, // heavy tick, heavy cross, cross, warning, arrow, bolt
	0x2705, 0x274C, 0x23ED, 0x23F3, 0x1F534, 0x1F7E2, // emoji tick, cross, skip, hourglass, red, green
	0xFE0F, // emoji variation selector
}

var scannedExtensions = map[string]bool{
	".go": true, ".ts": true, ".tsx": true, ".js": true, ".mjs": true, ".cjs": true,
	".kt": true, ".py": true, ".cs": true, ".md": true, ".json": true, ".toml": true,
}

var skippedFolders = map[string]bool{
	"node_modules": true, "testdata": true, "dist": true, "build": true, ".gradle": true,
	"test": true, "tests": true, "fixtures": true, ".git": true, "tui-research": true,
}

func isTestFile(name string) bool {
	return strings.HasSuffix(name, "_test.go") || strings.Contains(name, ".spec.") || strings.Contains(name, ".test.") ||
		strings.HasSuffix(name, "Test.kt") || strings.HasSuffix(name, "Tests.cs") || strings.HasPrefix(name, "test_")
}

func isGenerated(text string) bool {
	lines := strings.SplitN(text, "\n", 6)
	for i, l := range lines {
		if i < 5 && (strings.Contains(l, "Code generated") || strings.Contains(l, "DO NOT EDIT")) {
			return true
		}
	}
	return false
}

// TestNoBannedGlyphs fails, naming file and line, when a production source or
// doc line holds a glyph that overlaps in some fonts or has an emoji form.
// Use the marks constants instead.
func TestNoBannedGlyphs(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	marksDir := filepath.Join(root, "apps", "smith", "marks")
	for _, rel := range []string{"apps/smith", "libs", "packages", "apps/vscode/src", "apps/jetbrains/src/main"} {
		start := filepath.Join(root, filepath.FromSlash(rel))
		if _, err := os.Stat(start); err != nil {
			continue
		}
		_ = filepath.WalkDir(start, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				if skippedFolders[d.Name()] || path == marksDir {
					return filepath.SkipDir
				}
				return nil
			}
			if !scannedExtensions[filepath.Ext(d.Name())] || isTestFile(d.Name()) {
				return nil
			}
			raw, readErr := os.ReadFile(path)
			if readErr != nil || isGenerated(string(raw)) {
				return nil
			}
			for i, line := range strings.Split(string(raw), "\n") {
				for _, g := range banned {
					if strings.ContainsRune(line, g) {
						shown, _ := filepath.Rel(root, path)
						t.Errorf("%s:%d holds U+%04X: use the marks constants (✓ × ▲ » ≈)", filepath.ToSlash(shown), i+1, g)
					}
				}
			}
			return nil
		})
	}
}
