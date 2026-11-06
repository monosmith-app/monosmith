package pipeline

import (
	"strings"
)

// DropUse removes a module directory from go.work's use directives, in either
// form (use ./x, or a line of a use ( ... ) block), keeping everything else
// as it is. It reports whether anything was removed.
func DropUse(src, root string) (string, bool) {
	lines := strings.SplitAfter(src, "\n")
	inBlock, removed := false, false
	var out []string
	for _, line := range lines {
		code := strings.TrimSpace(line)
		if i := strings.Index(code, "//"); i >= 0 {
			code = strings.TrimSpace(code[:i])
		}
		switch {
		case strings.HasPrefix(code, "use (") || code == "use(":
			inBlock = true
		case inBlock && code == ")":
			inBlock = false
		case inBlock && sameDir(code, root):
			removed = true
			continue
		case strings.HasPrefix(code, "use ") && sameDir(strings.TrimSpace(strings.TrimPrefix(code, "use ")), root):
			removed = true
			continue
		}
		out = append(out, line)
	}
	return strings.Join(out, ""), removed
}

func sameDir(entry, root string) bool {
	entry = strings.Trim(entry, "\"`")
	return strings.TrimPrefix(entry, "./") == root
}
