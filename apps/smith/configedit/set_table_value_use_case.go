package configedit

import (
	"strconv"
	"strings"
)

// SetTableValue writes key = "value" under [table] in the smith.toml at root,
// for a table of named entries (plugins.external.kiwi) that Set, which edits
// the fixed tables, does not reach. Only that line changes; the whole file
// must still load before it is written.
func SetTableValue(root, table, key, value string) error {
	return edit(root, func(src string) (string, error) {
		return SetKey(src, table, key, strconv.Quote(value)), nil
	})
}

// QuoteKey is a table-name part as TOML spells it: bare when it can be,
// quoted otherwise.
func QuoteKey(name string) string {
	if name != "" && strings.Trim(name, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-") == "" {
		return name
	}
	return strconv.Quote(name)
}
