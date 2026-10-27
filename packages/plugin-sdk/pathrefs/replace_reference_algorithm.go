package pathrefs

import "strings"

// ReplaceReference replaces every whole reference old with new in text: an
// occurrence that is not part of a longer path or name ("libs/core" is not
// found in "libs/core-extra", "mylibs/core" or "../libs/core.rs"). It returns
// the new text and how many references changed.
func ReplaceReference(text, old, new string) (string, int) {
	if old == "" || old == new {
		return text, 0
	}
	var out strings.Builder
	count, from := 0, 0
	for {
		i := strings.Index(text[from:], old)
		if i < 0 {
			break
		}
		i += from
		end := i + len(old)
		if boundaryBefore(text, i) && boundaryAfter(text, end) {
			out.WriteString(text[from:i])
			out.WriteString(new)
			count++
		} else {
			out.WriteString(text[from:end])
		}
		from = end
	}
	if count == 0 {
		return text, 0
	}
	out.WriteString(text[from:])
	return out.String(), count
}

func boundaryBefore(text string, i int) bool {
	return i == 0 || !partOfReference(text[i-1], true)
}

func boundaryAfter(text string, end int) bool {
	return end >= len(text) || !partOfReference(text[end], false)
}

// partOfReference reports whether c continues a path or name. A slash after a
// reference continues it into a subpath, which is still the same project; a
// slash before one means a longer path leads up to it.
func partOfReference(c byte, before bool) bool {
	switch {
	case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9', c == '_', c == '-', c == '.':
		return true
	case before && (c == '/' || c == '@' || c == '\\'):
		return true
	}
	return false
}
