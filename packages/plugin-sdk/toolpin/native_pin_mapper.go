package toolpin

import (
	"encoding/json"
	"regexp"
	"strings"
)

var (
	goDirective   = regexp.MustCompile(`(?m)^go\s+(\d+\.\d+)`)
	rustChannel   = regexp.MustCompile(`(?m)^[ \t]*channel[ \t]*=[ \t]*"([^"]+)"`)
	rustVersion   = regexp.MustCompile(`(?m)^[ \t]*rust-version[ \t]*=[ \t]*"([^"]+)"`)
	mavenRelease  = regexp.MustCompile(`<maven\.compiler\.release>\s*(\d+)\s*</maven\.compiler\.release>`)
	versionNumber = regexp.MustCompile(`^\d+(\.\d+)*$`)
)

// nativeFile is one file a language keeps its pin in, and how to read the
// version out of its text. A parser returns "" when the file names none.
type nativeFile struct {
	name  string
	parse func(src string) string
}

// nativeFiles lists, by toolchain label, the files read in order.
var nativeFiles = map[string][]nativeFile{
	"go":     {{"go.work", firstGroup(goDirective)}},
	"rustc":  {{"rust-toolchain.toml", firstGroup(rustChannel)}, {"Cargo.toml", firstGroup(rustVersion)}},
	"dotnet": {{"global.json", globalJSONSDK}},
	"python": {{".python-version", plainVersion("")}},
	"node":   {{".nvmrc", plainVersion("v")}, {".node-version", plainVersion("v")}},
	"ruby":   {{".ruby-version", plainVersion("ruby-")}},
	"php":    {{".php-version", plainVersion("")}},
	"java":   {{".java-version", javaMajor}, {"pom.xml", firstGroup(mavenRelease)}},
}

// firstGroup parses a file with the first submatch of a pattern.
func firstGroup(re *regexp.Regexp) func(string) string {
	return func(src string) string {
		if m := re.FindStringSubmatch(src); m != nil {
			return m[1]
		}
		return ""
	}
}

// plainVersion parses a file that holds one version on its first line, after
// an optional prefix. A line that is not a number ("lts/*", "system") names none.
func plainVersion(prefix string) func(string) string {
	return func(src string) string {
		first, _, _ := strings.Cut(strings.TrimSpace(src), "\n")
		v := strings.TrimPrefix(strings.TrimSpace(first), prefix)
		if !versionNumber.MatchString(v) {
			return ""
		}
		return v
	}
}

// javaMajor parses .java-version: the major of "21.0.2".
func javaMajor(src string) string {
	major, _, _ := strings.Cut(plainVersion("")(src), ".")
	return major
}

// globalJSONSDK parses global.json: the major.minor channel of sdk.version.
func globalJSONSDK(src string) string {
	var g struct {
		SDK struct{ Version string } `json:"sdk"`
	}
	if json.Unmarshal([]byte(src), &g) != nil || g.SDK.Version == "" {
		return ""
	}
	parts := strings.SplitN(g.SDK.Version, ".", 3)
	if len(parts) < 2 {
		return g.SDK.Version
	}
	return parts[0] + "." + parts[1]
}
