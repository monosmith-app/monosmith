package npmbinaries

import (
	"maps"
	"slices"
	"strings"
)

// folder is a package's directory name: @acme/cli → acme-cli.
func folder(name string) string {
	return strings.ReplaceAll(strings.TrimPrefix(name, "@"), "/", "-")
}

func platformNames(platforms map[string]string) []string {
	return slices.Sorted(maps.Values(platforms))
}
