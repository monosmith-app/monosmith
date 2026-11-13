package projectnaming

import (
	"strings"
	"unicode"
)

// words splits text into lowercase runs of letters and digits.
func words(text string) []string {
	return strings.FieldsFunc(strings.ToLower(text), func(c rune) bool { return !unicode.IsLetter(c) && !unicode.IsDigit(c) })
}

func startsWithDigit(s string) bool {
	for _, c := range s {
		return unicode.IsDigit(c)
	}
	return false
}

// Kebab is text in lowercase kebab-case: letters and digits kept, every other
// run of characters one dash, no dash at either end.
//
// Parameters: text is any text, such as a workspace name.
// Returns: the kebab-case form, empty when text has no letter or digit.
func Kebab(text string) string { return strings.Join(words(text), "-") }

// Snake is text in lowercase snake_case, starting with a letter: a leading
// digit gets an x (3d becomes x3d).
//
// Parameters: text is any text.
// Returns: the snake_case form, empty when text has no letter or digit.
func Snake(text string) string {
	out := strings.Join(words(text), "_")
	if startsWithDigit(out) {
		return "x" + out
	}
	return out
}

// Pascal is text in UpperCamelCase, starting with a letter: an empty result
// or a leading digit gets an M (2fa becomes M2fa).
//
// Parameters: text is any text.
// Returns: the UpperCamelCase form, never empty.
func Pascal(text string) string {
	var b strings.Builder
	for _, w := range words(text) {
		r := []rune(w)
		b.WriteRune(unicode.ToUpper(r[0]))
		b.WriteString(string(r[1:]))
	}
	if b.Len() == 0 || startsWithDigit(b.String()) {
		return "M" + b.String()
	}
	return b.String()
}

// Flat is text reduced to lowercase letters and digits with no separator,
// starting with a letter: a Java package segment (web-api becomes webapi).
//
// Parameters: text is any text.
// Returns: the flat form, empty when text has no letter or digit.
func Flat(text string) string {
	out := strings.Join(words(text), "")
	if startsWithDigit(out) {
		return "x" + out
	}
	return out
}

// vbaIdentifier is a VBA module name: ASCII letters, digits and underscores,
// starting with a letter. Case is kept.
func vbaIdentifier(text string) string {
	var b strings.Builder
	for _, c := range text {
		if c == '_' || (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
			b.WriteRune(c)
		} else {
			b.WriteByte('_')
		}
	}
	out := b.String()
	if out == "" || !strings.ContainsAny(out[:1], "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ") {
		return "M" + out
	}
	return out
}

// pathElement is a Go module path element: lowercase letters, digits, dots,
// underscores and dashes, a run of anything else one dash.
func pathElement(text string) string {
	var b strings.Builder
	for _, c := range strings.ToLower(strings.TrimSpace(text)) {
		switch {
		case c >= 'a' && c <= 'z', c >= '0' && c <= '9', c == '.', c == '_', c == '-':
			b.WriteRune(c)
		case b.Len() > 0 && !strings.HasSuffix(b.String(), "-"):
			b.WriteByte('-')
		}
	}
	return strings.Trim(b.String(), "-.")
}
