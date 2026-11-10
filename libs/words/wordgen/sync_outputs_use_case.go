package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/libs/words/surface"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// Options say what to read and where to write. Every directory is used as given.
type Options struct {
	// Words is the library folder: it holds catalog/<locale>/ and surface/<name>.json.
	Words string
	// VSCodeModule is the folder of the typed TypeScript module.
	VSCodeModule string
	// VSCodeNls is the folder of package.json, which gets package.nls*.json.
	VSCodeNls string
	// JetBrainsResources is the folder of the bundle files (messages*.properties).
	JetBrainsResources string
	// JetBrainsKotlin is the folder of the Kotlin accessor.
	JetBrainsKotlin string
	// Surfaces are the surfaces to write: "vscode" and/or "jetbrains".
	Surfaces []string
	// Check writes nothing and only reports the files that are not up to date.
	Check bool
}

// output is one generated file.
type output struct {
	path string
	text string
	// optional is true when the file may be absent in check mode: its surface
	// claims no key, so there is nothing to keep up to date yet.
	optional bool
}

// Sync writes the files of each surface in o, or in check mode lists the ones that
// differ. A file that is missing counts as stale, except one of a surface with no
// key (an empty surface may have no files yet). Line endings do not count: a
// checkout may have CRLF.
//
// Returns: the paths that differ or are missing.
// Errors: a catalogue or manifest that cannot be read, a key the catalogue lacks,
// two keys with one Kotlin constant, or a file that cannot be written.
func Sync(o Options) (stale []string, err error) {
	catalogs, err := words.LoadCatalog(os.DirFS(filepath.Join(o.Words, "catalog")))
	if err != nil {
		return nil, err
	}
	var files []output
	for _, name := range o.Surfaces {
		data, err := os.ReadFile(filepath.Join(o.Words, "surface", name+".json"))
		if err != nil {
			return nil, err
		}
		m, err := surface.Parse(data)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		r, err := Resolve(catalogs, m)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		more, err := render(name, r, o)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		for i := range more {
			more[i].optional = len(r.Keys) == 0
		}
		files = append(files, more...)
	}
	for _, f := range files {
		current, readErr := os.ReadFile(f.path)
		if readErr == nil && bytes.Equal(bytes.ReplaceAll(current, []byte("\r\n"), []byte("\n")), []byte(f.text)) {
			continue
		}
		if o.Check {
			if readErr != nil && f.optional {
				continue
			}
			stale = append(stale, f.path)
			continue
		}
		stale = append(stale, f.path)
		if err := os.MkdirAll(filepath.Dir(f.path), 0o755); err != nil {
			return stale, err
		}
		if err := os.WriteFile(f.path, []byte(f.text), 0o644); err != nil {
			return stale, err
		}
	}
	return stale, nil
}

// render is the files one surface gets.
func render(name string, r Resolved, o Options) ([]output, error) {
	switch name {
	case "vscode":
		files := []output{
			{path: filepath.Join(o.VSCodeModule, "smith-words.config.ts"), text: TypeScript(r)},
			{path: filepath.Join(o.VSCodeNls, NlsFileName(referenceLocale)), text: NlsJSON(r, referenceLocale)},
		}
		for _, locale := range r.Locales {
			files = append(files, output{path: filepath.Join(o.VSCodeNls, NlsFileName(locale)), text: NlsJSON(r, locale)})
		}
		return files, nil
	case "jetbrains":
		kotlin, err := Kotlin(r)
		if err != nil {
			return nil, err
		}
		files := []output{
			{path: filepath.Join(o.JetBrainsKotlin, "SmithWordsConfig.kt"), text: kotlin},
			{path: filepath.Join(o.JetBrainsResources, PropertiesFileName(referenceLocale)), text: Properties(r, referenceLocale)},
		}
		for _, locale := range r.Locales {
			files = append(files, output{path: filepath.Join(o.JetBrainsResources, PropertiesFileName(locale)), text: Properties(r, locale)})
		}
		return files, nil
	}
	return nil, fmt.Errorf("no generator for surface %q", name)
}
