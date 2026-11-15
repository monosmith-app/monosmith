package languageserver

import (
	"strings"
	"unicode"
)

// cursorKind says what the text left of the cursor is.
type cursorKind int

const (
	cursorNone cursorKind = iota
	cursorHeader
	cursorKey
	cursorValue
)

// cursor is what the editor is typing: a table header, a key, or a value.
type cursor struct {
	kind cursorKind
	// path is the dotted path being typed, from the root. For a header or key
	// its last segment is the partial one; for a value it is the key's path.
	path []string
	// prefix is the partial text under the cursor.
	prefix string
}

// locate finds what is left of pos in text. It reads the closest table header
// above the line to give a key its table. A multi-line array whose line starts
// with a bracket and ends it is not taken for a header; other cases of
// multi-line values are not tracked.
func locate(text string, pos Position) cursor {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	if pos.Line < 0 || pos.Line >= len(lines) {
		return cursor{}
	}
	runes := []rune(lines[pos.Line])
	ch := max(0, min(pos.Character, len(runes)))
	before := string(runes[:ch])
	trimmed := strings.TrimLeft(before, " \t")
	if strings.HasPrefix(trimmed, "#") {
		return cursor{}
	}
	if strings.HasPrefix(trimmed, "[") {
		inner := strings.TrimLeft(trimmed, "[ \t")
		if strings.Contains(inner, "]") {
			return cursor{}
		}
		segs := splitKey(inner)
		return cursor{kind: cursorHeader, path: segs, prefix: segs[len(segs)-1]}
	}
	table := currentTable(lines[:pos.Line])
	if i := indexOutsideQuotes(before, '='); i >= 0 {
		return cursor{
			kind:   cursorValue,
			path:   append(table, splitKey(before[:i])...),
			prefix: strings.TrimSpace(before[i+1:]),
		}
	}
	segs := splitKey(trimmed)
	return cursor{kind: cursorKey, path: append(table, segs...), prefix: segs[len(segs)-1]}
}

// currentTable returns the path of the last table header in lines, or nil.
func currentTable(lines []string) []string {
	var table []string
	for _, line := range lines {
		t := strings.TrimSpace(line)
		if !strings.HasPrefix(t, "[") {
			continue
		}
		inner := strings.TrimLeft(t, "[ \t")
		end := indexOutsideQuotes(inner, ']')
		if end < 0 {
			continue
		}
		rest := strings.TrimLeft(inner[end:], "] \t")
		if rest != "" && !strings.HasPrefix(rest, "#") {
			continue
		}
		table = splitKey(inner[:end])
	}
	return table
}

// splitKey splits a dotted key into its segments, trimming blanks and quotes.
// It returns at least one segment, empty for empty input.
func splitKey(s string) []string {
	var segs []string
	var cur strings.Builder
	var quote rune
	for _, r := range s {
		switch {
		case quote != 0:
			if r == quote {
				quote = 0
			} else {
				cur.WriteRune(r)
			}
		case r == '"' || r == '\'':
			quote = r
		case r == '.':
			segs = append(segs, strings.TrimSpace(cur.String()))
			cur.Reset()
		default:
			cur.WriteRune(r)
		}
	}
	return append(segs, strings.TrimSpace(cur.String()))
}

// indexOutsideQuotes returns the byte index of the first c in s that is not
// inside a quoted string, or -1.
func indexOutsideQuotes(s string, c rune) int {
	var quote rune
	for i, r := range s {
		switch {
		case quote != 0:
			if r == quote {
				quote = 0
			}
		case r == '"' || r == '\'':
			quote = r
		case r == c:
			return i
		}
	}
	return -1
}

// isKeyChar reports whether r can be part of a bare TOML key.
func isKeyChar(r rune) bool {
	return r == '_' || r == '-' || unicode.IsLetter(r) || unicode.IsDigit(r)
}
