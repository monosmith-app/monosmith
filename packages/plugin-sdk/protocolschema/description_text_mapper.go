package protocolschema

import "strings"

// Paragraphs turns a Go doc comment into running text: the lines of a
// paragraph are joined with one space, and paragraphs are separated by a blank
// line. A comment that is absent gives "".
func Paragraphs(comment string) string {
	var paragraphs []string
	for _, block := range strings.Split(strings.TrimSpace(comment), "\n\n") {
		var words []string
		for _, line := range strings.Split(block, "\n") {
			if t := strings.TrimSpace(line); t != "" {
				words = append(words, t)
			}
		}
		if len(words) > 0 {
			paragraphs = append(paragraphs, strings.Join(words, " "))
		}
	}
	return strings.Join(paragraphs, "\n\n")
}

// WithoutSubject drops the Go name a doc comment starts with ("RoleLanguage
// finds projects": "Finds projects"), since the schema's reader sees the
// constant's value, not its name.
func WithoutSubject(name, doc string) string {
	rest, found := strings.CutPrefix(doc, name+" ")
	if !found || rest == "" {
		return doc
	}
	return strings.ToUpper(rest[:1]) + rest[1:]
}
