package flagclients

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/actions/flagvocabulary"
)

// repoRoot is the repository root, two folders above libs/actions.
const repoRoot = "../../.."

// The generated files must be what the vocabulary says. Fix: run `go run ./flaggen`
// in libs/actions.
func TestGeneratedFilesAreUpToDate(t *testing.T) {
	stale, err := Sync(repoRoot, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(stale) > 0 {
		t.Fatalf("not up to date: %s\nrun `go run ./flaggen` in libs/actions", strings.Join(stale, ", "))
	}
}

// The VS Code extension and the JetBrains plugin take the shared flag names from
// the generated constants: no old spelling, and no hand-typed spelling of a
// concept that the vocabulary lists commands for. Tests may pin the literal text.
func TestClientsDoNotHandTypeSharedFlags(t *testing.T) {
	var banned []string
	for _, c := range flagvocabulary.All() {
		for _, s := range c.Synonyms {
			banned = append(banned, "--"+s.Name)
			if s.Short != "" {
				banned = append(banned, "-"+s.Short)
			}
		}
		if len(c.Commands) > 0 {
			banned = append(banned, "--"+c.Flag)
			if c.Short != "" {
				banned = append(banned, "-"+c.Short)
			}
		}
	}
	var quoted []string
	for _, b := range banned {
		quoted = append(quoted, regexp.QuoteMeta(b))
	}
	pattern := regexp.MustCompile(`['"](` + strings.Join(quoted, "|") + `)['"]`)
	generated := map[string]bool{}
	for _, o := range Outputs {
		generated[filepath.ToSlash(o)] = true
	}
	roots := []string{"apps/vscode/src", "apps/jetbrains/src/main"}
	for _, r := range roots {
		_ = filepath.WalkDir(filepath.Join(repoRoot, r), func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			rel, _ := filepath.Rel(repoRoot, path)
			rel = filepath.ToSlash(rel)
			ext := filepath.Ext(path)
			if (ext != ".ts" && ext != ".kt") || generated[rel] || strings.HasSuffix(path, ".spec.ts") {
				return nil
			}
			data, _ := os.ReadFile(path)
			for i, line := range strings.Split(string(data), "\n") {
				if m := pattern.FindString(line); m != "" {
					t.Errorf("%s:%d types %s by hand: use the generated constant (flag/longFlag in smith-flags, SmithFlagsConfig in Kotlin)", rel, i+1, m)
				}
			}
			return nil
		})
	}
}
