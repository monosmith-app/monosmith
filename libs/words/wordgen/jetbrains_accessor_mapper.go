package main

import (
	"fmt"
	"strconv"
	"strings"
)

// kotlinSay is the run-time helper of the object.
const kotlinSay = `    /**
     * The text of a key in the default locale, with the values in {0}, {1}...
     *
     * @param key one of the constants above.
     * @param values the value of each placeholder, in the order the constant lists.
     * @return the text; the key itself when the bundle lacks it.
     */
    fun say(
        key: String,
        vararg values: Any,
    ): String {
        val bundle = ResourceBundle.getBundle(BUNDLE)
        if (!bundle.containsKey(key)) return key
        val text = bundle.getString(key)
        return if (values.isEmpty()) text else MessageFormat.format(text, *values)
    }
}
`

// ConstantName is the Kotlin constant of a key: upper case, every other character
// an underscore (cmd.add-title becomes CMD_ADD_TITLE).
func ConstantName(key string) string {
	var b strings.Builder
	for _, r := range strings.ToUpper(key) {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	if name := b.String(); name != "" && name[0] >= '0' && name[0] <= '9' {
		return "_" + name
	}
	return b.String()
}

// Kotlin renders the object SmithWordsConfig of the JetBrains plugin: one typed
// constant per key and say(key, values), which reads the bundle and fills the
// {0}, {1}... in the order the constant's comment lists.
//
// Errors: two keys that give the same constant name.
func Kotlin(r Resolved) (string, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "package dev.monosmith.jetbrains.words\n\nimport java.text.MessageFormat\nimport java.util.ResourceBundle\n\n/**\n * %s\n */\nobject SmithWordsConfig {\n", generatedHeader)
	b.WriteString("    /** The base name of the resource bundle (messages.properties). */\n    const val BUNDLE = \"messages\"\n\n")
	seen := map[string]string{}
	for _, key := range r.Keys {
		name := ConstantName(key)
		if other, dup := seen[name]; dup {
			return "", fmt.Errorf("the keys %q and %q give the same constant %s", other, key, name)
		}
		seen[name] = key
		doc := "The key " + key + "."
		if order := Placeholders(r.Texts[referenceLocale][key].Text); len(order) > 0 {
			var pos []string
			for i, n := range order {
				pos = append(pos, "{"+strconv.Itoa(i)+"} is "+n)
			}
			doc += " Values: " + strings.Join(pos, ", ") + "."
		}
		fmt.Fprintf(&b, "    /** %s */\n    const val %s = %s\n\n", doc, name, jsonString(key))
	}
	b.WriteString(kotlinSay)
	return b.String(), nil
}
