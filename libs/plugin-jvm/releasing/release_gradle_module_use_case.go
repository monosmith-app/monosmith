package releasing

import (
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"

	"github.com/monosmith-app/monosmith/libs/plugin-jvm/discovery"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// GradleRunner runs gradle; tests replace it.
var GradleRunner = func(dir string, argv []string, env ...string) error { return toolrun.Run(dir, argv, env...) }

// gradleScript is the workspace-relative build script of the Gradle module
// in root, or empty for a Maven module.
func gradleScript(workspaceRoot, root string) string {
	for _, name := range []string{"build.gradle.kts", "build.gradle"} {
		rel := path.Join(root, name)
		if _, err := os.Stat(filepath.Join(workspaceRoot, filepath.FromSlash(rel))); err == nil {
			return rel
		}
	}
	return ""
}

// writeGradleVersion records the version where the module keeps it: its
// script, or its gradle.properties when only that sets one. Modules that
// use it depend on its project, with no version to update.
func writeGradleVersion(workspaceRoot, script string, params protocol.VersionParams) (protocol.VersionResult, error) {
	result := protocol.VersionResult{Files: []string{}}
	src, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(script)))
	if err != nil {
		return result, err
	}
	rel, out, changed := script, "", false
	if scriptVersion.Match(src) {
		out, changed = SetScriptVersion(string(src), params.Version)
	} else {
		props := path.Join(params.Project.Root, "gradle.properties")
		raw, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(props)))
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return result, err
		}
		if propertyVersion.Match(raw) {
			rel = props
			out, changed = SetPropertyVersion(string(raw), params.Version)
		} else {
			out, changed = SetScriptVersion(string(src), params.Version)
		}
	}
	if !changed {
		return result, nil
	}
	result.Files = append(result.Files, rel)
	if params.DryRun {
		return result, nil
	}
	return result, os.WriteFile(filepath.Join(workspaceRoot, filepath.FromSlash(rel)), []byte(out), 0o644)
}

// publishGradle publishes the module's publications to the repository its
// script declares as smith (from MAVEN_REPOSITORY_URL, MAVEN_USERNAME and
// MAVEN_PASSWORD, which env overrides when smith.toml names the registry).
func publishGradle(workspaceRoot string, project protocol.Project, result protocol.PublishResult, env []string) (protocol.PublishResult, error) {
	argv := []string{"gradle", "--console=plain", "-q", discovery.ProjectPath(project.Root) + ":publishAllPublicationsToSmithRepository"}
	if err := GradleRunner(workspaceRoot, argv, env...); err != nil {
		return result, err
	}
	result.Published = true
	return result, nil
}
