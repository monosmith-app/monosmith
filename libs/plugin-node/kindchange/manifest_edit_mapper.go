package kindchange

import (
	"regexp"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
)

var (
	smithKind   = regexp.MustCompile(`("smith"[ \t\r\n]*:[ \t\r\n]*\{[^{}]*?"kind"[ \t]*:[ \t]*)"[^"]*"`)
	privateLine = regexp.MustCompile(`(?m)^([ \t]*)"private"[ \t]*:[ \t]*(?:true|false)[ \t]*(,?)[ \t]*\r?\n`)
	// anchors are the lines a new "private" goes after, as npm init writes it.
	anchors = []*regexp.Regexp{
		regexp.MustCompile(`(?m)^([ \t]*)"version"[ \t]*:[^\r\n]*,[ \t]*(\r?\n)`),
		regexp.MustCompile(`(?m)^([ \t]*)"name"[ \t]*:[^\r\n]*,[ \t]*(\r?\n)`),
	}
)

// setKind records the kind in the manifest's smith object.
func setKind(src, kind string) (string, error) {
	if smithKind.MatchString(src) {
		return smithKind.ReplaceAllString(src, `${1}"`+kind+`"`), nil
	}
	doc, err := jsondoc.ParseDocument([]byte(src))
	if err != nil {
		return src, err
	}
	doc.Object("smith", true).SetString("kind", kind)
	return string(doc.Bytes()), nil
}

// setPrivate makes the manifest private (a lib or an app is never published)
// or drops the field (a package is). It edits the line in place, so the rest
// of the file keeps its layout, and rewrites the document only for layouts it
// cannot edit that way (one line, or "private" as the last field).
func setPrivate(src string, private bool) (string, error) {
	if private {
		if loc := privateLine.FindStringSubmatchIndex(src); loc != nil {
			line := src[loc[0]:loc[1]]
			return src[:loc[0]] + regexp.MustCompile(`false`).ReplaceAllString(line, "true") + src[loc[1]:], nil
		}
		for _, anchor := range anchors {
			if loc := anchor.FindStringSubmatchIndex(src); loc != nil {
				indent, eol := src[loc[2]:loc[3]], src[loc[4]:loc[5]]
				return src[:loc[1]] + indent + `"private": true,` + eol + src[loc[1]:], nil
			}
		}
	} else {
		loc := privateLine.FindStringSubmatchIndex(src)
		if loc == nil && !strings.Contains(src, `"private"`) {
			return src, nil
		}
		if loc != nil && src[loc[4]:loc[5]] == "," {
			return src[:loc[0]] + src[loc[1]:], nil
		}
	}
	doc, err := jsondoc.ParseDocument([]byte(src))
	if err != nil {
		return src, err
	}
	if private {
		_ = doc.SetJSON("private", true)
	} else {
		doc.Delete("private")
	}
	return string(doc.Bytes()), nil
}
