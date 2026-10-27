package kindchange

import (
	"fmt"
	"path"
	"regexp"

	"github.com/monosmith-app/monosmith/libs/plugin-jvm/discovery"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/fileedits"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/pathrefs"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// moduleFiles are the files of a module, or of the workspace root, that name
// other modules by directory or by Gradle project path.
var moduleFiles = []string{"pom.xml", "settings.gradle.kts", "settings.gradle", "build.gradle.kts", "build.gradle", "gradle.properties"}

// relativeString matches a quoted relative path in a build script.
var relativeString = regexp.MustCompile(`(["'])(\.{1,2}/[^"']*)(["'])`)

// Change answers what making the module params.Project a params.To changes.
// A Maven module (pom.xml) and a Gradle module (build.gradle[.kts] with a
// gradle.properties beside it) say their kind differently; the move is the
// same for both: the root pom's modules, the settings' includes, project
// dependencies and relative paths all name the directory.
func Change(workspaceRoot string, params protocol.ChangeKindParams) (protocol.ChangeKindResult, error) {
	p := params.Project
	edits := fileedits.New(workspaceRoot)
	var result protocol.ChangeKindResult
	var err error
	if _, maven := edits.Read(path.Join(p.Root, "pom.xml")); maven {
		err = changeMaven(edits, params, &result)
	} else {
		err = changeGradle(edits, params, &result)
	}
	if err != nil {
		return result, err
	}
	if params.NewRoot != p.Root {
		moveOwnFiles(edits, params)
		retargetOthers(edits, params)
		retargetRoot(edits, params)
	}
	result.Files = edits.Edits()
	return result, nil
}

// moveOwnFiles keeps the module's own relative paths (its parent pom, shared
// scripts) pointing at the right places once its directory moves.
func moveOwnFiles(edits *fileedits.Set, params protocol.ChangeKindParams) {
	p := params.Project
	for _, name := range []string{"pom.xml", "build.gradle.kts", "build.gradle"} {
		file := path.Join(p.Root, name)
		text, ok := edits.Read(file)
		if !ok {
			continue
		}
		if name == "pom.xml" {
			text = relativePath.ReplaceAllStringFunc(text, func(m string) string {
				parts := relativePath.FindStringSubmatch(m)
				return parts[1] + pathrefs.Rebase(parts[2], p.Root, params.NewRoot) + parts[3]
			})
		} else {
			text = relativeString.ReplaceAllStringFunc(text, func(m string) string {
				parts := relativeString.FindStringSubmatch(m)
				return parts[1] + pathrefs.Rebase(parts[2], p.Root, params.NewRoot) + parts[3]
			})
		}
		edits.Write(file, text)
	}
}

// retargetOthers points every other module's files at the new directory.
func retargetOthers(edits *fileedits.Set, params protocol.ChangeKindParams) {
	for _, other := range params.Projects {
		if other.Language != "jvm" || other.Root == params.Project.Root {
			continue
		}
		retargetIn(edits, other.Root, params)
	}
}

// retargetRoot updates the root pom's modules and the settings' includes.
func retargetRoot(edits *fileedits.Set, params protocol.ChangeKindParams) {
	retargetIn(edits, "", params)
}

// retargetIn rewrites, in the module files of dir, the relative paths and the
// Gradle project path that name the moved module.
func retargetIn(edits *fileedits.Set, dir string, params protocol.ChangeKindParams) {
	p := params.Project
	oldPath, newPath := discovery.ProjectPath(p.Root), discovery.ProjectPath(params.NewRoot)
	for _, name := range moduleFiles {
		file := path.Join(dir, name)
		text, ok := edits.Read(file)
		if !ok {
			continue
		}
		text, _ = pathrefs.Retarget(text, dir, p.Root, params.NewRoot)
		text, _ = pathrefs.ReplaceReference(text, oldPath, newPath)
		edits.Write(file, text)
	}
}

func warnf(result *protocol.ChangeKindResult, format string, args ...any) {
	result.Warnings = append(result.Warnings, fmt.Sprintf(format, args...))
}
