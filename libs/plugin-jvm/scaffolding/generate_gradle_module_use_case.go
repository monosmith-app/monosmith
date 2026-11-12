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

	"github.com/monosmith-app/monosmith/libs/plugin-jvm/discovery"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/projectnaming"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/templating"
)

// gradleState is what Gradle and the Kotlin compiler keep in the workspace
// root: caches, never sources.
var gradleState = []string{".gradle", ".kotlin"}

// GenerateGradle writes a Kotlin module (the kotlin stack) or an Android
// app (the android stack), includes it in the settings, and writes the
// settings and root script the first time; an Android app also pins AGP in
// the root script.
func GenerateGradle(workspaceRoot string, params protocol.GenerateParams) (protocol.GenerateResult, error) {
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(params.Directory))
	if _, err := os.Stat(dir); err == nil {
		return protocol.GenerateResult{}, fmt.Errorf("%s already exists", params.Directory)
	}
	android := params.Stack == AndroidStack
	if android && params.Kind != protocol.ProjectApp {
		return protocol.GenerateResult{}, fmt.Errorf("the android stack writes apps; a library shared with an Android app is a Kotlin module: smith add %s %s --stack %s", params.Kind, params.Name, KotlinStack)
	}
	settingsFile := filepath.Join(workspaceRoot, settings)
	workspace := params.Workspace.Name
	// The artifact starts with the namespace, or with the existing build's name.
	naming := params.Workspace
	templates := maps.Clone(kotlinSliceTemplates)
	maps.Copy(templates, gradleModuleTemplates)
	if params.Kind == protocol.ProjectApp {
		maps.Copy(templates, kotlinAppTemplates)
	}
	if android {
		templates = maps.Clone(androidTemplates)
	}
	src, err := os.ReadFile(settingsFile)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		maps.Copy(templates, gradleRootTemplates)
	case err != nil:
		return protocol.GenerateResult{}, err
	default:
		if name := RootProjectName(string(src)); name != "" {
			workspace, naming.Namespace = name, name
		}
	}
	data := templateData{
		Dir: params.Directory, Name: params.Name, Workspace: workspace, Group: projectnaming.GroupID(naming),
		ArtifactID: cmp.Or(params.PackageName, projectnaming.Name(naming, projectnaming.Maven, params.Name)), Kind: string(params.Kind), Stack: params.Stack,
	}
	data.Package = data.Group + "." + projectnaming.Flat(params.Name)
	data.PackagePath = strings.ReplaceAll(data.Package, ".", "/")
	files, err := templating.Write(workspaceRoot, templates, data)
	if err != nil {
		return protocol.GenerateResult{}, err
	}
	if src, err = os.ReadFile(settingsFile); err != nil {
		return protocol.GenerateResult{}, err
	}
	path := discovery.ProjectPath(params.Directory)
	if out, changed := AddInclude(string(src), path); changed {
		if err := os.WriteFile(settingsFile, []byte(out), 0o644); err != nil {
			return protocol.GenerateResult{}, err
		}
		if !slices.Contains(files, settings) {
			files = append(files, settings)
		}
	}
	if android {
		changed, err := pinAndroidPlugin(workspaceRoot)
		if err != nil {
			return protocol.GenerateResult{}, err
		}
		if changed && !slices.Contains(files, rootScript) {
			files = append(files, rootScript)
		}
	}
	if added, err := ignoreGradleState(workspaceRoot); err != nil {
		return protocol.GenerateResult{}, err
	} else if added {
		files = append(files, ".gitignore")
	}
	return protocol.GenerateResult{Files: files, Notes: []string{"Depend on it from another module with implementation(project(\"" + path + "\"))."}}, nil
}

// pinAndroidPlugin pins AGP in the root script, once.
func pinAndroidPlugin(workspaceRoot string) (bool, error) {
	file := filepath.Join(workspaceRoot, rootScript)
	src, err := os.ReadFile(file)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return false, err
	}
	out, changed := AddRootPlugin(string(src), androidPlugin)
	if !changed {
		return false, nil
	}
	return true, os.WriteFile(file, []byte(out), 0o644)
}

// ignoreGradleState adds Gradle's and Kotlin's workspace caches to the
// .gitignore, so they never count as sources.
func ignoreGradleState(workspaceRoot string) (bool, error) {
	file := filepath.Join(workspaceRoot, ".gitignore")
	raw, err := os.ReadFile(file)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return false, err
	}
	ignored := map[string]bool{}
	for _, line := range strings.Split(string(raw), "\n") {
		ignored[strings.Trim(strings.TrimSpace(line), "/")] = true
	}
	var missing []string
	for _, dir := range gradleState {
		if !ignored[dir] {
			missing = append(missing, dir+"/")
		}
	}
	if len(missing) == 0 {
		return false, nil
	}
	text := string(raw)
	if text != "" && !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	return true, os.WriteFile(file, []byte(text+"\n# Gradle\n"+strings.Join(missing, "\n")+"\n"), 0o644)
}
