package dependencies

import (
	"regexp"
	"strings"
)

var (
	tableHeader = regexp.MustCompile(`^\s*\[([^\[\]]+)\]\s*(#.*)?\r?$`)
	keyedLine   = regexp.MustCompile(`^(\s*)("[^"]+"|'[^']+'|[A-Za-z0-9_-]+)(\s*=\s*)(.*)$`)
	firstString = regexp.MustCompile(`"[^"]*"`)
	versionPair = regexp.MustCompile(`(\bversion\s*=\s*)"[^"]*"`)
	versionLine = regexp.MustCompile(`^(\s*version\s*=\s*)"[^"]*"`)
	packagePair = regexp.MustCompile(`\bpackage\s*=\s*"([^"]*)"`)
)

// SetSpec sets the version requirement of one dependency in a Cargo.toml
// source, in the table section ("dependencies", "workspace.dependencies"...),
// and says whether it found the dependency. It edits only that string:
// comments, spacing, features and every other key stay as written. The
// dependency may be a plain string, an inline table or a [section.name]
// table of its own; a `package` rename is matched by the crate's real name.
func SetSpec(src, section, name, spec string) (string, bool) {
	lines := strings.Split(src, "\n")
	table := ""
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if m := tableHeader.FindStringSubmatch(line); m != nil {
			table = strings.ReplaceAll(strings.TrimSpace(m[1]), " ", "")
			key, ok := strings.CutPrefix(table, section+".")
			if !ok || strings.Contains(key, ".") {
				continue
			}
			end := i + 1
			for end < len(lines) && !tableHeader.MatchString(lines[end]) {
				end++
			}
			if crateOf(unquote(key), lines[i+1:end]) != name {
				continue
			}
			for j := i + 1; j < end; j++ {
				if versionLine.MatchString(lines[j]) {
					lines[j] = versionLine.ReplaceAllString(lines[j], `${1}"`+spec+`"`)
					return strings.Join(lines, "\n"), true
				}
			}
			continue
		}
		if table != section {
			continue
		}
		m := keyedLine.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		key, rest := unquote(m[2]), m[4]
		switch {
		case strings.HasPrefix(rest, `"`):
			if key != name {
				continue
			}
			loc := firstString.FindStringIndex(rest)
			rest = rest[:loc[0]] + `"` + spec + `"` + rest[loc[1]:]
		case strings.HasPrefix(rest, "{"):
			if crateOf(key, []string{rest}) != name || !versionPair.MatchString(rest) {
				continue
			}
			rest = versionPair.ReplaceAllString(rest, `${1}"`+spec+`"`)
		default:
			continue
		}
		lines[i] = m[1] + m[2] + m[3] + rest
		return strings.Join(lines, "\n"), true
	}
	return src, false
}

// crateOf is the crate a dependency key names: the `package` rename in its
// lines, else the key.
func crateOf(key string, lines []string) string {
	for _, l := range lines {
		if m := packagePair.FindStringSubmatch(l); m != nil {
			return m[1]
		}
	}
	return key
}

func unquote(key string) string {
	return strings.Trim(key, `"'`)
}
