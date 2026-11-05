package guide

import (
	"strings"
	"sync"

	"github.com/monosmith-app/monosmith/apps/smith/configedit"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pathfinder"
)

// pathWords are the words of an argument or option name that say it holds a
// path (`<sample-folder>`, `<file-or-url>`, `--cache-dir`, `--recipe-file`,
// `--extends`).
var pathWords = map[string]bool{"dir": true, "directory": true, "folder": true, "file": true, "path": true, "recipe": true, "extends": true}

// asksForPath reports whether a generated form's argument or option, by its
// name, takes a path, so its question shows the path finder.
func asksForPath(name string) bool {
	for _, w := range strings.FieldsFunc(name, func(r rune) bool { return r == '-' || r == '_' || r == ':' }) {
		if pathWords[strings.ToLower(w)] {
			return true
		}
	}
	return false
}

// settingEntries are the settings of smith.toml the path finder offers beside
// the files, with the first sentence of what each does.
var settingEntries = sync.OnceValue(func() []pathfinder.Setting {
	fields, err := configedit.Fields()
	if err != nil {
		return nil
	}
	out := make([]pathfinder.Setting, 0, len(fields))
	for _, f := range fields {
		hint, _, _ := strings.Cut(f.Description, ". ")
		out = append(out, pathfinder.Setting{Path: f.Path, Hint: strings.TrimSuffix(hint, ".")})
	}
	return out
})
