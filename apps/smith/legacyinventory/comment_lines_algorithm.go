package legacyinventory

import "strings"

// WithoutCommentLines drops the comments a file is certain to have, so a
// package named only in prose ("do not edit @mnci/eslint-config") does not
// count as used. It is deliberately cautious: only a line that starts with a
// comment marker, and a block comment that starts a line, are dropped. A
// trailing comment stays, because telling it from a string or a regular
// expression takes a parser, and keeping text only ever keeps a package.
func WithoutCommentLines(text, ext string) string {
	hash := ext == ".yml" || ext == ".yaml" || ext == ".sh" || ext == ".toml" || ext == ".ps1" || ext == ".cmd" || ext == ""
	slash := !hash && ext != ".html" && ext != ".vue" && ext != ".svelte" && ext != ".astro"
	var out strings.Builder
	inBlock := false
	for _, line := range strings.SplitAfter(text, "\n") {
		trimmed := strings.TrimSpace(line)
		switch {
		case inBlock:
			if end := strings.Index(line, "*/"); end >= 0 {
				inBlock = false
				out.WriteString(line[end+2:])
			}
		case hash && strings.HasPrefix(trimmed, "#"):
		case slash && strings.HasPrefix(trimmed, "//"):
		case slash && strings.HasPrefix(trimmed, "/*"):
			start := strings.Index(line, "/*")
			if end := strings.Index(line[start+2:], "*/"); end >= 0 {
				out.WriteString(line[start+2+end+2:])
			} else {
				inBlock = true
			}
		default:
			out.WriteString(line)
		}
	}
	return out.String()
}
