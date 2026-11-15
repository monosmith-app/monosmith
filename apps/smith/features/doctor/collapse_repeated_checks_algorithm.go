package doctor

import (
	"sort"
	"strconv"
	"strings"
)

// PassedLine is one line of the passed checks: a check, or the checks that
// differ only in which project they are about.
type PassedLine struct {
	Text string
	// Count is how many checks the line stands for.
	Count int
}

// CollapsePassed folds passed checks that read the same for every project
// ("npm workspaces include packages/a", "... packages/b") into one line,
// "npm workspaces include each project (23 projects)". A check that is not
// about a project, or is the only one of its kind, keeps its text. names are
// the projects' names and folders.
func CollapsePassed(findings []Finding, names []string) []PassedLine {
	sorted := append([]string(nil), names...)
	sort.Slice(sorted, func(i, j int) bool { return len(sorted[i]) > len(sorted[j]) })
	type group struct {
		line     int
		original string
		generic  string
	}
	var lines []PassedLine
	groups := map[string]*group{}
	for _, f := range findings {
		if !f.OK {
			continue
		}
		text := f.Check
		for _, name := range sorted {
			if name != "" {
				text = replaceToken(text, name, "each project")
			}
		}
		if text == f.Check {
			lines = append(lines, PassedLine{Text: f.Check, Count: 1})
			continue
		}
		key := f.Source + "\x00" + text
		if g, seen := groups[key]; seen {
			lines[g.line].Count++
			continue
		}
		groups[key] = &group{line: len(lines), original: f.Check, generic: text}
		lines = append(lines, PassedLine{Text: f.Check, Count: 1})
	}
	for _, g := range groups {
		if n := lines[g.line].Count; n > 1 {
			lines[g.line].Text = g.generic + " (" + strconv.Itoa(n) + " projects)"
		}
	}
	return lines
}

// replaceToken replaces name where it stands alone: not inside a longer word
// or path.
func replaceToken(s, name, with string) string {
	var b strings.Builder
	for {
		i := strings.Index(s, name)
		if i < 0 {
			b.WriteString(s)
			return b.String()
		}
		end := i + len(name)
		if (i > 0 && isNameChar(s[i-1])) || (end < len(s) && isNameChar(s[end])) {
			b.WriteString(s[:end])
			s = s[end:]
			continue
		}
		b.WriteString(s[:i])
		b.WriteString(with)
		s = s[end:]
	}
}

func isNameChar(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || strings.IndexByte("_-./@", c) >= 0
}
