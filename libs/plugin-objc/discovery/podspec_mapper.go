package discovery

import (
	"regexp"
	"strings"
)

// ManifestPattern matches the file that makes a directory an Objective-C
// project.
const ManifestPattern = "*.podspec"

// Podspec is what smith reads from a podspec. A podspec is Ruby; smith reads
// its plain assignments and never runs it.
type Podspec struct {
	// Name and Version are the pod's. Version is empty when it is computed
	// (s.version = File.read(...)).
	Name, Version string
	// Kind and Stack are the smithKind and smithStack comments.
	Kind, Stack string
	// Dependencies are the pods it depends on (s.dependency 'Name'), a
	// subspec reference reduced to its pod.
	Dependencies []string
}

var (
	nameField    = regexp.MustCompile(`(?m)^\s*\w+\.name\s*=\s*["']([^"']+)["']`)
	versionField = regexp.MustCompile(`(?m)^(\s*\w+\.version\s*=\s*)(["'])([^"'\r\n]*)(["'])`)
	dependency   = regexp.MustCompile(`(?m)^\s*\w+\.(?:dependency|add_dependency)\s+["']([^"']+)["']`)
	smithComment = regexp.MustCompile(`(?m)^#\s*smith(Kind|Stack):\s*(\S+)`)
)

// ParsePodspec reads a podspec.
func ParsePodspec(src string) Podspec {
	var p Podspec
	if m := nameField.FindStringSubmatch(src); m != nil {
		p.Name = m[1]
	}
	if m := versionField.FindStringSubmatch(src); m != nil {
		p.Version = m[3]
	}
	for _, m := range smithComment.FindAllStringSubmatch(src, -1) {
		if m[1] == "Kind" {
			p.Kind = m[2]
		} else {
			p.Stack = m[2]
		}
	}
	for _, m := range dependency.FindAllStringSubmatch(src, -1) {
		p.Dependencies = append(p.Dependencies, strings.SplitN(m[1], "/", 2)[0])
	}
	return p
}

// SetKind writes the smithKind comment, replacing it or adding it as the
// first line, and says whether the podspec changed.
func SetKind(src, kind string) (string, bool) {
	line := "# smithKind: " + kind
	re := regexp.MustCompile(`(?m)^#\s*smithKind:[^\r\n]*`)
	if re.MatchString(src) {
		out := re.ReplaceAllLiteralString(src, line)
		return out, out != src
	}
	eol := "\n"
	if strings.Contains(src, "\r\n") {
		eol = "\r\n"
	}
	return line + eol + src, true
}

// SetVersion writes version into the literal version assignment, keeping its
// quotes, and says whether the podspec changed. A podspec whose version is
// computed is left alone.
func SetVersion(src, version string) (string, bool) {
	out := versionField.ReplaceAllString(src, "${1}${2}"+version+"${4}")
	return out, out != src
}
