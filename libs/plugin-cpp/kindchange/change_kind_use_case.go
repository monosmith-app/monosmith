package kindchange

import (
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-cpp/discovery"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/fileedits"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/pathrefs"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

const listsName = "CMakeLists.txt"

var (
	kindSetting = regexp.MustCompile(`(?i)(\bset\s*\(\s*SMITH_KIND\s+"?)([A-Za-z0-9_-]+)("?\s*\))`)
	projectLine = regexp.MustCompile(`(?im)^[^\n#]*\bproject\s*\([^)]*\)[ \t]*\r?\n`)
	// relativeToken is a relative path in a CMake file, with the variable that
	// anchors it when it has one: ${CMAKE_CURRENT_SOURCE_DIR}/../core.
	relativeToken = regexp.MustCompile(`((?:\$\{(?:CMAKE_CURRENT_SOURCE_DIR|CMAKE_CURRENT_LIST_DIR|PROJECT_SOURCE_DIR)\}/)?)(\.{1,2}/[^\s"')]*)`)
	// subdirectoryCall is an add_subdirectory whose source is written without
	// a leading ./ (add_subdirectory(libs/core core)): the root file's way.
	subdirectoryCall = regexp.MustCompile(`(?i)(\badd_subdirectory\s*\(\s*"?(?:\$\{(?:CMAKE_CURRENT_SOURCE_DIR|CMAKE_CURRENT_LIST_DIR|PROJECT_SOURCE_DIR)\}/)?)([A-Za-z0-9_][^"\s)$]*)`)
)

func isCMakeFile(name string) bool {
	return name == listsName || strings.HasSuffix(name, ".cmake")
}

// Change answers what making the project params.Project a params.To changes:
// the SMITH_KIND setting in its CMakeLists.txt, and every add_subdirectory and
// relative path that names its directory.
func Change(workspaceRoot string, params protocol.ChangeKindParams) (protocol.ChangeKindResult, error) {
	p := params.Project
	edits := fileedits.New(workspaceRoot)
	manifest := path.Join(p.Root, listsName)
	src, ok := edits.Read(manifest)
	if !ok {
		return protocol.ChangeKindResult{}, fmt.Errorf("%s not found", manifest)
	}
	var result protocol.ChangeKindResult

	src = setKind(src, string(params.To))
	edits.Write(manifest, src)
	result.Warnings = append(result.Warnings, programWarnings(src, manifest, params.To)...)

	if params.NewRoot != p.Root {
		moveOwnFiles(edits, params)
		retargetOthers(edits, params)
	}
	result.Files = edits.Edits()
	return result, nil
}

// setKind sets the SMITH_KIND setting, adding it under the project() call when
// there is none.
func setKind(src, kind string) string {
	if kindSetting.MatchString(src) {
		return kindSetting.ReplaceAllString(src, "${1}"+kind+"${3}")
	}
	eol := "\n"
	if strings.Contains(src, "\r\n") {
		eol = "\r\n"
	}
	if loc := projectLine.FindStringIndex(src); loc != nil {
		return src[:loc[1]] + `set(SMITH_KIND "` + kind + `")` + eol + src[loc[1]:]
	}
	return src
}

// programWarnings say what the kind change makes of the project's program.
func programWarnings(src, manifest string, to protocol.ProjectKind) []string {
	lists := discovery.ParseLists(src)
	switch {
	case to == protocol.ProjectApp && lists.Executable == "":
		return []string{fmt.Sprintf("an app runs a program, and %s has no add_executable: add one before running it", manifest)}
	case to != protocol.ProjectApp && lists.Executable != "":
		return []string{fmt.Sprintf("this project builds a program (%s); as a %s it is added by other projects and the program is kept but not deployed", lists.Executable, to)}
	}
	return nil
}

// moveOwnFiles keeps the project's own relative paths (an add_subdirectory of
// another project, an include of shared CMake code) pointing at the right
// places once its directory moves.
func moveOwnFiles(edits *fileedits.Set, params protocol.ChangeKindParams) {
	p := params.Project
	for _, file := range edits.Walk(p.Root, isCMakeFile) {
		text, _ := edits.Read(file)
		dir := path.Dir(file)
		edits.Write(file, relativeToken.ReplaceAllStringFunc(text, func(m string) string {
			parts := relativeToken.FindStringSubmatch(m)
			return parts[1] + pathrefs.RebaseFrom(parts[2], dir, p.Root, params.NewRoot)
		}))
	}
}

// retargetOthers points every other project's CMake files, and the workspace
// root's, at the new directory.
func retargetOthers(edits *fileedits.Set, params protocol.ChangeKindParams) {
	p := params.Project
	follow := func(file string) {
		text, _ := edits.Read(file)
		dir := path.Dir(file)
		if dir == "." {
			dir = ""
		}
		text = subdirectoryCall.ReplaceAllStringFunc(text, func(m string) string {
			parts := subdirectoryCall.FindStringSubmatch(m)
			return parts[1] + pathrefs.Follow(parts[2], dir, p.Root, params.NewRoot)
		})
		text = relativeToken.ReplaceAllStringFunc(text, func(m string) string {
			parts := relativeToken.FindStringSubmatch(m)
			return parts[1] + pathrefs.Follow(parts[2], dir, p.Root, params.NewRoot)
		})
		edits.Write(file, text)
	}
	for _, file := range edits.List("", isCMakeFile) {
		follow(file)
	}
	for _, other := range params.Projects {
		if other.Language != "cpp" || other.Root == p.Root {
			continue
		}
		for _, file := range edits.Walk(other.Root, isCMakeFile) {
			follow(file)
		}
	}
}
