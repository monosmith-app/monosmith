package dependencies

import (
	"regexp"
	"strings"
)

var (
	topLevelKey = regexp.MustCompile(`^([A-Za-z_][\w-]*):`)
	versionLine = regexp.MustCompile(`^(\s+version:\s*)(["']?)([^"'#\s]+)(["']?)(\s*#.*)?(\r?)$`)
)

// SetSpec sets the version constraint of one dependency in a pubspec.yaml
// source, in the section ("dependencies" or "dev_dependencies"), and says
// whether it found the dependency. The constraint is the value after the
// name (`http: ^1.0.0`) or the `version:` line of a longer entry; only that
// text changes, so quotes, comments and every other key stay as written.
func SetSpec(src, section, name, spec string) (string, bool) {
	lines := strings.Split(src, "\n")
	in := false
	entry := regexp.MustCompile(`^(\s+` + regexp.QuoteMeta(name) + `:\s*)(["']?)([^"'#\s]+)(["']?)(\s*#.*)?(\r?)$`)
	bare := regexp.MustCompile(`^(\s+)` + regexp.QuoteMeta(name) + `:\s*(#.*)?\r?$`)
	for i, line := range lines {
		if m := topLevelKey.FindStringSubmatch(line); m != nil {
			in = m[1] == section
			continue
		}
		if !in {
			continue
		}
		if entry.MatchString(line) {
			lines[i] = entry.ReplaceAllString(line, "${1}${2}"+spec+"${4}${5}${6}")
			return strings.Join(lines, "\n"), true
		}
		if m := bare.FindStringSubmatch(line); m != nil {
			for j := i + 1; j < len(lines); j++ {
				next := lines[j]
				if strings.TrimSpace(next) != "" && indent(next) <= len(m[1]) {
					break
				}
				if versionLine.MatchString(next) {
					lines[j] = versionLine.ReplaceAllString(next, "${1}${2}"+spec+"${4}${5}${6}")
					return strings.Join(lines, "\n"), true
				}
			}
		}
	}
	return src, false
}

func indent(line string) int { return len(line) - len(strings.TrimLeft(line, " \t")) }
