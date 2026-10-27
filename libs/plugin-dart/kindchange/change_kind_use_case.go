package kindchange

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/fileedits"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/pathrefs"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

const manifestName = "pubspec.yaml"

var (
	smithBlock = regexp.MustCompile(`(?m)^smith:[ \t]*\r?\n(?:[ \t]+[^\n]*\n?)*`)
	kindKey    = regexp.MustCompile(`(?m)^([ \t]+kind:[ \t]*)(\S+)`)
	publishTo  = regexp.MustCompile(`(?m)^publish_to:[ \t]*(\S+)[ \t]*\r?\n`)
	versionKey = regexp.MustCompile(`(?m)^version:[^\n]*\n`)
	relative   = regexp.MustCompile(`(:[ \t]*['"]?)(\.{1,2}/[^\s'"#]+)`)
)

// Change answers what making the package params.Project a params.To changes:
// the smith kind in its pubspec.yaml, the publish_to guard (a package is
// published, a lib and an app are not), and the workspace members and path
// dependencies that name its directory.
func Change(workspaceRoot string, params protocol.ChangeKindParams) (protocol.ChangeKindResult, error) {
	p := params.Project
	edits := fileedits.New(workspaceRoot)
	manifest := path.Join(p.Root, manifestName)
	src, ok := edits.Read(manifest)
	if !ok {
		return protocol.ChangeKindResult{}, fmt.Errorf("%s not found", manifest)
	}
	var result protocol.ChangeKindResult

	src = setKind(src, string(params.To))
	if params.To == protocol.ProjectPackage {
		if m := publishTo.FindStringSubmatch(src); m != nil && m[1] == "none" {
			src = publishTo.ReplaceAllString(src, "")
		}
		if !regexp.MustCompile(`(?m)^(repository|homepage):`).MatchString(src) {
			result.Warnings = append(result.Warnings, fmt.Sprintf("pub.dev asks for a repository (or homepage) in %s before the first release; add one", manifest))
		}
	} else if !publishTo.MatchString(src) {
		src = guard(src)
	}
	moving := params.NewRoot != p.Root
	if moving {
		src = rebasePaths(src, p.Root, params.NewRoot)
	}
	edits.Write(manifest, src)
	result.Warnings = append(result.Warnings, programWarnings(workspaceRoot, p, params.To)...)

	if moving {
		moveOwnFiles(edits, params)
		retargetDependents(edits, params)
		if err := retargetWorkspace(edits, params); err != nil {
			return result, err
		}
		result.Warnings = append(result.Warnings, "run smith install afterwards: pub links workspace packages by directory, and .dart_tool still points at the old one")
	}
	result.Files = edits.Edits()
	return result, nil
}

// setKind records the kind in the pubspec's smith section.
func setKind(src, kind string) string {
	if block := smithBlock.FindStringIndex(src); block != nil {
		section := src[block[0]:block[1]]
		if kindKey.MatchString(section) {
			return src[:block[0]] + kindKey.ReplaceAllString(section, "${1}"+kind) + src[block[1]:]
		}
		trimmed := strings.TrimRight(section, "\r\n")
		return src[:block[0]] + trimmed + "\n  kind: " + kind + "\n" + src[block[1]:]
	}
	if src != "" && !strings.HasSuffix(src, "\n") {
		src += "\n"
	}
	return src + "\nsmith:\n  kind: " + kind + "\n"
}

// guard keeps a package that is not released from being published by accident:
// publish_to: none, after the version as the generated pubspec has it.
func guard(src string) string {
	if loc := versionKey.FindStringIndex(src); loc != nil {
		return src[:loc[1]] + "publish_to: none\n" + src[loc[1]:]
	}
	return "publish_to: none\n" + src
}

// programWarnings say what the kind change makes of the package's entry point.
func programWarnings(workspaceRoot string, p protocol.Project, to protocol.ProjectKind) []string {
	exists := func(rel string) bool {
		_, err := os.Stat(filepath.Join(workspaceRoot, filepath.FromSlash(p.Root), filepath.FromSlash(rel)))
		return err == nil
	}
	hasMain := exists("bin/main.dart") || exists("lib/main.dart")
	switch {
	case to == protocol.ProjectApp && !hasMain:
		return []string{"an app runs a program, and this package has no bin/main.dart or lib/main.dart: add one before running it"}
	case to != protocol.ProjectApp && hasMain:
		return []string{fmt.Sprintf("this package has a program (bin/main.dart or lib/main.dart); as a %s it is used by other packages and the program is kept but not deployed", to)}
	}
	return nil
}

// rebasePaths keeps relative paths in a file of the package pointing at the
// same places once its directory moves; dir is the file's directory.
func rebasePaths(src, oldRoot, newRoot string) string {
	return relative.ReplaceAllStringFunc(src, func(m string) string {
		parts := relative.FindStringSubmatch(m)
		return parts[1] + pathrefs.Rebase(parts[2], oldRoot, newRoot)
	})
}

// moveOwnFiles rebases the paths in the package's other YAML files (the
// analysis options that include the workspace's).
func moveOwnFiles(edits *fileedits.Set, params protocol.ChangeKindParams) {
	p := params.Project
	for _, file := range edits.List(p.Root, func(name string) bool { return name == "analysis_options.yaml" }) {
		text, _ := edits.Read(file)
		edits.Write(file, rebasePaths(text, p.Root, params.NewRoot))
	}
}

// retargetDependents points every other package's pubspec at the new
// directory.
func retargetDependents(edits *fileedits.Set, params protocol.ChangeKindParams) {
	for _, other := range params.Projects {
		if other.Language != "dart" || other.Root == params.Project.Root {
			continue
		}
		file := path.Join(other.Root, manifestName)
		if text, ok := edits.Read(file); ok {
			updated, _ := pathrefs.Retarget(text, other.Root, params.Project.Root, params.NewRoot)
			edits.Write(file, updated)
		}
	}
}

// retargetWorkspace updates the root pubspec's workspace members and the
// lockfile.
func retargetWorkspace(edits *fileedits.Set, params protocol.ChangeKindParams) error {
	for _, file := range []string{manifestName, "pubspec.lock"} {
		if text, ok := edits.Read(file); ok {
			updated, _ := pathrefs.Retarget(text, "", params.Project.Root, params.NewRoot)
			edits.Write(file, updated)
		}
	}
	return nil
}
