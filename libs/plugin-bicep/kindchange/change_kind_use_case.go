package kindchange

import (
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-bicep/discovery"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/fileedits"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/pathrefs"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

var (
	kindStatement = regexp.MustCompile(`(?m)^(metadata\s+smithKind\s*=\s*')([^']*)(')`)
	versionLine   = regexp.MustCompile(`(?m)^metadata\s+version\s*=[^\n]*\n`)
	// relativeString is a quoted relative path in a Bicep file: a module's
	// source, a parameters file's using, loadTextContent.
	relativeString = regexp.MustCompile(`'(\.{1,2}/[^'\r\n]*)'`)
)

func isBicepFile(name string) bool {
	return strings.HasSuffix(name, ".bicep") || strings.HasSuffix(name, ".bicepparam")
}

// Change answers what making the project params.Project a params.To changes:
// the smithKind metadata of its main.bicep, and the module sources and
// parameter files that name its directory. A Bicep project is released by its
// tag whatever its kind, so nothing else depends on it.
func Change(workspaceRoot string, params protocol.ChangeKindParams) (protocol.ChangeKindResult, error) {
	p := params.Project
	edits := fileedits.New(workspaceRoot)
	manifest := path.Join(p.Root, discovery.Manifest)
	src, ok := edits.Read(manifest)
	if !ok {
		return protocol.ChangeKindResult{}, fmt.Errorf("%s not found", manifest)
	}
	var result protocol.ChangeKindResult

	edits.Write(manifest, setKind(src, string(params.To)))
	if _, hasParams := edits.Read(path.Join(p.Root, "main.bicepparam")); params.To != protocol.ProjectApp && hasParams {
		result.Warnings = append(result.Warnings, fmt.Sprintf("%s has a parameters file (main.bicepparam), which deploys it; as a %s it is a module other templates use, and the file is kept", p.Root, params.To))
	}
	if params.To == protocol.ProjectPackage {
		result.Warnings = append(result.Warnings, "a Bicep project is released by its tag: there is no registry to publish a package to yet")
	}

	if params.NewRoot != p.Root {
		moveOwnFiles(edits, params)
		retargetDependents(edits, params)
	}
	result.Files = edits.Edits()
	return result, nil
}

// setKind sets the smithKind metadata, adding it under the version when there
// is none.
func setKind(src, kind string) string {
	if kindStatement.MatchString(src) {
		return kindStatement.ReplaceAllString(src, "${1}"+kind+"${3}")
	}
	statement := "metadata smithKind = '" + kind + "'\n"
	if loc := versionLine.FindStringIndex(src); loc != nil {
		return src[:loc[1]] + statement + src[loc[1]:]
	}
	return statement + src
}

// moveOwnFiles keeps the project's own relative paths (the modules it uses)
// pointing at the right places once its directory moves.
func moveOwnFiles(edits *fileedits.Set, params protocol.ChangeKindParams) {
	p := params.Project
	for _, file := range edits.Walk(p.Root, isBicepFile) {
		text, _ := edits.Read(file)
		dir := path.Dir(file)
		edits.Write(file, relativeString.ReplaceAllStringFunc(text, func(m string) string {
			return `'` + pathrefs.RebaseFrom(relativeString.FindStringSubmatch(m)[1], dir, p.Root, params.NewRoot) + `'`
		}))
	}
}

// retargetDependents points every other project's module sources at the new
// directory.
func retargetDependents(edits *fileedits.Set, params protocol.ChangeKindParams) {
	p := params.Project
	for _, other := range params.Projects {
		if other.Language != "bicep" || other.Root == p.Root {
			continue
		}
		for _, file := range edits.Walk(other.Root, isBicepFile) {
			text, _ := edits.Read(file)
			dir := path.Dir(file)
			edits.Write(file, relativeString.ReplaceAllStringFunc(text, func(m string) string {
				return `'` + pathrefs.Follow(relativeString.FindStringSubmatch(m)[1], dir, p.Root, params.NewRoot) + `'`
			}))
		}
	}
}
