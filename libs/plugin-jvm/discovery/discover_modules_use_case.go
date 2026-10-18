package discovery

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/projectkind"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// Discover describes the module behind each pom.xml and each Gradle build
// script. The root pom (the aggregator and parent), any other pom-packaged
// module, the root build script and buildSrc only organise the build, so
// they are not projects.
func Discover(workspaceRoot string, manifests []string) (protocol.DiscoverResult, error) {
	result := protocol.DiscoverResult{Projects: []protocol.Project{}}
	for _, manifest := range manifests {
		root := path.Dir(manifest)
		if root == "." || slices.Contains(strings.Split(root, "/"), "buildSrc") {
			continue
		}
		src, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(manifest)))
		if err != nil {
			return result, err
		}
		var project protocol.Project
		var ok bool
		if path.Base(manifest) == "pom.xml" {
			project, ok, err = mavenModule(root, src)
		} else {
			project, ok, err = gradleModule(workspaceRoot, root, src)
		}
		if err != nil {
			return result, fmt.Errorf("%s: %w", manifest, err)
		}
		if ok {
			result.Projects = append(result.Projects, project)
		}
	}
	result.SetToolchain(func() string { return toolrun.Toolchain("java", "java", "-version") })
	return result, nil
}

func mavenModule(root string, src []byte) (protocol.Project, bool, error) {
	pom, err := ParsePom(src)
	if err != nil || pom.Packaging == "pom" {
		return protocol.Project{}, false, err
	}
	kind := projectkind.Resolve(pom.Kind, root, protocol.ProjectLib)
	return protocol.Project{
		Name:        path.Base(root),
		Root:        root,
		Language:    "jvm",
		Kind:        kind,
		Exports:     []string{pom.Coordinates()},
		Imports:     pom.Dependencies,
		Version:     pom.Version,
		Description: pom.Description,
		Tasks:       Tasks(root, pom.Coordinates(), pom.ArtifactID, kind),
		Stack:       pom.Stack,
	}, true, nil
}

// gradleModule describes a Gradle module. It exports its project path, what
// other modules depend on with project(...), and its coordinates when it
// publishes.
func gradleModule(workspaceRoot, root string, script []byte) (protocol.Project, bool, error) {
	properties, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(root), "gradle.properties"))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return protocol.Project{}, false, err
	}
	b := ParseGradleBuild(root, script, properties)
	exports := []string{b.Path}
	if c := b.Coordinates(); c != "" {
		exports = append(exports, c)
	}
	kind := projectkind.Resolve(b.Kind, root, protocol.ProjectLib)
	return protocol.Project{
		Name:     path.Base(root),
		Root:     root,
		Language: "jvm",
		Kind:     kind,
		Exports:  exports,
		Imports:  b.Projects,
		Version:  b.Version,
		Tasks:    GradleTasks(root, b, kind),
		Stack:    b.Stack,
	}, true, nil
}
