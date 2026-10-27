package kindchange

import (
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-go/discovery"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/fileedits"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/pathrefs"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

var (
	kindComment = regexp.MustCompile(`(?m)^//\s*smith:kind\s+\w+[ \t]*(\r?\n)?`)
	replacePath = regexp.MustCompile(`(=>[ \t]*)(\.{1,2}/[^\s]*)`)
)

// Change answers what making the module params.Project a params.To changes.
// Go has no publish guard: a package is the module that is tagged and
// fetched, so the kind is the comment smith reads, and the rest is the move.
func Change(workspaceRoot string, params protocol.ChangeKindParams) (protocol.ChangeKindResult, error) {
	p := params.Project
	edits := fileedits.New(workspaceRoot)
	manifest := path.Join(p.Root, "go.mod")
	src, ok := edits.Read(manifest)
	if !ok {
		return protocol.ChangeKindResult{}, fmt.Errorf("%s not found", manifest)
	}
	mod, err := discovery.ParseModule(src)
	if err != nil {
		return protocol.ChangeKindResult{}, fmt.Errorf("%s: %w", manifest, err)
	}
	var result protocol.ChangeKindResult

	comment := "// smith:kind " + string(params.To)
	if kindComment.MatchString(src) {
		src = kindComment.ReplaceAllString(src, comment+"${1}")
	} else {
		src = comment + "\n" + src
	}
	moving := params.NewRoot != p.Root
	newModule := mod.Path
	if moving {
		src = rebaseReplaces(src, p.Root, params.NewRoot)
		if prefix, ends := strings.CutSuffix(mod.Path, p.Root); ends && (prefix == "" || strings.HasSuffix(prefix, "/")) {
			newModule = prefix + params.NewRoot
		}
	}
	edits.Write(manifest, src)
	result.Warnings = append(result.Warnings, mainWarnings(workspaceRoot, p, params.To)...)

	if moving {
		retargetDependents(edits, params)
		renameModule(edits, params, mod.Path, newModule)
		if err := retargetWorkspace(edits, params); err != nil {
			return result, err
		}
		if p.Kind == protocol.ProjectPackage || params.To == protocol.ProjectPackage {
			result.Warnings = append(result.Warnings, fmt.Sprintf("a Go module is fetched at the tags of its directory (%s/vX.Y.Z), so the versions already tagged stay under %s/ and the next ones go under %s/", p.Root, p.Root, params.NewRoot))
		}
		if newModule != mod.Path {
			result.Warnings = append(result.Warnings, fmt.Sprintf("the module path changes from %s to %s: code outside this workspace that imports it must change too", mod.Path, newModule))
		}
	}
	result.Files = edits.Edits()
	return result, nil
}

// mainWarnings say what the kind change makes of the module's program.
func mainWarnings(workspaceRoot string, p protocol.Project, to protocol.ProjectKind) []string {
	hasMain := discovery.HoldsMainPackage(filepath.Join(workspaceRoot, filepath.FromSlash(p.Root)))
	switch {
	case to == protocol.ProjectApp && !hasMain:
		return []string{"an app runs a program, and the top of this module has no package main: add one before building it"}
	case to != protocol.ProjectApp && hasMain:
		return []string{fmt.Sprintf("the top of this module is a package main, which other modules cannot import; as a %s it is kept but not deployed", to)}
	}
	return nil
}

// rebaseReplaces keeps the module's own replace directives pointing at the
// same places once its directory moves.
func rebaseReplaces(src, oldRoot, newRoot string) string {
	return replacePath.ReplaceAllStringFunc(src, func(m string) string {
		parts := replacePath.FindStringSubmatch(m)
		return parts[1] + pathrefs.Rebase(parts[2], oldRoot, newRoot)
	})
}

// retargetDependents points the replace directives of every other module at
// the new directory.
func retargetDependents(edits *fileedits.Set, params protocol.ChangeKindParams) {
	for _, other := range params.Projects {
		if other.Language != "go" || other.Root == params.Project.Root {
			continue
		}
		file := path.Join(other.Root, "go.mod")
		if text, ok := edits.Read(file); ok {
			updated, _ := pathrefs.Retarget(text, other.Root, params.Project.Root, params.NewRoot)
			edits.Write(file, updated)
		}
	}
}

// renameModule changes the module path where it is declared, required,
// replaced and imported: in the go.mod and the .go files of every Go module.
func renameModule(edits *fileedits.Set, params protocol.ChangeKindParams, oldPath, newPath string) {
	if oldPath == newPath {
		return
	}
	roots := map[string]bool{params.Project.Root: true}
	for _, other := range params.Projects {
		if other.Language == "go" {
			roots[other.Root] = true
		}
	}
	for root := range roots {
		files := edits.Walk(root, func(name string) bool { return strings.HasSuffix(name, ".go") || name == "go.mod" })
		for _, file := range files {
			if text, ok := edits.Read(file); ok {
				updated, _ := pathrefs.ReplaceReference(text, oldPath, newPath)
				edits.Write(file, updated)
			}
		}
	}
}

// retargetWorkspace points go.work's use directive at the new directory.
func retargetWorkspace(edits *fileedits.Set, params protocol.ChangeKindParams) error {
	if text, ok := edits.Read("go.work"); ok {
		updated, _ := pathrefs.Retarget(text, "", params.Project.Root, params.NewRoot)
		edits.Write("go.work", updated)
	}
	return nil
}
