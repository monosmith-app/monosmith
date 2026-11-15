package adoptsplit

import (
	"regexp"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// devGroup is one list of development dependencies in pyproject.toml.
type devGroup struct {
	table, key string
}

// devGroups are the lists of pyproject.toml that hold development
// dependencies: the optional-dependency extras and the PEP 735 groups that
// are conventionally for development.
var devGroups = func() []devGroup {
	var out []devGroup
	for _, key := range []string{"dev", "test", "tests", "lint"} {
		out = append(out, devGroup{"project.optional-dependencies", key}, devGroup{"dependency-groups", key})
	}
	return out
}()

// readDevGroup returns the entries of a development list, as written in the
// requirement strings, and false when the list is missing or holds anything
// but strings (a PEP 735 include table, for one).
func readDevGroup(text string, g devGroup) ([]string, bool) {
	start, end, ok := keyArray(text, g)
	if !ok {
		return nil, false
	}
	var out struct{ V []string }
	if toml.Unmarshal([]byte("v = "+text[start:end]), &out) != nil {
		return nil, false
	}
	return out.V, true
}

// writeDevGroup replaces the list with the entries that remain, or removes
// the key when none do.
func writeDevGroup(text string, g devGroup, remain []string) string {
	start, end, ok := keyArray(text, g)
	if !ok {
		return text
	}
	line := lineStart(text, start)
	tail := end
	if i := strings.IndexByte(text[tail:], '\n'); i >= 0 {
		tail += i + 1
	} else {
		tail = len(text)
	}
	if len(remain) == 0 {
		return text[:line] + text[tail:]
	}
	quoted := make([]string, len(remain))
	for i, r := range remain {
		quoted[i] = `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(r) + `"`
	}
	eol := "\n"
	if strings.Contains(text, "\r\n") {
		eol = "\r\n"
	}
	return text[:start] + "[" + strings.Join(quoted, ", ") + "]" + eol + text[tail:]
}

func lineStart(text string, i int) int {
	return strings.LastIndexByte(text[:i], '\n') + 1
}

var tableHeader = regexp.MustCompile(`(?m)^[ \t]*\[\[?([^\]\r\n]+)\]\]?[ \t]*(?:#[^\r\n]*)?$`)

// keyArray finds the array value of key in the table: the span of its
// brackets, quotes respected. It reports false when the table or key is
// missing or the value is not an array.
func keyArray(text string, g devGroup) (start, end int, ok bool) {
	bodyStart, bodyEnd := -1, len(text)
	for _, m := range tableHeader.FindAllStringSubmatchIndex(text, -1) {
		if bodyStart >= 0 {
			bodyEnd = m[0]
			break
		}
		if strings.TrimSpace(text[m[2]:m[3]]) == g.table {
			bodyStart = m[1]
		}
	}
	if bodyStart < 0 {
		return 0, 0, false
	}
	loc := regexp.MustCompile(`(?m)^[ \t]*` + regexp.QuoteMeta(g.key) + `[ \t]*=[ \t]*\[`).FindStringIndex(text[bodyStart:bodyEnd])
	if loc == nil {
		return 0, 0, false
	}
	start = bodyStart + loc[1] - 1
	depth, quote := 0, byte(0)
	for i := start; i < bodyEnd; i++ {
		c := text[i]
		switch {
		case quote != 0:
			if c == '\\' && quote == '"' {
				i++
			} else if c == quote {
				quote = 0
			}
		case c == '"' || c == '\'':
			quote = c
		case c == '[':
			depth++
		case c == ']':
			depth--
			if depth == 0 {
				return start, i + 1, true
			}
		}
	}
	return 0, 0, false
}

// takeRuffTables removes the [tool.ruff] tables from the text and returns
// them as the text of a ruff.toml: [tool.ruff] becomes the top of the file and
// [tool.ruff.lint] becomes [lint]. It returns "" when there are none.
func takeRuffTables(text string) (rest, ruff string) {
	lines := strings.SplitAfter(text, "\n")
	var keptLines []string
	var top, tables []string
	var current *[]string
	for _, line := range lines {
		if m := tableHeader.FindStringSubmatch(strings.TrimRight(line, "\r\n")); m != nil {
			name := strings.TrimSpace(m[1])
			switch {
			case name == "tool.ruff":
				current = &top
				continue
			case strings.HasPrefix(name, "tool.ruff."):
				current = &tables
				*current = append(*current, strings.Replace(line, "tool.ruff.", "", 1))
				continue
			default:
				current = nil
			}
		}
		if current != nil {
			*current = append(*current, line)
		} else {
			keptLines = append(keptLines, line)
		}
	}
	if len(top) == 0 && len(tables) == 0 {
		return text, ""
	}
	return strings.TrimRight(strings.Join(keptLines, ""), "\r\n") + "\n", strings.TrimSpace(strings.TrimSpace(strings.Join(top, ""))+"\n\n"+strings.Join(tables, "")) + "\n"
}

// dropEmptyTables removes the header of each table that has no key left, so
// moving the last development dependency out does not leave an empty table.
func dropEmptyTables(text string, tables ...string) string {
	for _, table := range tables {
		for _, m := range tableHeader.FindAllStringSubmatchIndex(text, -1) {
			if strings.TrimSpace(text[m[2]:m[3]]) != table {
				continue
			}
			end := len(text)
			for _, next := range tableHeader.FindAllStringIndex(text[m[1]:], 1) {
				end = m[1] + next[0]
			}
			if strings.TrimSpace(text[m[1]:end]) == "" {
				text = text[:m[0]] + strings.TrimLeft(text[end:], "\r\n")
			}
			break
		}
	}
	return text
}
