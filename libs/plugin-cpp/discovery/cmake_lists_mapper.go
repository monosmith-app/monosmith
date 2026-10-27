package discovery

import (
	"path"
	"regexp"
	"slices"
	"strings"
)

// Lists is what smith reads from a CMakeLists.txt: the plain project() call,
// the SMITH_KIND and SMITH_STACK settings, and the add_subdirectory calls.
type Lists struct {
	// Project is false for a CMakeLists.txt with no project() call: a
	// subdirectory of a project, not a project.
	Project       bool
	Name, Version string
	Kind, Stack   string
	// Native is set(SMITH_NATIVE ON): the project is built on each operating
	// system, so its release archives come from a native job per runner.
	Native bool
	// Executable is the first add_executable target, an app's to start.
	Executable string
	// Subdirectories are the add_subdirectory sources, as written.
	Subdirectories []string
}

var (
	projectCall = regexp.MustCompile(`(?is)\bproject\s*\(\s*([A-Za-z0-9_.+-]+)([^)]*)\)`)
	versionArg  = regexp.MustCompile(`(?i)\bVERSION\s+([0-9][0-9.]*)`)
	setting     = regexp.MustCompile(`(?i)\bset\s*\(\s*(SMITH_KIND|SMITH_STACK|SMITH_NATIVE)\s+"?([A-Za-z0-9_-]+)"?\s*\)`)
	executable  = regexp.MustCompile(`(?i)\badd_executable\s*\(\s*([A-Za-z0-9_.+-]+)`)
	subdir      = regexp.MustCompile(`(?i)\badd_subdirectory\s*\(\s*"?([^"\s)]+)"?`)
	comment     = regexp.MustCompile(`(?m)#[^\n]*`)
)

// ParseLists reads a CMakeLists.txt, ignoring comments.
func ParseLists(src string) Lists {
	src = comment.ReplaceAllString(src, "")
	var l Lists
	if m := projectCall.FindStringSubmatch(src); m != nil {
		l.Project, l.Name = true, m[1]
		if v := versionArg.FindStringSubmatch(m[2]); v != nil {
			l.Version = v[1]
		}
	}
	for _, m := range setting.FindAllStringSubmatch(src, -1) {
		switch strings.ToUpper(m[1]) {
		case "SMITH_KIND":
			l.Kind = m[2]
		case "SMITH_STACK":
			l.Stack = m[2]
		case "SMITH_NATIVE":
			l.Native = slices.Contains([]string{"ON", "TRUE", "YES", "1"}, strings.ToUpper(m[2]))
		}
	}
	if m := executable.FindStringSubmatch(src); m != nil {
		l.Executable = m[1]
	}
	for _, m := range subdir.FindAllStringSubmatch(src, -1) {
		l.Subdirectories = append(l.Subdirectories, m[1])
	}
	return l
}

// SubdirectoryRoot resolves an add_subdirectory source written in the
// project at root to a workspace-relative directory, or "" when it leaves
// the workspace or names a variable smith cannot resolve.
func SubdirectoryRoot(root, source string) string {
	for _, prefix := range []string{"${CMAKE_CURRENT_SOURCE_DIR}/", "${CMAKE_CURRENT_LIST_DIR}/", "${PROJECT_SOURCE_DIR}/"} {
		source = strings.TrimPrefix(source, prefix)
	}
	if strings.Contains(source, "${") || path.IsAbs(source) {
		return ""
	}
	resolved := path.Clean(path.Join(root, source))
	if resolved == "." || strings.HasPrefix(resolved, "../") || resolved == ".." {
		return ""
	}
	return resolved
}
