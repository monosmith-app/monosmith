package xmltext

import (
	"regexp"
	"strings"
)

func elementPattern(name string) *regexp.Regexp {
	q := regexp.QuoteMeta(name)
	return regexp.MustCompile(`(<` + q + `>)([^<]*)(</` + q + `>)`)
}

// Element returns the text of the first <name> element, and whether it is there.
func Element(src, name string) (string, bool) {
	m := elementPattern(name).FindStringSubmatch(src)
	if m == nil {
		return "", false
	}
	return strings.TrimSpace(m[2]), true
}

// SetElement sets the text of the first <name> element. When there is none it
// adds one as the first child of the first <parent> element, indented like the
// children already there; ok is false when there is no such parent either.
func SetElement(src, parent, name, value string) (out string, ok bool) {
	re := elementPattern(name)
	if loc := re.FindStringSubmatchIndex(src); loc != nil {
		return src[:loc[4]] + value + src[loc[5]:], true
	}
	open := regexp.MustCompile(`<` + regexp.QuoteMeta(parent) + `>[ \t]*(\r?\n)?`).FindStringSubmatchIndex(src)
	if open == nil {
		return src, false
	}
	eol := lineEnding(src)
	element := "<" + name + ">" + value + "</" + name + ">"
	if open[2] < 0 { // the parent is written on one line: <parent>...</parent>
		return src[:open[1]] + element + src[open[1]:], true
	}
	return src[:open[1]] + indentOfLine(src, open[1]) + element + eol + src[open[1]:], true
}

// RemoveElement deletes the first <name> element, and the line it was alone
// on; it reports whether the element was there.
func RemoveElement(src, name string) (string, bool) {
	q := regexp.QuoteMeta(name)
	re := regexp.MustCompile(`(?m)^[ \t]*<` + q + `>[^<]*</` + q + `>[ \t]*\r?\n?`)
	if loc := re.FindStringIndex(src); loc != nil {
		return src[:loc[0]] + src[loc[1]:], true
	}
	inline := regexp.MustCompile(`<` + q + `>[^<]*</` + q + `>`)
	if loc := inline.FindStringIndex(src); loc != nil {
		return src[:loc[0]] + src[loc[1]:], true
	}
	return src, false
}

// SetElementBefore is SetElement that, when it has to add the element, adds it
// on its own line just above the first <before> element, so it lands where a
// person would put it. Without a <before> element it adds it like SetElement.
func SetElementBefore(src, parent, name, value, before string) (out string, ok bool) {
	if _, found := Element(src, name); found {
		return SetElement(src, parent, name, value)
	}
	loc := regexp.MustCompile(`(?m)^([ \t]*)<` + regexp.QuoteMeta(before) + `>`).FindStringSubmatchIndex(src)
	if loc == nil {
		return SetElement(src, parent, name, value)
	}
	line := src[loc[2]:loc[3]] + "<" + name + ">" + value + "</" + name + ">" + lineEnding(src)
	return src[:loc[0]] + line + src[loc[0]:], true
}

// AppendElement sets the text of the first <name> element; when there is none
// it adds one as the last child of the first <parent> element, indented one
// step deeper than the closing tag. ok is false when there is no such parent
// written over several lines.
func AppendElement(src, parent, name, value string) (out string, ok bool) {
	if _, found := Element(src, name); found {
		return SetElement(src, parent, name, value)
	}
	loc := regexp.MustCompile(`(?m)^([ \t]*)</` + regexp.QuoteMeta(parent) + `>`).FindStringSubmatchIndex(src)
	if loc == nil {
		return src, false
	}
	line := src[loc[2]:loc[3]] + "  <" + name + ">" + value + "</" + name + ">" + lineEnding(src)
	return src[:loc[0]] + line + src[loc[0]:], true
}

// indentOfLine is the leading whitespace of the line starting at from.
func indentOfLine(src string, from int) string {
	end := from
	for end < len(src) && (src[end] == ' ' || src[end] == '\t') {
		end++
	}
	return src[from:end]
}

func lineEnding(src string) string {
	if strings.Contains(src, "\r\n") {
		return "\r\n"
	}
	return "\n"
}
