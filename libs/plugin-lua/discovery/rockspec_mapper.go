package discovery

import (
	"regexp"
	"strings"
)

// ManifestPattern matches the file that makes a directory a Lua project.
const ManifestPattern = "*.rockspec"

// Rockspec is what smith reads from a rockspec. A rockspec is Lua code; smith
// reads its plain top-level assignments and never runs it.
type Rockspec struct {
	// Package is the rock's name; Version its LuaRocks version, with the
	// revision (1.2.0-1).
	Package, Version string
	// Kind and Stack are the smithKind and smithStack comments.
	Kind, Stack string
	// Dependencies are the names of the rocks it depends on, lua excluded.
	Dependencies []string
}

var (
	packageField = regexp.MustCompile(`(?m)^package\s*=\s*["']([^"']+)["']`)
	versionField = regexp.MustCompile(`(?m)^version\s*=\s*["']([^"']+)["']`)
	smithComment = regexp.MustCompile(`(?m)^--\s*smith(Kind|Stack):\s*(\S+)`)
	dependencies = regexp.MustCompile(`(?ms)^dependencies\s*=\s*\{(.*?)\}`)
	quoted       = regexp.MustCompile(`["']([^"']+)["']`)
)

// ParseRockspec reads a rockspec.
func ParseRockspec(src string) Rockspec {
	var r Rockspec
	if m := packageField.FindStringSubmatch(src); m != nil {
		r.Package = m[1]
	}
	if m := versionField.FindStringSubmatch(src); m != nil {
		r.Version = m[1]
	}
	for _, m := range smithComment.FindAllStringSubmatch(src, -1) {
		if m[1] == "Kind" {
			r.Kind = m[2]
		} else {
			r.Stack = m[2]
		}
	}
	if block := dependencies.FindStringSubmatch(src); block != nil {
		for _, m := range quoted.FindAllStringSubmatch(block[1], -1) {
			name := strings.FieldsFunc(m[1], func(c rune) bool { return strings.ContainsRune(" <>=~", c) })
			if len(name) > 0 && name[0] != "lua" {
				r.Dependencies = append(r.Dependencies, name[0])
			}
		}
	}
	return r
}

var revision = regexp.MustCompile(`-\d+$`)

// ReleaseVersion is the version without LuaRocks' revision: 1.2.0-1 is 1.2.0.
func ReleaseVersion(luarocksVersion string) string {
	return revision.ReplaceAllString(luarocksVersion, "")
}
