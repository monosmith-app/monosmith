package words

import "strings"

// Say is the text for key, its {name} placeholders filled from pairs (name,
// value, name, value, ...). A key the catalog lacks gives the key itself, so a
// missing text is visible on screen and the guard test fails on it.
func Say(key string, pairs ...any) string {
	return fill(lookup(key), pairs)
}

// Count is the text for key in the plural form that fits n: key.one for 1, else
// key.other. {count} is filled with n; pairs fill the other placeholders. The
// rule is English's; a locale with other plural forms adds its own rule here.
func Count(key string, n int, pairs ...any) string {
	form := "other"
	if n == 1 {
		form = "one"
	}
	return fill(lookup(key+"."+form), append([]any{"count", n}, pairs...))
}

// fill replaces each {name} in text with the value paired with name. A
// placeholder with no pair stays as written, so the gap shows.
func fill(text string, pairs []any) string {
	if len(pairs) == 0 || !strings.Contains(text, "{") {
		return text
	}
	oldnew := make([]string, 0, len(pairs))
	for i := 0; i+1 < len(pairs); i += 2 {
		name, _ := pairs[i].(string)
		oldnew = append(oldnew, "{"+name+"}", sprint(pairs[i+1]))
	}
	return strings.NewReplacer(oldnew...).Replace(text)
}
