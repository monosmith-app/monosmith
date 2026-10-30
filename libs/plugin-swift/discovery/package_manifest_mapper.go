package discovery

import (
	"path"
	"regexp"
	"strings"
)

// Manifest is the file that makes a directory a Swift project.
const Manifest = "Package.swift"

// Package is what smith reads from a Package.swift. The manifest is Swift
// code; smith reads it with regular expressions and never runs it.
type Package struct {
	// Kind, Stack and Version are the smithKind, smithStack and smithVersion
	// comments.
	Kind, Stack, Version string
	// Paths are the local packages it depends on (.package(path: "../x")),
	// relative to it.
	Paths []string
	// Executable is true when it declares an executable product or target.
	Executable bool
}

var (
	comment    = regexp.MustCompile(`(?m)^//\s*smith(Kind|Stack|Version):\s*(\S+)`)
	pathDep    = regexp.MustCompile(`\.package\(\s*(?:name:\s*"[^"]*"\s*,\s*)?path:\s*"([^"]+)"`)
	executable = regexp.MustCompile(`\.(?:executableTarget|executable)\(`)
	toolsLine  = regexp.MustCompile(`\A//\s*swift-tools-version[^\r\n]*(\r?\n)`)
)

// ParsePackage reads a Package.swift.
func ParsePackage(src string) Package {
	var p Package
	for _, m := range comment.FindAllStringSubmatch(src, -1) {
		switch m[1] {
		case "Kind":
			p.Kind = m[2]
		case "Stack":
			p.Stack = m[2]
		case "Version":
			p.Version = m[2]
		}
	}
	for _, m := range pathDep.FindAllStringSubmatch(src, -1) {
		p.Paths = append(p.Paths, m[1])
	}
	p.Executable = executable.MatchString(src)
	return p
}

// PackageRoot is the workspace directory of a path dependency of the project
// in root, or empty when it leaves the workspace.
func PackageRoot(root, dep string) string {
	dir := path.Clean(path.Join(root, dep))
	if dir == ".." || strings.HasPrefix(dir, "../") || dir == "." {
		return ""
	}
	return dir
}

// SetField writes one header comment (field is Kind, Stack or Version),
// replacing it or adding it under the swift-tools-version line, which must
// stay first (as the first line of a manifest without one), in the file's own
// line ending. It says whether the manifest changed.
func SetField(src, field, value string) (string, bool) {
	line := "// smith" + field + ": " + value
	re := regexp.MustCompile(`(?m)^//\s*smith` + field + `:[^\r\n]*`)
	if re.MatchString(src) {
		out := re.ReplaceAllLiteralString(src, line)
		return out, out != src
	}
	eol := "\n"
	if strings.Contains(src, "\r\n") {
		eol = "\r\n"
	}
	if loc := toolsLine.FindStringIndex(src); loc != nil {
		return src[:loc[1]] + line + eol + src[loc[1]:], true
	}
	return line + eol + src, true
}
