package kindchange

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-rust/discovery"
	"github.com/monosmith-app/monosmith/libs/plugin-rust/scaffolding"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/fileedits"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/pathrefs"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/tomltext"
)

const (
	manifestName  = "Cargo.toml"
	smithTable    = "package.metadata.smith"
	packageTable  = "package"
	pathValueExpr = `(path[ \t]*=[ \t]*")([^"]+)(")`
)

var pathValue = regexp.MustCompile(pathValueExpr)

// Change answers what making the crate params.Project a params.To changes:
// the kind in its manifest, the publish guard (a package is published, a lib
// and an app are not), and every path dependency and workspace member that
// names its directory.
func Change(workspaceRoot string, params protocol.ChangeKindParams) (protocol.ChangeKindResult, error) {
	p := params.Project
	edits := fileedits.New(workspaceRoot)
	manifest := path.Join(p.Root, manifestName)
	src, ok := edits.Read(manifest)
	if !ok {
		return protocol.ChangeKindResult{}, fmt.Errorf("%s not found", manifest)
	}
	var result protocol.ChangeKindResult

	src = tomltext.SetString(src, smithTable, "kind", string(params.To))
	if params.To == protocol.ProjectPackage {
		if raw, found := tomltext.Raw(src, packageTable, "publish"); found && raw == "false" {
			src, _ = tomltext.Remove(src, packageTable, "publish")
		}
		for _, missing := range missingRegistryFields(src) {
			result.Warnings = append(result.Warnings, fmt.Sprintf("crates.io asks for %s in %s before the first release; add it", missing, manifest))
		}
	} else {
		src = tomltext.SetRaw(src, packageTable, "publish", "false")
	}
	if params.NewRoot != p.Root {
		src = rebasePaths(src, p.Root, params.NewRoot)
	}
	edits.Write(manifest, src)
	result.Warnings = append(result.Warnings, binaryWarnings(workspaceRoot, src, p, params.To)...)

	if params.NewRoot != p.Root {
		retargetDependents(edits, params)
		if err := retargetWorkspace(edits, params); err != nil {
			return result, err
		}
	}
	result.Files = edits.Edits()
	return result, nil
}

// missingRegistryFields are the fields crates.io requires that the manifest
// does not set (inheriting one from the workspace counts as setting it).
func missingRegistryFields(src string) []string {
	var missing []string
	if !regexp.MustCompile(`(?m)^[ \t]*description(\.workspace)?[ \t]*=`).MatchString(src) {
		missing = append(missing, "a description")
	}
	if !regexp.MustCompile(`(?m)^[ \t]*license(-file|\.workspace)?[ \t]*=`).MatchString(src) {
		missing = append(missing, "a license")
	}
	return missing
}

// binaryWarnings say what the kind change makes of the crate's entry point.
func binaryWarnings(workspaceRoot, manifest string, p protocol.Project, to protocol.ProjectKind) []string {
	_, err := os.Stat(filepath.Join(workspaceRoot, filepath.FromSlash(p.Root), "src", "main.rs"))
	hasBinary := err == nil || strings.Contains(manifest, "[[bin]]")
	switch {
	case to == protocol.ProjectApp && !hasBinary:
		return []string{"an app runs a program, and this crate has no src/main.rs: add one before running it"}
	case to != protocol.ProjectApp && hasBinary:
		return []string{fmt.Sprintf("this crate has a binary (src/main.rs); as a %s it is built as a dependency for other crates, and the binary is kept but not deployed", to)}
	}
	return nil
}

// rebasePaths keeps the crate's own path dependencies pointing at the same
// places once its directory moves.
func rebasePaths(src, oldRoot, newRoot string) string {
	return pathValue.ReplaceAllStringFunc(src, func(m string) string {
		parts := pathValue.FindStringSubmatch(m)
		return parts[1] + pathrefs.Rebase(parts[2], oldRoot, newRoot) + parts[3]
	})
}

// retargetDependents points every other crate's path dependencies at the new
// directory.
func retargetDependents(edits *fileedits.Set, params protocol.ChangeKindParams) {
	for _, other := range params.Projects {
		if other.Language != "rust" || other.Root == params.Project.Root {
			continue
		}
		file := path.Join(other.Root, manifestName)
		if text, ok := edits.Read(file); ok {
			updated, _ := pathrefs.Retarget(text, other.Root, params.Project.Root, params.NewRoot)
			edits.Write(file, updated)
		}
	}
}

// retargetWorkspace updates the root manifest: explicit members and path
// dependencies naming the old directory, and a member entry when none of the
// members covers the new one.
func retargetWorkspace(edits *fileedits.Set, params protocol.ChangeKindParams) error {
	text, ok := edits.Read(manifestName)
	if !ok {
		return nil
	}
	text, _ = pathrefs.Retarget(text, "", params.Project.Root, params.NewRoot)
	crate, err := discovery.ParseCargo([]byte(text))
	if err != nil {
		return fmt.Errorf("%s: %w", manifestName, err)
	}
	if len(crate.Members) > 0 && !covered(crate.Members, params.NewRoot) {
		if text, err = scaffolding.AddMember(text, params.NewRoot); err != nil {
			return err
		}
	}
	edits.Write(manifestName, text)
	return nil
}

func covered(members []string, dir string) bool {
	for _, m := range members {
		if ok, _ := path.Match(m, dir); ok {
			return true
		}
	}
	return false
}
