package words

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"
	"sync/atomic"

	"github.com/monosmith-app/monosmith/libs/words/catalog"
)

// referenceLocale is the locale every other falls back to, key by key.
const referenceLocale = "en"

// catalogs is every locale's texts, by locale then key; it is built once.
var catalogs = mustLoad(catalog.Files)

// active is the locale in use, set by Use.
var active atomic.Value

func init() { Use(os.Getenv("SMITH_LANG")) }

// Use makes locale the one Say and Count read ("pt-BR" and "pt" both choose the
// pt catalog when there is one). A locale with no catalog, or "", keeps English.
func Use(locale string) {
	active.Store(match(locale))
}

// match is the catalog locale that fits a language tag: the tag itself, else
// its language ("pt-BR" gives "pt"), else the reference locale.
func match(tag string) string {
	tag = strings.ReplaceAll(strings.TrimSpace(tag), "_", "-")
	if i := strings.IndexAny(tag, ".@"); i >= 0 {
		tag = tag[:i]
	}
	for _, candidate := range []string{tag, strings.SplitN(tag, "-", 2)[0]} {
		if _, ok := catalogs[candidate]; ok && candidate != "" {
			return candidate
		}
	}
	return referenceLocale
}

// lookup is the text of key in the active locale, else in English, else key.
func lookup(key string) string {
	if locale, _ := active.Load().(string); locale != "" {
		if text, ok := catalogs[locale][key]; ok {
			return text
		}
	}
	if text, ok := catalogs[referenceLocale][key]; ok {
		return text
	}
	return key
}

// Reference is the English texts, by key: the catalog every surface falls back
// to. The map is a copy, so a caller may keep or change it.
func Reference() map[string]string {
	out := make(map[string]string, len(catalogs[referenceLocale]))
	for key, text := range catalogs[referenceLocale] {
		out[key] = text
	}
	return out
}

// mustLoad is LoadCatalog for the catalog that ships inside the binary: a file
// that is not a JSON object of strings, or a key two files both define, is a
// build error, so it panics.
func mustLoad(files fs.FS) map[string]map[string]string {
	out, err := LoadCatalog(files)
	if err != nil {
		panic("words: " + err.Error())
	}
	return out
}

// LoadCatalog reads <locale>/*.json from files into one map per locale.
//
// Parameters: files is the catalogue folder (libs/words/catalog: the embedded one,
// or the folder on disk for the generator and its tests).
// Errors: a file that is not a JSON object of strings, or a key that two files
// of one locale both define.
func LoadCatalog(files fs.FS) (map[string]map[string]string, error) {
	out := map[string]map[string]string{}
	err := fs.WalkDir(files, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || path.Ext(p) != ".json" {
			return err
		}
		locale := path.Base(path.Dir(p))
		raw, err := fs.ReadFile(files, p)
		if err != nil {
			return err
		}
		var texts map[string]string
		if err := json.Unmarshal(raw, &texts); err != nil {
			return fmt.Errorf("%s: %w", p, err)
		}
		if out[locale] == nil {
			out[locale] = map[string]string{}
		}
		for key, text := range texts {
			if _, dup := out[locale][key]; dup {
				return fmt.Errorf("%s: key %q is also defined in another file of %s", p, key, locale)
			}
			out[locale][key] = text
		}
		return nil
	})
	return out, err
}
