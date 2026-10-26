package configedit

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// HasTable is whether the smith.toml at root has [table], even empty.
func HasTable(root, table string) (bool, error) {
	raw, err := os.ReadFile(filepath.Join(root, workspace.FileName))
	if err != nil {
		return false, err
	}
	_, ok := tableBody(string(raw), table)
	return ok, nil
}

// AddTable adds an empty [table] at the end of the smith.toml at root, for a
// table whose presence is the setting (conventions.vertical-slices). A table
// that is already there is left as it is, with what it holds.
func AddTable(root, table string) error {
	return edit(root, func(src string) (string, error) {
		if _, ok := tableBody(src, table); ok {
			return src, nil
		}
		if src != "" && !strings.HasSuffix(src, "\n") {
			src += "\n"
		}
		return src + "\n[" + table + "]\n", nil
	})
}

// RemoveTable removes [table] and what it holds (its header and its settings);
// the comments above it, and every other table, stay.
func RemoveTable(root, table string) error {
	return edit(root, func(src string) (string, error) {
		body, ok := tableBody(src, table)
		if !ok {
			return "", fmt.Errorf("[%s] is not in %s", table, workspace.FileName)
		}
		header := regexp.MustCompile(`(?m)^[ \t]*\[` + regexp.QuoteMeta(table) + `\][ \t]*(#.*)?\r?\n?`)
		start := header.FindStringIndex(src)[0]
		// The blank line that set the table apart goes with it.
		switch before := src[:start]; {
		case strings.HasSuffix(before, "\r\n\r\n"):
			start -= 2
		case strings.HasSuffix(before, "\n\n"):
			start--
		}
		return src[:start] + src[body.end:], nil
	})
}
