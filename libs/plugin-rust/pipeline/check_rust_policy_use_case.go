package pipeline

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

// CheckRustPolicy is the Rust unsafe-count policy: the .rs files under roots
// (skipping target and hidden folders) hold at most maxUnsafe unsafe blocks,
// functions, impls and traits in all. It returns one line per file with
// unsafe code, then a summary when the budget is exceeded.
func CheckRustPolicy(roots []string, maxUnsafe int) ([]string, error) {
	counts := map[string]int{}
	total := 0
	for _, root := range roots {
		err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if p != root && (d.Name() == "target" || d.Name()[0] == '.') {
					return filepath.SkipDir
				}
				return nil
			}
			if filepath.Ext(p) != ".rs" {
				return nil
			}
			src, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			if n := CountUnsafe(string(src)); n > 0 {
				counts[filepath.ToSlash(p)] = n
				total += n
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	if total <= maxUnsafe {
		return nil, nil
	}
	var lines []string
	for p, n := range counts {
		lines = append(lines, fmt.Sprintf("%s: %d unsafe", p, n))
	}
	sort.Strings(lines)
	return append(lines, fmt.Sprintf("%d unsafe in all, more than the %d allowed (audit_max_unsafe)", total, maxUnsafe)), nil
}
