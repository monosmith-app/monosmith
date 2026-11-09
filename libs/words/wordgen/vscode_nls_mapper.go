package main

import (
	"bytes"
	"encoding/json"
	"strings"
)

// jsonString is s as a JSON string (also a valid TypeScript string literal),
// with no HTML escaping so a text reads as it was written.
func jsonString(s string) string {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(s)
	return strings.TrimRight(b.String(), "\n")
}

// positionalText is the text of key in the locale with {name} turned into {n}, n
// following the placeholder order of the English text.
func positionalText(r Resolved, locale, key string) string {
	return Positional(r.Texts[locale][key].Text, Placeholders(r.Texts[referenceLocale][key].Text))
}

// NlsJSON renders a package.nls JSON file of VS Code: key to text, with {0},
// {1}... for the placeholders. English holds every key; another locale holds only
// the keys it translates, because VS Code falls back to package.nls.json itself.
func NlsJSON(r Resolved, locale string) string {
	var lines []string
	for _, key := range r.Keys {
		if locale != referenceLocale && !r.Texts[locale][key].Own {
			continue
		}
		lines = append(lines, "  "+jsonString(key)+": "+jsonString(positionalText(r, locale, key)))
	}
	if len(lines) == 0 {
		return "{}\n"
	}
	return "{\n" + strings.Join(lines, ",\n") + "\n}\n"
}

// NlsFileName is the name of the package.nls file of a locale: VS Code wants the
// locale in lower case (package.nls.pt-br.json).
func NlsFileName(locale string) string {
	if locale == referenceLocale {
		return "package.nls.json"
	}
	return "package.nls." + strings.ToLower(locale) + ".json"
}
