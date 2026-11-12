package scaffolding

import (
	"cmp"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"unicode"

	"github.com/monosmith-app/monosmith/libs/plugin-jvm/discovery"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/templating"
)

type templateData struct {
	Dir, Up, Name, Workspace, Group, ArtifactID, Package, PackagePath, Kind, Stack string
}

// identifier keeps a name's letters and digits, lowercased: a Java package
// segment (web-api → webapi), which starts with a letter (3d → x3d).
func identifier(name string) string {
	var b strings.Builder
	for _, c := range strings.ToLower(name) {
		if unicode.IsLetter(c) || unicode.IsDigit(c) {
			if b.Len() == 0 && unicode.IsDigit(c) {
				b.WriteByte('x')
			}
			b.WriteRune(c)
		}
	}
	return b.String()
}

// GroupID is a new workspace's Maven groupId, dev.<workspace>: a placeholder
// for the owner's reverse domain, which a repository like Maven Central
// verifies before it accepts a deploy.
func GroupID(workspace string) string { return "dev." + identifier(workspace) }

// Generate writes a module, lists it in the root pom, and writes the root
// pom the first time. The kotlin and android stacks write a Gradle module
// instead.
func Generate(workspaceRoot string, params protocol.GenerateParams) (protocol.GenerateResult, error) {
	if params.Stack == KotlinStack || params.Stack == AndroidStack {
		return GenerateGradle(workspaceRoot, params)
	}
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(params.Directory))
	if _, err := os.Stat(dir); err == nil {
		return protocol.GenerateResult{}, fmt.Errorf("%s already exists", params.Directory)
	}
	workspace := params.Workspace.Name
	data := templateData{
		Dir: params.Directory, Up: strings.Repeat("../", strings.Count(params.Directory, "/")+1),
		Name: params.Name, Workspace: workspace, Group: GroupID(workspace),
		ArtifactID: cmp.Or(params.PackageName, workspace+"-"+params.Name), Kind: string(params.Kind), Stack: params.Stack,
	}
	templates := maps.Clone(sliceTemplates)
	maps.Copy(templates, moduleTemplate)
	if params.Kind == protocol.ProjectApp {
		maps.Copy(templates, appTemplates)
	}
	rootFile := filepath.Join(workspaceRoot, rootPom)
	src, err := os.ReadFile(rootFile)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		maps.Copy(templates, rootTemplates)
	case err != nil:
		return protocol.GenerateResult{}, err
	default:
		root, err := discovery.ParsePom(src)
		if err != nil {
			return protocol.GenerateResult{}, fmt.Errorf("%s: %w", rootPom, err)
		}
		data.Group, data.Workspace = root.GroupID, strings.TrimSuffix(root.ArtifactID, "-parent")
	}
	data.Package = data.Group + "." + identifier(params.Name)
	data.PackagePath = strings.ReplaceAll(data.Package, ".", "/")
	files, err := templating.Write(workspaceRoot, templates, data)
	if err != nil {
		return protocol.GenerateResult{}, err
	}
	if src, err = os.ReadFile(rootFile); err != nil {
		return protocol.GenerateResult{}, err
	}
	if out, changed := AddModule(string(src), params.Directory); changed {
		if err := os.WriteFile(rootFile, []byte(out), 0o644); err != nil {
			return protocol.GenerateResult{}, err
		}
		if !slices.Contains(files, rootPom) {
			files = append(files, rootPom)
		}
	}
	return protocol.GenerateResult{Files: files, Notes: []string{"Depend on it from another module with <dependency><groupId>" + data.Group + "</groupId><artifactId>" + data.ArtifactID + "</artifactId><version>0.0.0</version></dependency>."}}, nil
}
