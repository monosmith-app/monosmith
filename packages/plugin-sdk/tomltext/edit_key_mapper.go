package tomltext

import (
	"regexp"
	"strconv"
	"strings"
)

// Raw returns the value text of key in [table], as written ("false",
// "\"lib\""), and whether it is there.
func Raw(src, table, key string) (string, bool) {
	start, end, ok := body(src, table)
	if !ok {
		return "", false
	}
	m := keyLine(key).FindStringSubmatch(src[start:end])
	if m == nil {
		return "", false
	}
	return strings.TrimSpace(m[2]), true
}

// SetString sets key = "value" in [table]: it replaces the key's value,
// adds the key under the table's header, or adds the table at the end.
func SetString(src, table, key, value string) string {
	return SetRaw(src, table, key, strconv.Quote(value))
}

// SetRaw sets key to a value written as TOML text.
func SetRaw(src, table, key, raw string) string {
	eol := lineEnding(src)
	start, end, ok := body(src, table)
	if !ok {
		if src != "" && !strings.HasSuffix(src, "\n") {
			src += eol
		}
		if src != "" {
			src += eol
		}
		return src + "[" + table + "]" + eol + key + " = " + raw + eol
	}
	re := keyLine(key)
	if loc := re.FindStringSubmatchIndex(src[start:end]); loc != nil {
		valueStart, valueEnd := start+loc[4], start+loc[5]
		return src[:valueStart] + raw + src[valueEnd:]
	}
	// A new key goes after the table's last one, where a person would add it.
	last := start + len(strings.TrimRight(src[start:end], " \t\r\n"))
	if last == start {
		return src[:start] + key + " = " + raw + eol + src[start:]
	}
	return src[:last] + eol + key + " = " + raw + src[last:]
}

// Remove deletes key from [table]; it reports whether the key was there.
func Remove(src, table, key string) (string, bool) {
	start, end, ok := body(src, table)
	if !ok {
		return src, false
	}
	loc := keyLine(key).FindStringIndex(src[start:end])
	if loc == nil {
		return src, false
	}
	return src[:start+loc[0]] + src[start+loc[1]:], true
}

// body is the span of text under [table]'s header, up to the next header.
func body(src, table string) (start, end int, ok bool) {
	header := regexp.MustCompile(`(?m)^[ \t]*\[` + regexp.QuoteMeta(table) + `\][ \t]*(#[^\r\n]*)?\r?\n?`)
	loc := header.FindStringIndex(src)
	if loc == nil {
		return 0, 0, false
	}
	start, end = loc[1], len(src)
	if next := anyHeader.FindStringIndex(src[start:]); next != nil {
		end = start + next[0]
	}
	return start, end, true
}

var anyHeader = regexp.MustCompile(`(?m)^[ \t]*\[`)

// keyLine matches the whole line "key = value # comment". Group 2 is the
// value; the index form of group 2 sits at positions 4 and 5.
func keyLine(key string) *regexp.Regexp {
	return regexp.MustCompile(`(?m)^([ \t]*` + regexp.QuoteMeta(key) + `[ \t]*=[ \t]*)("[^"\r\n]*"|'[^'\r\n]*'|[^#\r\n]*?)([ \t]*(?:#[^\r\n]*)?(?:\r?\n|\z))`)
}

func lineEnding(src string) string {
	if strings.Contains(src, "\r\n") {
		return "\r\n"
	}
	return "\n"
}
