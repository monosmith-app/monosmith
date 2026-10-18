package scaffolding

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTestSuffixFollowsTheWorkspace(t *testing.T) {
	cases := []struct {
		name  string
		files []string
		want  string
	}{
		{"no tests yet", []string{"packages/a/src/index.ts"}, "spec"},
		{"test files", []string{"packages/a/src/x.use-case.test.ts", "apps/b/src/y.test.tsx"}, "test"},
		{"spec files", []string{"packages/a/src/x.use-case.spec.ts"}, "spec"},
		{"more test than spec", []string{"a/x.test.ts", "a/y.test.ts", "a/z.spec.ts"}, "test"},
		{"a tie keeps the default", []string{"a/x.test.ts", "a/z.spec.ts"}, "spec"},
		{"installed and built files do not count", []string{"node_modules/p/x.test.js", "dist/a/y.test.js", "packages/a/src/z.spec.ts"}, "spec"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			root := t.TempDir()
			for _, f := range c.files {
				p := filepath.Join(root, filepath.FromSlash(f))
				_ = os.MkdirAll(filepath.Dir(p), 0o755)
				_ = os.WriteFile(p, nil, 0o644)
			}
			if got := TestSuffix(root); got != c.want {
				t.Errorf("TestSuffix = %q, want %q", got, c.want)
			}
		})
	}
}
