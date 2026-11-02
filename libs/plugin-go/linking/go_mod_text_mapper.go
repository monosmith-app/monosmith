package linking

import (
	"regexp"
	"strings"
)

// placeholderVersion is the version written for a module of the workspace:
// go.work (or a replace directive) points Go at the folder, so the version is
// never fetched and any valid one will do.
const placeholderVersion = "v0.0.0"

func lineEnding(src string) string {
	if strings.Contains(src, "\r\n") {
		return "\r\n"
	}
	return "\n"
}

func requirePattern(path string) *regexp.Regexp {
	return regexp.MustCompile(`(?m)^[ \t]*(?:require[ \t]+)?` + regexp.QuoteMeta(path) + `[ \t]+v\S+[^\r\n]*\r?\n?`)
}

// HasRequire reports whether a go.mod text requires the module.
func HasRequire(gomod, path string) bool {
	return requirePattern(path).MatchString(gomod)
}

// AddRequire writes `require <path> <version>` into a go.mod text: inside the
// first require block when the file has one, else as a line of its own. The rest
// of the file is kept. It returns false when the file already requires the
// module.
func AddRequire(gomod, path, version string) (string, bool) {
	if HasRequire(gomod, path) {
		return gomod, false
	}
	if version == "" {
		version = placeholderVersion
	}
	nl := lineEnding(gomod)
	if open := regexp.MustCompile(`(?m)^require[ \t]*\([ \t]*\r?\n`).FindStringIndex(gomod); open != nil {
		if end := strings.Index(gomod[open[1]:], ")"); end >= 0 {
			at := open[1] + end
			return gomod[:at] + "\t" + path + " " + version + nl + gomod[at:], true
		}
	}
	if gomod != "" && !strings.HasSuffix(gomod, "\n") {
		gomod += nl
	}
	return gomod + nl + "require " + path + " " + version + nl, true
}

// RemoveRequire takes the module's require line out of a go.mod text, and its
// replace directive too. It returns false when the file required nothing of
// that name.
func RemoveRequire(gomod, path string) (string, bool) {
	out := requirePattern(path).ReplaceAllString(gomod, "")
	replace := regexp.MustCompile(`(?m)^replace[ \t]+` + regexp.QuoteMeta(path) + `[ \t]*(?:v\S+[ \t]*)?=>[^\r\n]*\r?\n?`)
	out = replace.ReplaceAllString(out, "")
	return out, out != gomod
}

// AddReplace points a module at a folder with `replace <path> => <folder>`,
// for a workspace with no go.work. It returns false when the file already
// replaces the module.
func AddReplace(gomod, path, folder string) (string, bool) {
	if regexp.MustCompile(`(?m)^replace[ \t]+` + regexp.QuoteMeta(path) + `[ \t]`).MatchString(gomod) {
		return gomod, false
	}
	nl := lineEnding(gomod)
	if gomod != "" && !strings.HasSuffix(gomod, "\n") {
		gomod += nl
	}
	return gomod + nl + "replace " + path + " => " + folder + nl, true
}

// WorkUses reports whether a go.work text lists the folder (workspace-relative,
// forward slashes) in a use directive.
func WorkUses(gowork, folder string) bool {
	want := strings.TrimPrefix(strings.Trim(folder, "/"), "./")
	inBlock := false
	for _, raw := range strings.Split(gowork, "\n") {
		line := strings.TrimSpace(raw)
		if i := strings.Index(line, "//"); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		switch {
		case line == "use (":
			inBlock = true
			continue
		case inBlock && line == ")":
			inBlock = false
			continue
		case !inBlock && strings.HasPrefix(line, "use "):
			line = strings.TrimSpace(strings.TrimPrefix(line, "use "))
		case !inBlock:
			continue
		}
		if strings.TrimPrefix(strings.Trim(line, `"`), "./") == want {
			return true
		}
	}
	return false
}
