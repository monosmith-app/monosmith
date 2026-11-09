package main

import (
	"fmt"
	"strconv"
	"strings"
)

// propertiesEscape escapes s for a .properties file: backslash, line breaks and
// every non-ASCII rune as \uXXXX. A leading space of a value is escaped too. When
// key is true, spaces, ':' and '=' are escaped.
func propertiesEscape(s string, key bool) string {
	var b strings.Builder
	for i, r := range s {
		switch {
		case r == '\\':
			b.WriteString(`\\`)
		case r == '\n':
			b.WriteString(`\n`)
		case r == '\r':
			b.WriteString(`\r`)
		case r == '\t':
			b.WriteString(`\t`)
		case r == ' ' && (key || i == 0):
			b.WriteString(`\ `)
		case key && (r == ':' || r == '='):
			b.WriteString(`\` + string(r))
		case r > 126 || r < 32:
			for _, u := range utf16Units(r) {
				fmt.Fprintf(&b, `\u%04x`, u)
			}
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// utf16Units are the UTF-16 code units of r, which a \u escape needs.
func utf16Units(r rune) []int {
	if r < 0x10000 {
		return []int{int(r)}
	}
	r -= 0x10000
	return []int{0xD800 + int(r>>10), 0xDC00 + int(r&0x3FF)}
}

// Properties renders a messages.properties resource bundle of JetBrains: key to
// text, with {0}, {1}... for the placeholders, in the order of the English text
// (a comment above each such key lists the order). A text with placeholders has
// its apostrophes doubled, which MessageFormat needs; a text without placeholders
// is read as it is. English holds every key; another locale holds only the keys
// it translates, because a resource bundle falls back to the base file itself.
func Properties(r Resolved, locale string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n", generatedHeader)
	for _, key := range r.Keys {
		if locale != referenceLocale && !r.Texts[locale][key].Own {
			continue
		}
		order := Placeholders(r.Texts[referenceLocale][key].Text)
		text := r.Texts[locale][key].Text
		if len(order) > 0 {
			var pos []string
			for i, name := range order {
				pos = append(pos, "{"+strconv.Itoa(i)+"}="+name)
			}
			fmt.Fprintf(&b, "# %s\n", strings.Join(pos, " "))
			text = strings.ReplaceAll(positionalText(r, locale, key), "'", "''")
		}
		fmt.Fprintf(&b, "%s=%s\n", propertiesEscape(key, true), propertiesEscape(text, false))
	}
	return b.String()
}

// PropertiesFileName is the name of the bundle file of a locale
// (messages_pt_BR.properties).
func PropertiesFileName(locale string) string {
	if locale == referenceLocale {
		return "messages.properties"
	}
	return "messages_" + strings.ReplaceAll(locale, "-", "_") + ".properties"
}
