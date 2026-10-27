package kindchange

import (
	"encoding/json"
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/fileedits"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/pathrefs"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

const manifestName = "package.json"

// relativeString matches a JSON string that is a relative path, with the
// protocol npm gives a local dependency ("file:../core").
var relativeString = regexp.MustCompile(`"((?:file:|link:|portal:)?)(\.{1,2}/[^"]*)"`)

// workspaceFiles are the root files that name projects by directory: the
// manifest's workspaces, the tsconfig paths and references, the lockfile.
var workspaceFiles = []string{manifestName, "tsconfig.json", "tsconfig.base.json", "package-lock.json"}

// Change answers what making the project params.Project a params.To changes:
// the kind in package.json, private (a package is published, a lib and an app
// are not), and every reference to its directory in the workspace.
func Change(workspaceRoot string, params protocol.ChangeKindParams) (protocol.ChangeKindResult, error) {
	p := params.Project
	edits := fileedits.New(workspaceRoot)
	manifest := path.Join(p.Root, manifestName)
	src, ok := edits.Read(manifest)
	if !ok {
		return protocol.ChangeKindResult{}, fmt.Errorf("%s not found", manifest)
	}
	var result protocol.ChangeKindResult
	src, err := setKind(src, string(params.To))
	if err != nil {
		return result, fmt.Errorf("%s: %w", manifest, err)
	}
	if src, err = setPrivate(src, params.To != protocol.ProjectPackage); err != nil {
		return result, fmt.Errorf("%s: %w", manifest, err)
	}
	doc, err := jsondoc.ParseDocument([]byte(src))
	if err != nil {
		return result, fmt.Errorf("%s: %w", manifest, err)
	}
	if params.To == protocol.ProjectPackage {
		for _, field := range []string{"license", "description"} {
			if !doc.Has(field) {
				result.Warnings = append(result.Warnings, fmt.Sprintf("%s has no %q: the registry page of a published package shows it", manifest, field))
			}
		}
	}
	edits.Write(manifest, src)
	result.Warnings = append(result.Warnings, entryWarnings(doc, manifest, params.To)...)

	if params.NewRoot != p.Root {
		moveOwnFiles(edits, params)
		retargetDependents(edits, params)
		if err := retargetWorkspace(edits, params); err != nil {
			return result, err
		}
		result.Warnings = append(result.Warnings, "run smith install afterwards: npm links workspace packages by directory, and node_modules still points at the old one")
	}
	result.Files = edits.Edits()
	return result, nil
}

// entryWarnings say what the kind change makes of the project's entry points.
func entryWarnings(doc *jsondoc.Document, manifest string, to protocol.ProjectKind) []string {
	var raw map[string]json.RawMessage
	if json.Unmarshal(doc.Bytes(), &raw) != nil {
		return nil
	}
	_, hasStart := scriptOf(raw, "start")
	_, hasBin := raw["bin"]
	switch {
	case to == protocol.ProjectApp && !hasStart && !hasBin:
		return []string{manifest + " has no start script or bin: an app runs something, add one before deploying it"}
	case to != protocol.ProjectApp && (hasStart || hasBin):
		return []string{fmt.Sprintf("%s has a start script or a bin; as a %s it is installed by other projects and the program is kept but not deployed", manifest, to)}
	}
	return nil
}

func scriptOf(raw map[string]json.RawMessage, name string) (string, bool) {
	var scripts map[string]string
	if json.Unmarshal(raw["scripts"], &scripts) != nil {
		return "", false
	}
	s, ok := scripts[name]
	return s, ok
}

// configFiles lists the manifest and tsconfig files directly in a project.
func configFiles(edits *fileedits.Set, root string) []string {
	return edits.List(root, func(name string) bool {
		return name == manifestName || (strings.HasPrefix(name, "tsconfig") && strings.HasSuffix(name, ".json"))
	})
}

// moveOwnFiles keeps the project's own relative references (to the shared
// tsconfig, to sibling projects) and its build folder pointing at the right
// places once its directory moves.
func moveOwnFiles(edits *fileedits.Set, params protocol.ChangeKindParams) {
	p := params.Project
	for _, file := range configFiles(edits, p.Root) {
		text, _ := edits.Read(file)
		text = relativeString.ReplaceAllStringFunc(text, func(m string) string {
			parts := relativeString.FindStringSubmatch(m)
			return `"` + parts[1] + pathrefs.Rebase(parts[2], p.Root, params.NewRoot) + `"`
		})
		// A build goes to dist/<project dir>: the folder follows the directory.
		text, _ = pathrefs.Retarget(text, params.NewRoot, protocol.ProjectBuildDir(p.Root), protocol.ProjectBuildDir(params.NewRoot))
		edits.Write(file, text)
	}
}

// retargetDependents points every other project's manifest and tsconfig
// references at the new directory.
func retargetDependents(edits *fileedits.Set, params protocol.ChangeKindParams) {
	for _, other := range params.Projects {
		if other.Language != "node" || other.Root == params.Project.Root {
			continue
		}
		for _, file := range configFiles(edits, other.Root) {
			text, _ := edits.Read(file)
			updated, _ := pathrefs.Retarget(text, other.Root, params.Project.Root, params.NewRoot)
			edits.Write(file, updated)
		}
	}
}

// retargetWorkspace updates the root files, and lists the new directory in
// the workspaces when none of the patterns covers it.
func retargetWorkspace(edits *fileedits.Set, params protocol.ChangeKindParams) error {
	for _, file := range workspaceFiles {
		text, ok := edits.Read(file)
		if !ok {
			continue
		}
		text, _ = pathrefs.Retarget(text, "", params.Project.Root, params.NewRoot)
		if file == manifestName {
			var err error
			if text, err = ensureWorkspace(text, params.NewRoot); err != nil {
				return err
			}
		}
		edits.Write(file, text)
	}
	return nil
}

// ensureWorkspace adds dir's parent as a workspaces pattern when the root
// manifest lists workspaces and none of them matches dir.
func ensureWorkspace(src, dir string) (string, error) {
	var manifest struct {
		Workspaces json.RawMessage `json:"workspaces"`
	}
	if err := json.Unmarshal([]byte(src), &manifest); err != nil || len(manifest.Workspaces) == 0 {
		return src, err
	}
	var patterns []string
	object := false
	if json.Unmarshal(manifest.Workspaces, &patterns) != nil {
		var wrapped struct {
			Packages []string `json:"packages"`
		}
		if err := json.Unmarshal(manifest.Workspaces, &wrapped); err != nil {
			return src, nil
		}
		patterns, object = wrapped.Packages, true
	}
	if covered(patterns, dir) {
		return src, nil
	}
	patterns = append(patterns, path.Dir(dir)+"/*")
	doc, err := jsondoc.ParseDocument([]byte(src))
	if err != nil {
		return src, err
	}
	if object {
		err = doc.Object("workspaces", true).SetJSON("packages", patterns)
	} else {
		err = doc.SetJSON("workspaces", patterns)
	}
	if err != nil {
		return src, err
	}
	return string(doc.Bytes()), nil
}

func covered(patterns []string, dir string) bool {
	included := false
	for _, pattern := range patterns {
		negated := strings.HasPrefix(pattern, "!")
		pattern = strings.TrimPrefix(strings.TrimPrefix(pattern, "!"), "./")
		if matches(pattern, dir) {
			included = !negated
		}
	}
	return included
}

func matches(pattern, dir string) bool {
	if prefix, ok := strings.CutSuffix(pattern, "/**"); ok {
		return strings.HasPrefix(dir, prefix+"/")
	}
	ok, _ := path.Match(pattern, dir)
	return ok
}
