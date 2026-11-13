package toolchainversions

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// literal matches a shared package written with a version, in Go
// (`"typescript": "~6.0.3"`) or in TOML (`typescript = "~6.0.3"`).
var literal = regexp.MustCompile(`(?:^|[\s{,])"?(typescript|@types/node|esbuild|vitest|jest|@jest/globals)"?\s*[:=]\s*"[~^]?\d`)

// TestNoStackWritesAToolchainVersion fails when a stack hard-codes the range
// of a package this slice owns.
func TestNoStackWritesAToolchainVersion(t *testing.T) {
	root := filepath.Join("..", "..", "plugin-node-toolchains")
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || strings.HasSuffix(path, "_test.go") {
			return err
		}
		if ext := filepath.Ext(path); ext != ".go" && ext != ".toml" {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for i, line := range strings.Split(string(raw), "\n") {
			if literal.MatchString(line) {
				t.Errorf("%s:%d writes a toolchain version; read toolchainversions instead: %s", path, i+1, strings.TrimSpace(line))
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// TestRangesCoverEveryConstant keeps Ranges in step with the constants.
func TestRangesCoverEveryConstant(t *testing.T) {
	r := Ranges()
	for pkg, want := range map[string]string{"typescript": TypeScript, "@types/node": TypesNode, "esbuild": Esbuild, "vitest": Vitest, "jest": Jest, "@jest/globals": Jest} {
		if r[pkg] != want {
			t.Errorf("%s: got %q, want %q", pkg, r[pkg], want)
		}
	}
}
