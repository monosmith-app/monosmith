package declaredimports

import (
	"regexp"
	"strings"
)

var (
	importLine = regexp.MustCompile(`^([ \t]*)(?:from[ \t]+([A-Za-z_][\w.]*)[ \t]+import\b|import[ \t]+([A-Za-z_][\w.]*(?:[ \t]+as[ \t]+\w+)?(?:[ \t]*,[ \t]*[A-Za-z_][\w.]*(?:[ \t]+as[ \t]+\w+)?)*))`)
	// optionalBlock opens a block whose imports are not required: a guarded
	// import, its fallback, or one only a type checker reads.
	optionalBlock = regexp.MustCompile(`^(?:try\s*:|except\b.*:|if\s+(?:typing\.)?TYPE_CHECKING\s*:)`)
)

// FileImports returns the top-level modules a Python source imports. Relative
// imports, imports in docstrings, and imports inside try/except blocks or
// `if TYPE_CHECKING:` (optional by intent) are left out.
func FileImports(src string) []string {
	lines := strings.Split(src, "\n")
	var out []string
	inString := ""
	for i, line := range lines {
		if inString != "" {
			if strings.Contains(line, inString) {
				inString = ""
			}
			continue
		}
		if q := openedString(line); q != "" {
			inString = q
			continue
		}
		m := importLine.FindStringSubmatch(line)
		if m == nil || optional(lines, i, len(m[1])) {
			continue
		}
		modules := m[2]
		if modules == "" {
			modules = m[3]
		}
		for _, mod := range strings.Split(modules, ",") {
			mod = strings.TrimSpace(mod)
			if before, _, ok := strings.Cut(mod, " "); ok {
				mod = before
			}
			if top, _, _ := strings.Cut(mod, "."); top != "" {
				out = append(out, top)
			}
		}
	}
	return out
}

// openedString returns the triple quote a line opens and does not close.
func openedString(line string) string {
	for _, q := range []string{`"""`, `'''`} {
		if strings.Count(line, q)%2 == 1 {
			return q
		}
	}
	return ""
}

// optional reports whether the nearest enclosing block of the line at index i
// (indented by indent) is an optionalBlock.
func optional(lines []string, i, indent int) bool {
	for j := i - 1; j >= 0 && indent > 0; j-- {
		line := strings.TrimRight(lines[j], " \t\r")
		trimmed := strings.TrimLeft(line, " \t")
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if depth := len(line) - len(trimmed); depth < indent {
			if optionalBlock.MatchString(trimmed) {
				return true
			}
			indent = depth
		}
	}
	return false
}
