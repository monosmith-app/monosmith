package workspace

import "strings"

// DefaultTagFormat is the release tag prefix of a project that sets none.
const DefaultTagFormat = "{name}@"

// ExpandTagFormat turns a [release] tag_format into the tag prefix of the
// project called name: {name} becomes the name and {scope} the workspace
// scope as written in [workspace] scope. An empty format is DefaultTagFormat.
func ExpandTagFormat(format, scope, name string) string {
	if format == "" {
		format = DefaultTagFormat
	}
	return strings.NewReplacer("{scope}", scope, "{name}", name).Replace(format)
}
