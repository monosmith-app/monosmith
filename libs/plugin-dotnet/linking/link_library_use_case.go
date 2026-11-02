package linking

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-dotnet/dependencies"
	"github.com/monosmith-app/monosmith/libs/plugin-dotnet/discovery"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// ReferenceInclude is the path of the library's project file as seen from the
// project's file, with forward slashes (../../libs/Core/Core.csproj). MSBuild
// accepts them on every system; backslashes would be wrong in a Linux checkout.
func ReferenceInclude(projectFile, libraryFile string) string {
	rel, err := filepath.Rel(filepath.FromSlash(path.Dir(projectFile)), filepath.FromSlash(libraryFile))
	if err != nil {
		rel = libraryFile
	}
	return filepath.ToSlash(rel)
}

// slashReference rewrites the backslash form of the new reference, which
// dotnet add reference writes whatever the system, to forward slashes.
//
// Parameters: workspaceRoot is the workspace root; projectFile and libraryFile
// are workspace-relative project files.
// Returns: the error from reading or writing the project file.
func slashReference(workspaceRoot, projectFile, libraryFile string) error {
	full := filepath.Join(workspaceRoot, filepath.FromSlash(projectFile))
	raw, err := os.ReadFile(full)
	if err != nil {
		return err
	}
	slashed := ReferenceInclude(projectFile, libraryFile)
	fixed := strings.ReplaceAll(string(raw), `Include="`+strings.ReplaceAll(slashed, "/", `\`)+`"`, `Include="`+slashed+`"`)
	if fixed == string(raw) {
		return nil
	}
	return os.WriteFile(full, []byte(fixed), 0o644)
}

// AddArgs is the dotnet command that adds a ProjectReference to the library
// and restores the project.
func AddArgs(projectFile, libraryFile string) []string {
	return []string{"add", projectFile, "reference", libraryFile}
}

// RemoveArgs is the dotnet command that removes the ProjectReference.
func RemoveArgs(projectFile, libraryFile string) []string {
	return []string{"remove", projectFile, "reference", libraryFile}
}

// files finds the project file of both projects, or fails naming the folder
// that has none.
func files(workspaceRoot string, project, library protocol.Project) (string, string, error) {
	pf, err := dependencies.ProjectFileOf(workspaceRoot, project)
	if err != nil {
		return "", "", err
	}
	lf, err := dependencies.ProjectFileOf(workspaceRoot, library)
	if err != nil {
		return "", "", err
	}
	return pf, lf, nil
}

// references reports whether the project file already has a ProjectReference
// to the library's file.
func references(workspaceRoot, projectFile, libraryFile string) (bool, error) {
	raw, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(projectFile)))
	if err != nil {
		return false, fmt.Errorf("%s: %w", projectFile, err)
	}
	parsed, err := discovery.ParseProject(projectFile, raw)
	if err != nil {
		return false, fmt.Errorf("%s: %w", projectFile, err)
	}
	want := path.Clean(path.Join(path.Dir(projectFile), strings.ReplaceAll(ReferenceInclude(projectFile, libraryFile), `\`, "/")))
	return slices.ContainsFunc(parsed.References, func(ref string) bool {
		return path.Clean(path.Join(path.Dir(projectFile), strings.ReplaceAll(ref, `\`, "/"))) == want
	}), nil
}

// Link adds a ProjectReference from the project to the library with dotnet add
// reference, which also restores the project. With params.Plan nothing is
// written.
func Link(workspaceRoot string, params protocol.LinkParams) (protocol.LinkResult, error) {
	pf, lf, err := files(workspaceRoot, params.Project, params.Library)
	if err != nil {
		return protocol.LinkResult{}, err
	}
	res := protocol.LinkResult{Manifest: pf, Line: fmt.Sprintf(`<ProjectReference Include=%q />`, ReferenceInclude(pf, lf))}
	if params.Plan {
		return res, nil
	}
	already, err := references(workspaceRoot, pf, lf)
	if err != nil {
		return protocol.LinkResult{}, err
	}
	if already {
		res.Notes = []string{params.Project.Name + " already references " + params.Library.Name}
		return res, nil
	}
	if err := dependencies.Dotnet(workspaceRoot, AddArgs(pf, lf)); err != nil {
		return protocol.LinkResult{}, err
	}
	if err := slashReference(workspaceRoot, pf, lf); err != nil {
		return protocol.LinkResult{}, err
	}
	res.Files = []string{pf}
	return res, nil
}

// Unlink removes the ProjectReference to the library. When the project file had
// none, the result has no files.
func Unlink(workspaceRoot string, params protocol.UnlinkParams) (protocol.DependencyResult, error) {
	pf, lf, err := files(workspaceRoot, params.Project, params.Library)
	if err != nil {
		return protocol.DependencyResult{}, err
	}
	has, err := references(workspaceRoot, pf, lf)
	if err != nil || !has {
		return protocol.DependencyResult{}, err
	}
	if err := dependencies.Dotnet(workspaceRoot, RemoveArgs(pf, lf)); err != nil {
		return protocol.DependencyResult{}, err
	}
	return protocol.DependencyResult{Files: []string{pf}}, nil
}
