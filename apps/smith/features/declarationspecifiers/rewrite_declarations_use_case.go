package declarationspecifiers

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Run rewrites the specifiers of every .d.ts file under dir and returns how
// many files changed.
func Run(dir string) (int, error) {
	exists := func(path string) bool {
		_, err := os.Stat(path)
		return err == nil
	}
	changed := 0
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".d.ts") {
			return err
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		out := Rewrite(string(raw), filepath.Dir(path), exists)
		if out == string(raw) {
			return nil
		}
		changed++
		return os.WriteFile(path, []byte(out), 0o644)
	})
	return changed, err
}
