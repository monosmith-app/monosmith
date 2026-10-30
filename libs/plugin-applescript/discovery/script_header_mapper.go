package discovery

import (
	"regexp"
	"sort"
	"strings"
)

// ManifestPattern matches the plain-text source that makes a directory an
// AppleScript project. Compiled scripts (.scpt) and bundles (.scptd) are
// build outputs and are not read.
const ManifestPattern = "*.applescript"

// MainName is the main script of a project that has one by that name.
const MainName = "main.applescript"

// Header is what smith reads from the comments of a main script.
type Header struct {
	Kind, Stack, Version string
}

var header = regexp.MustCompile(`(?m)^--\s*smith(Kind|Stack|Version):\s*(\S+)`)

// ParseHeader reads the smithKind, smithStack and smithVersion comments.
func ParseHeader(src string) Header {
	var h Header
	for _, m := range header.FindAllStringSubmatch(src, -1) {
		switch m[1] {
		case "Kind":
			h.Kind = m[2]
		case "Stack":
			h.Stack = m[2]
		case "Version":
			h.Version = m[2]
		}
	}
	return h
}

// SetField writes one header comment (field is Kind, Stack or Version),
// replacing it or adding it as the first line, and says whether the script
// changed.
func SetField(src, field, value string) (string, bool) {
	line := "-- smith" + field + ": " + value
	re := regexp.MustCompile(`(?m)^--\s*smith` + field + `:[^\r\n]*`)
	out := line + "\n" + src
	if re.MatchString(src) {
		out = re.ReplaceAllLiteralString(src, line)
	}
	return out, out != src
}

// PickMain is the main script of a directory from the base names of its
// .applescript files: main.applescript, else the first by name.
func PickMain(names []string) string {
	var scripts []string
	for _, n := range names {
		if strings.HasSuffix(n, ".applescript") {
			scripts = append(scripts, n)
		}
	}
	sort.Strings(scripts)
	for _, n := range scripts {
		if n == MainName {
			return n
		}
	}
	if len(scripts) > 0 {
		return scripts[0]
	}
	return ""
}
