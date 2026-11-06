package main

import (
	"fmt"
	"strings"
)

// generatedHeader tells a reader that the file is written by the generator.
const generatedHeader = "Generated from libs/words/catalog by `go run ./wordgen` in libs/words. Do not edit: change the catalogue and run it again."

// sayFunction is the run-time helper of the module.
const sayFunction = "/**\n" +
	" * The text of a key, each {name} filled with the value of that name. A name with\n" +
	" * no value stays as written, so the gap shows.\n" +
	" *\n" +
	" * @param key the key of the text.\n" +
	" * @param values the value of each placeholder, by name.\n" +
	" * @param locale the locale tag (pt-BR): its own table, else its language, else English.\n" +
	" */\n" +
	"export function say(key: WordKey, values: Record<string, string | number> = {}, locale = 'en'): string {\n" +
	"  const tag = locale.toLowerCase()\n" +
	"  const texts = translations[tag] ?? translations[tag.split('-')[0] ?? ''] ?? english\n" +
	"  return texts[key].replace(/\\{([a-z][A-Za-z0-9]*)\\}/g, (whole: string, name: string) =>\n" +
	"    name in values ? String(values[name]) : whole,\n" +
	"  )\n" +
	"}\n"

// TypeScript renders the typed module of the VS Code extension. `nls` maps each
// key to its %key% reference for package.json; `say` gives a text at run time,
// with each {name} filled by name, in the locale asked for (English when the
// extension has no such locale). The module imports nothing, so a unit test can
// load it without VS Code.
func TypeScript(r Resolved) string {
	var b strings.Builder
	fmt.Fprintf(&b, "// %s\n\n", generatedHeader)
	b.WriteString("/** The %key% reference of each text, for package.json. */\nexport const nls = {")
	for _, key := range r.Keys {
		fmt.Fprintf(&b, "\n  %s: %s,", jsonString(key), jsonString("%"+key+"%"))
	}
	if len(r.Keys) > 0 {
		b.WriteString("\n")
	}
	b.WriteString("} as const\n\n/** Every key of the catalogue this extension uses. */\nexport type WordKey = keyof typeof nls\n\n")
	b.WriteString("type Texts = Record<WordKey, string>\n\n/** The English texts. */\nconst english: Texts = ")
	writeTable(&b, r, referenceLocale, "")
	b.WriteString("\n\n/** The other locales, each complete: a key the locale lacks has the English text. */\nconst translations: Partial<Record<string, Texts>> = {")
	for _, locale := range r.Locales {
		fmt.Fprintf(&b, "\n  %s: ", jsonString(strings.ToLower(locale)))
		writeTable(&b, r, locale, "  ")
		b.WriteString(",")
	}
	if len(r.Locales) > 0 {
		b.WriteString("\n")
	}
	b.WriteString("}\n\n")
	b.WriteString(sayFunction)
	return b.String()
}

// writeTable writes the {...} of one locale's texts, each line indented.
func writeTable(b *strings.Builder, r Resolved, locale, indent string) {
	b.WriteString("{")
	for _, key := range r.Keys {
		fmt.Fprintf(b, "\n%s  %s: %s,", indent, jsonString(key), jsonString(r.Texts[locale][key].Text))
	}
	if len(r.Keys) > 0 {
		b.WriteString("\n" + indent)
	}
	b.WriteString("}")
}
