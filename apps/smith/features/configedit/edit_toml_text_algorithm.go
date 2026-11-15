package configedit

import (
	"regexp"
	"strings"
)

// span is where a table's body or a key's value sits in the text.
type span struct{ start, end int }

// tableBody finds the body of [table]: from after its header line to the next
// table header (or the end). ok is false when the file has no such table.
func tableBody(src, table string) (span, bool) {
	header := regexp.MustCompile(`(?m)^[ \t]*\[` + regexp.QuoteMeta(table) + `\][ \t]*(#.*)?\r?\n?`)
	loc := header.FindStringIndex(src)
	if loc == nil {
		return span{}, false
	}
	end := len(src)
	if next := regexp.MustCompile(`(?m)^[ \t]*\[`).FindStringIndex(src[loc[1]:]); next != nil {
		end = loc[1] + next[0]
	}
	return span{loc[1], end}, true
}

// keyValue finds key's value inside body: from after `=` to the end of the
// value, which may span lines (an array), stopping before a trailing comment.
// line is the whole assignment line(s), newline included.
func keyValue(src string, body span, key string) (value, line span, ok bool) {
	re := regexp.MustCompile(`(?m)^[ \t]*` + regexp.QuoteMeta(key) + `[ \t]*=[ \t]*`)
	loc := re.FindStringIndex(src[body.start:body.end])
	if loc == nil {
		return span{}, span{}, false
	}
	start := body.start + loc[1]
	end := valueEnd(src, start)
	lineEnd := end
	for lineEnd < len(src) && src[lineEnd] != '\n' {
		lineEnd++
	}
	if lineEnd < len(src) {
		lineEnd++
	}
	value = span{start, end}
	for value.end > value.start && (src[value.end-1] == ' ' || src[value.end-1] == '\t' || src[value.end-1] == '\r') {
		value.end--
	}
	return value, span{body.start + loc[0], lineEnd}, true
}

// valueEnd scans a TOML value from i: through strings and brackets, up to the
// newline or comment that ends it at depth 0.
func valueEnd(src string, i int) int {
	depth := 0
	for i < len(src) {
		c := src[i]
		switch {
		case c == '"' || c == '\'':
			i++
			for i < len(src) && src[i] != c && src[i] != '\n' {
				if c == '"' && src[i] == '\\' {
					i++
				}
				i++
			}
		case c == '[' || c == '{':
			depth++
		case c == ']' || c == '}':
			depth--
		case (c == '\n' || c == '#') && depth <= 0:
			return i
		}
		i++
	}
	return len(src)
}

// SetKey sets key = value in [table], changing only that value: a comment on
// its line stays. A missing key is added at the end of the table's settings,
// a missing table at the end of the file.
func SetKey(src, table, key, value string) string {
	assignment := key + " = " + value + "\n"
	body, ok := tableBody(src, table)
	if !ok {
		if src != "" && !strings.HasSuffix(src, "\n") {
			src += "\n"
		}
		return src + "\n[" + table + "]\n" + assignment
	}
	if v, _, ok := keyValue(src, body, key); ok {
		return src[:v.start] + value + src[v.end:]
	}
	// After the table's last line that is not blank or a comment, so the
	// comments introducing the next table stay with it.
	at, offset := body.start, body.start
	for _, l := range strings.SplitAfter(src[body.start:body.end], "\n") {
		offset += len(l)
		if trimmed := strings.TrimSpace(l); trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			at = offset
		}
	}
	if at > 0 && src[at-1] != '\n' {
		assignment = "\n" + assignment
	}
	return src[:at] + assignment + src[at:]
}

// UnsetKey removes key from [table] (its whole line), so it takes its
// default; removed is false when it was not set.
func UnsetKey(src, table, key string) (out string, removed bool) {
	body, ok := tableBody(src, table)
	if !ok {
		return src, false
	}
	_, line, ok := keyValue(src, body, key)
	if !ok {
		return src, false
	}
	out = src[:line.start] + src[line.end:]
	// A table left with nothing in it goes too (with the blank line before
	// it), unless something was written about it.
	if body, _ := tableBody(out, table); strings.TrimSpace(out[body.start:body.end]) == "" {
		header := regexp.MustCompile(`(?m)(\r?\n)?^\[` + regexp.QuoteMeta(table) + `\][ \t]*(\r?\n|\z)`)
		if loc := header.FindStringIndex(out); loc != nil {
			out = out[:loc[0]] + out[loc[1]:]
		}
	}
	return out, true
}

// GetKey returns key's value text in [table] as written.
func GetKey(src, table, key string) (string, bool) {
	body, ok := tableBody(src, table)
	if !ok {
		return "", false
	}
	v, _, ok := keyValue(src, body, key)
	if !ok {
		return "", false
	}
	return src[v.start:v.end], true
}
