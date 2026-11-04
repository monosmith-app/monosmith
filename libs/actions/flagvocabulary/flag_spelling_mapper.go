package flagvocabulary

import "strings"

// Arg returns the spelling a command line smith builds should use for a
// concept: the shorthand ("-p") when it has one, else the long form
// ("--no-install"). The TUI, the guided forms and the hints take every flag name
// from here, so they cannot drift from the commands.
//
// It panics when id is not a concept (see Find).
func Arg(id string) string {
	c := Find(id)
	if c.Short != "" {
		return "-" + c.Short
	}
	return "--" + c.Flag
}

// Long returns the canonical long form of a concept ("--project"). It panics
// when id is not a concept.
func Long(id string) string { return "--" + Find(id).Flag }

// Is reports whether arg, one word of a command line, is a spelling of the
// concept: the canonical shorthand or long form, or a synonym ("-p", "--project",
// "-w", "--workspace", "--projects"). It panics when id is not a concept.
func Is(id, arg string) bool {
	c := Find(id)
	if arg == "--"+c.Flag || (c.Short != "" && arg == "-"+c.Short) {
		return true
	}
	for _, s := range c.Synonyms {
		if arg == "--"+s.Name || (s.Short != "" && arg == "-"+s.Short) {
			return true
		}
	}
	return false
}

// Inline reports whether arg is the "--name=value" form of a long spelling of
// the concept, and returns the value. It panics when id is not a concept.
func Inline(id, arg string) (value string, ok bool) {
	name, value, found := strings.Cut(arg, "=")
	if !found || !strings.HasPrefix(name, "--") {
		return "", false
	}
	return value, Is(id, name)
}
