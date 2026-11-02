package jsondoc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
)

// errNotObject means the text is not a JSON object (comments allowed).
var errNotObject = errors.New("the file is not a JSON object")

// entrySpan is where one top-level entry of a JSON object sits in its text.
type entrySpan struct {
	key                  string
	valueStart, valueEnd int
}

// objectLayout is the shape of the top-level object of a JSON text that may
// hold comments and trailing commas (a tsconfig does).
type objectLayout struct {
	entries []entrySpan
	// open and close are the offsets of the braces.
	open, close int
}

// StripComments removes the comments and the trailing commas of a JSON text, so a
// strict JSON decoder reads it. Strings are kept as they are.
func StripComments(src []byte) []byte {
	var out bytes.Buffer
	for i := 0; i < len(src); {
		switch {
		case src[i] == '"':
			j := stringEnd(src, i)
			out.Write(src[i:j])
			i = j
		case src[i] == '/' && i+1 < len(src) && (src[i+1] == '/' || src[i+1] == '*'):
			i = skipComment(src, i)
			out.WriteByte(' ')
		case src[i] == ',':
			j := skipSpace(src, i+1)
			if j < len(src) && (src[j] == '}' || src[j] == ']') {
				i++
				continue
			}
			out.WriteByte(',')
			i++
		default:
			out.WriteByte(src[i])
			i++
		}
	}
	return out.Bytes()
}

// stringEnd returns the offset after the string that starts at i.
func stringEnd(src []byte, i int) int {
	for j := i + 1; j < len(src); j++ {
		switch src[j] {
		case '\\':
			j++
		case '"':
			return j + 1
		}
	}
	return len(src)
}

// skipComment returns the offset after the comment that starts at i.
func skipComment(src []byte, i int) int {
	if src[i+1] == '/' {
		if n := bytes.IndexByte(src[i:], '\n'); n >= 0 {
			return i + n
		}
		return len(src)
	}
	if n := bytes.Index(src[i+2:], []byte("*/")); n >= 0 {
		return i + 2 + n + 2
	}
	return len(src)
}

// skipSpace returns the offset of the next byte that is not white space or part
// of a comment.
func skipSpace(src []byte, i int) int {
	for i < len(src) {
		switch {
		case src[i] == ' ' || src[i] == '\t' || src[i] == '\n' || src[i] == '\r':
			i++
		case src[i] == '/' && i+1 < len(src) && (src[i+1] == '/' || src[i+1] == '*'):
			i = skipComment(src, i)
		default:
			return i
		}
	}
	return i
}

// valueEnd returns the offset after the value that starts at i.
func valueEnd(src []byte, i int) int {
	switch src[i] {
	case '"':
		return stringEnd(src, i)
	case '{', '[':
		depth := 0
		for j := i; j < len(src); {
			switch {
			case src[j] == '"':
				j = stringEnd(src, j)
				continue
			case src[j] == '/' && j+1 < len(src) && (src[j+1] == '/' || src[j+1] == '*'):
				j = skipComment(src, j)
				continue
			case src[j] == '{' || src[j] == '[':
				depth++
			case src[j] == '}' || src[j] == ']':
				depth--
				if depth == 0 {
					return j + 1
				}
			}
			j++
		}
		return len(src)
	}
	j := i
	for j < len(src) && !strings.ContainsRune(" \t\r\n,}]/", rune(src[j])) {
		j++
	}
	return j
}

// layoutOf finds the top-level entries of the object in src.
func layoutOf(src []byte) (objectLayout, error) {
	i := skipSpace(src, 0)
	if i >= len(src) || src[i] != '{' {
		return objectLayout{}, errNotObject
	}
	l := objectLayout{open: i}
	i = skipSpace(src, i+1)
	for i < len(src) && src[i] != '}' {
		if src[i] != '"' {
			return objectLayout{}, errNotObject
		}
		keyEnd := stringEnd(src, i)
		var key string
		if err := json.Unmarshal(src[i:keyEnd], &key); err != nil {
			return objectLayout{}, errNotObject
		}
		i = skipSpace(src, keyEnd)
		if i >= len(src) || src[i] != ':' {
			return objectLayout{}, errNotObject
		}
		i = skipSpace(src, i+1)
		if i >= len(src) {
			return objectLayout{}, errNotObject
		}
		end := valueEnd(src, i)
		l.entries = append(l.entries, entrySpan{key: key, valueStart: i, valueEnd: end})
		i = skipSpace(src, end)
		if i < len(src) && src[i] == ',' {
			i = skipSpace(src, i+1)
		}
	}
	if i >= len(src) {
		return objectLayout{}, errNotObject
	}
	l.close = i
	return l, nil
}

// SetKeysInPlace sets the top-level keys of the JSON object in src to the given
// values and returns the new text. A key that exists keeps its place, and
// everything outside the value of a key stays byte for byte: comments, the
// order of the other keys, the indentation. A key that does not exist is added
// at the end. It returns an error when src is not a JSON object.
func SetKeysInPlace(src []byte, keys map[string]any) ([]byte, error) {
	for _, k := range sortedKeysOf(keys) {
		l, err := layoutOf(src)
		if err != nil {
			return nil, err
		}
		indent := indentOf(src, l)
		value, err := formatValue(keys[k], indent)
		if err != nil {
			return nil, err
		}
		src = setOne(src, l, k, value, indent)
	}
	return src, nil
}

func sortedKeysOf(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func setOne(src []byte, l objectLayout, key, value, indent string) []byte {
	for _, e := range l.entries {
		if e.key == key {
			return splice(src, e.valueStart, e.valueEnd, value)
		}
	}
	quoted, _ := json.Marshal(key)
	entry := indent + string(quoted) + ": " + value
	if len(l.entries) == 0 {
		return splice(src, l.open+1, l.close, "\n"+entry+"\n")
	}
	last := l.entries[len(l.entries)-1]
	return splice(src, last.valueEnd, last.valueEnd, ",\n"+entry)
}

func splice(src []byte, from, to int, with string) []byte {
	out := make([]byte, 0, len(src)+len(with))
	out = append(out, src[:from]...)
	out = append(out, with...)
	return append(out, src[to:]...)
}

// indentOf is the white space before the first key, or two spaces.
func indentOf(src []byte, l objectLayout) string {
	if len(l.entries) == 0 {
		return "  "
	}
	start := l.entries[0].valueStart
	line := bytes.LastIndexByte(src[:start], '\n')
	if line < 0 {
		return "  "
	}
	rest := string(src[line+1 : start])
	trimmed := strings.TrimLeft(rest, " \t")
	if indent := rest[:len(rest)-len(trimmed)]; indent != "" && strings.HasPrefix(trimmed, "\"") {
		return indent
	}
	return "  "
}

// formatValue renders a value as indented JSON; every line after the first
// starts with indent.
func formatValue(v any, indent string) (string, error) {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	enc.SetIndent(indent, "  ")
	if err := enc.Encode(v); err != nil {
		return "", fmt.Errorf("a key value must be JSON: %w", err)
	}
	return strings.TrimRight(b.String(), "\n"), nil
}
