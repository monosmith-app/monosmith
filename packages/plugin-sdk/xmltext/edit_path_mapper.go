package xmltext

import (
	"encoding/xml"
	"io"
	"strings"
)

// span is where an element sits in the text.
type span struct {
	openEnd     int // where its start tag ends: the text starts here
	closeStart  int // where its end tag begins: the text ends here
	hasChildren bool
}

// locate reads src once and returns the element at each prefix of path (the
// first element named so, among the direct children of the one before). The
// first name of path is the document's root element. ok is false when the
// text is not well-formed or an end tag is written in a form this edit does
// not handle (<name/> or a space inside the tag).
func locate(src string, path []string) (spans []*span, ok bool) {
	d := xml.NewDecoder(strings.NewReader(src))
	spans = make([]*span, len(path))
	var match []*span // the element at each depth of the stack when it is on the path, else nil
	for {
		tok, err := d.Token()
		if err == io.EOF {
			return spans, len(match) == 0
		}
		if err != nil {
			return nil, false
		}
		switch t := tok.(type) {
		case xml.StartElement:
			depth := len(match)
			if depth > 0 && match[depth-1] != nil {
				match[depth-1].hasChildren = true
			}
			var sp *span
			if depth < len(path) && path[depth] == t.Name.Local && spans[depth] == nil && (depth == 0 || match[depth-1] != nil) {
				sp = &span{openEnd: int(d.InputOffset())}
				spans[depth] = sp
			}
			match = append(match, sp)
		case xml.EndElement:
			sp := match[len(match)-1]
			match = match[:len(match)-1]
			if sp == nil {
				continue
			}
			end := int(d.InputOffset())
			tag := "</" + t.Name.Local + ">"
			sp.closeStart = end - len(tag)
			if sp.closeStart < sp.openEnd || src[sp.closeStart:end] != tag {
				return nil, false
			}
		}
	}
}

// Path returns the text of the element at path (project, build, directory: each
// a direct child of the one before) and whether it is there.
func Path(src string, path ...string) (string, bool) {
	spans, ok := locate(src, path)
	if !ok || len(path) == 0 {
		return "", false
	}
	sp := spans[len(path)-1]
	if sp == nil || sp.hasChildren {
		return "", false
	}
	return strings.TrimSpace(src[sp.openEnd:sp.closeStart]), true
}

// SetPath sets the text of the element at path, adding the elements that are
// missing below the deepest one that exists, indented like the file. It
// keeps every other byte. ok is false when the file is not well-formed, the
// first name is not its root element, or an element on the way is written in
// a form this edit does not handle.
func SetPath(src, value string, path ...string) (string, bool) {
	spans, ok := locate(src, path)
	if !ok || len(path) == 0 || spans[0] == nil {
		return src, false
	}
	if last := spans[len(path)-1]; last != nil {
		if last.hasChildren {
			return src, false
		}
		return src[:last.openEnd] + value + src[last.closeStart:], true
	}
	deepest := 0
	for i, sp := range spans {
		if sp != nil {
			deepest = i
		}
	}
	anchor := spans[deepest]
	eol := lineEnding(src)
	unit := indentUnit(src, spans[0].openEnd)
	base := lineIndent(src, anchor.closeStart)
	missing := path[deepest+1:]
	// The new elements go on lines of their own, just above the closing tag of
	// the deepest element that exists.
	var b strings.Builder
	for i, name := range missing {
		b.WriteString(base + strings.Repeat(unit, i+1) + "<" + name + ">")
		if i < len(missing)-1 {
			b.WriteString(eol)
		}
	}
	b.WriteString(value)
	for i := len(missing) - 1; i >= 0; i-- {
		if i < len(missing)-1 {
			b.WriteString(base + strings.Repeat(unit, i+1))
		}
		b.WriteString("</" + missing[i] + ">" + eol)
	}
	at := anchor.closeStart
	if strings.TrimSpace(src[lineStart(src, at):at]) == "" {
		at = lineStart(src, at)
		return src[:at] + b.String() + src[at:], true
	}
	return src[:at] + eol + b.String() + base + src[at:], true
}

func lineStart(src string, i int) int {
	for i > 0 && src[i-1] != '\n' {
		i--
	}
	return i
}

// lineIndent is the leading whitespace of the line that holds offset i.
func lineIndent(src string, i int) string {
	return indentOfLine(src, lineStart(src, i))
}

// indentUnit is the step of the file's indentation: the leading whitespace of
// the first indented line after the root element's start tag, four spaces
// when there is none.
func indentUnit(src string, from int) string {
	rest := src[from:]
	if nl := strings.IndexByte(rest, '\n'); nl >= 0 {
		if ws := indentOfLine(rest, nl+1); ws != "" && nl+1+len(ws) < len(rest) && rest[nl+1+len(ws)] == '<' {
			return ws
		}
	}
	return "    "
}
