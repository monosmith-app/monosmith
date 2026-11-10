package main

import (
	"fmt"
	"slices"
	"sort"

	"github.com/monosmith-app/monosmith/libs/words/surface"
)

// referenceLocale is the locale every other falls back to, key by key.
const referenceLocale = "en"

// Entry is the text of one key in one locale.
type Entry struct {
	// Text is the catalogue text, with {name} placeholders.
	Text string
	// Own is false when the locale lacks the key and Text is the English one.
	Own bool
}

// Resolved is what one surface uses of the catalogue.
type Resolved struct {
	// Keys are the claimed keys, sorted.
	Keys []string
	// Locales are the locales other than English that have at least one claimed
	// key of their own, sorted.
	Locales []string
	// Texts holds, for English and for each of Locales, an entry for every key:
	// a key the locale lacks carries the English text.
	Texts map[string]map[string]Entry
}

// Resolve picks the keys a surface claims out of the catalogues and applies the
// English fallback key by key.
//
// Parameters: catalogs is every locale's texts, by locale then key; m is the
// surface manifest.
// Errors: an exact key of the manifest that English lacks, and a translation that
// uses a placeholder its English text lacks (the call would not fill it).
func Resolve(catalogs map[string]map[string]string, m surface.Manifest) (Resolved, error) {
	english := catalogs[referenceLocale]
	var keys []string
	for _, key := range m.Keys {
		if _, ok := english[key]; !ok {
			return Resolved{}, fmt.Errorf("the surface claims %q, which the catalogue lacks", key)
		}
	}
	for key := range english {
		if m.Claims(key) {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	out := Resolved{Keys: keys, Texts: map[string]map[string]Entry{referenceLocale: {}}}
	for _, key := range keys {
		out.Texts[referenceLocale][key] = Entry{Text: english[key], Own: true}
	}
	var locales []string
	for locale := range catalogs {
		if locale != referenceLocale {
			locales = append(locales, locale)
		}
	}
	sort.Strings(locales)
	for _, locale := range locales {
		table := map[string]Entry{}
		anyOwn := false
		for _, key := range keys {
			text, own := catalogs[locale][key]
			if !own {
				table[key] = Entry{Text: english[key]}
				continue
			}
			for _, name := range Placeholders(text) {
				if !slices.Contains(Placeholders(english[key]), name) {
					return Resolved{}, fmt.Errorf("%s: %q uses {%s}, which the English text does not have", locale, key, name)
				}
			}
			table[key] = Entry{Text: text, Own: true}
			anyOwn = true
		}
		if anyOwn {
			out.Locales = append(out.Locales, locale)
			out.Texts[locale] = table
		}
	}
	return out, nil
}
