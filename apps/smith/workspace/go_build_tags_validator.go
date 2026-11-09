package workspace

import (
	"fmt"
	"regexp"
	"sort"
)

var goTagPattern = regexp.MustCompile(`^[A-Za-z0-9_.]+$`)

// goTagKeys are the keys of [projects.<name>.settings.go] that hold build tags.
var goTagKeys = []string{"tags", "tags_linux", "tags_windows", "tags_darwin"}

// validateGoBuildTags refuses a [projects.<name>.settings.go] table whose
// build tag keys (tags, tags_linux, tags_windows, tags_darwin) are not lists
// of tag names (letters, digits, "_" and "."), so a typo fails at load, not
// inside a build.
func validateGoBuildTags(projects map[string]ProjectOverride) error {
	names := make([]string, 0, len(projects))
	for name := range projects {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		table := projects[name].Settings["go"]
		for _, key := range goTagKeys {
			raw, ok := table[key]
			if !ok {
				continue
			}
			at := fmt.Sprintf("%s: [projects.%s.settings.go] %s", FileName, name, key)
			list, ok := raw.([]any)
			if !ok {
				return fmt.Errorf("%s must be a list of build tags, such as [\"production\", \"gtk3\"]", at)
			}
			for _, it := range list {
				if s, ok := it.(string); !ok || !goTagPattern.MatchString(s) {
					return fmt.Errorf("%s has %v, which is not a build tag: use letters, digits, \"_\" and \".\" in each tag", at, it)
				}
			}
		}
	}
	return nil
}
