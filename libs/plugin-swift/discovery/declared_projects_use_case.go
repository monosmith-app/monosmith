package discovery

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

var importLine = regexp.MustCompile(`(?m)^[ \t]*(?:@\w+[ \t]+)*import[ \t]+(?:(?:typealias|struct|class|enum|protocol|func|var|let)[ \t]+)?([A-Za-z_]\w*)`)

// DiscoverWithDeclared is Discover plus the projects smith.toml declares: a
// SwiftPM target that is a folder (Sources/<Target>) of the Package.swift
// above it, with no Package.swift of its own. A declaration is refused with an
// error that names it when its folder is missing, has a Package.swift, has
// none above it, or asks for the kind package.
func DiscoverWithDeclared(workspaceRoot string, manifests []string, declared []protocol.DeclaredProject) (protocol.DiscoverResult, error) {
	result, err := Discover(workspaceRoot, manifests)
	if err != nil || len(declared) == 0 {
		return result, err
	}
	for _, d := range declared {
		project, err := describeDeclared(workspaceRoot, d)
		if err != nil {
			return result, fmt.Errorf("[projects.%s] at %s: %w", d.Name, d.Root, err)
		}
		result.Projects = append(result.Projects, project)
	}
	return result, nil
}

// describeDeclared describes one declared folder as the SwiftPM target named
// for it, in the package whose Package.swift is above it. The target is the
// project's export; the modules its sources import are its imports. Its tasks
// build and test that target through --package-path, and the manifest above is
// an input of both, since the folder does not hold it.
func describeDeclared(workspaceRoot string, d protocol.DeclaredProject) (protocol.Project, error) {
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(d.Root))
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return protocol.Project{}, fmt.Errorf("the folder does not exist")
	}
	if _, err := os.Stat(filepath.Join(dir, Manifest)); err == nil {
		return protocol.Project{}, fmt.Errorf("the folder has a Package.swift of its own, so it is a project smith finds itself: remove the declaration")
	}
	manifestDir, src := "", ""
	for up := path.Dir(d.Root); ; up = path.Dir(up) {
		if raw, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(up), Manifest)); err == nil {
			manifestDir, src = up, string(raw)
			break
		}
		if up == "." {
			break
		}
	}
	if src == "" {
		return protocol.Project{}, fmt.Errorf("no Package.swift is above the folder: add one, or give the folder a Package.swift of its own")
	}
	target := path.Base(d.Root)
	kind := d.Kind
	switch kind {
	case "":
		kind = protocol.ProjectLib
		if regexp.MustCompile(`\.executableTarget\(\s*name:\s*"` + regexp.QuoteMeta(target) + `"`).MatchString(src) {
			kind = protocol.ProjectApp
		}
	case protocol.ProjectPackage:
		return protocol.Project{}, fmt.Errorf("a target of a package cannot be released on its own (a Package.swift has one version): declare it as a lib or an app")
	}
	imports, err := scanImports(dir)
	if err != nil {
		return protocol.Project{}, err
	}
	back := strings.TrimSuffix(strings.Repeat("../", strings.Count(d.Root, "/")+1-segments(manifestDir)), "/")
	manifestFile := "{workspaceRoot}/" + Manifest
	if manifestDir != "." {
		manifestFile = "{workspaceRoot}/" + manifestDir + "/" + Manifest
	}
	resolved := strings.TrimSuffix(manifestFile, Manifest) + "Package.resolved"
	tests := strings.TrimSuffix(manifestFile, Manifest) + "Tests/" + target + "Tests/**/*.swift"
	out := protocol.BuildDirFromProject(d.Root)
	sources := []string{"**/*.swift", manifestFile, resolved}
	return protocol.Project{
		Name:     d.Name,
		Root:     d.Root,
		Language: "swift",
		Kind:     kind,
		Exports:  []string{target},
		Imports:  imports,
		Tasks: map[string]protocol.Task{
			"build": {Command: []string{"swift", "build", "--package-path", back, "--target", target, "--scratch-path", out}, Inputs: sources, Outputs: []string{out}, Cache: true, Description: "swift build --target " + target + ", into dist/"},
			"test":  {Command: []string{"swift", "test", "--package-path", back, "--filter", "^" + target + "Tests\\."}, Inputs: append(append([]string{}, sources...), tests), Cache: true, Description: "swift test, the tests of " + target},
		},
	}, nil
}

// segments is the number of folders in a workspace-relative path ("." has none).
func segments(p string) int {
	if p == "." {
		return 0
	}
	return strings.Count(p, "/") + 1
}

// scanImports returns the module names the .swift files under dir import,
// sorted and unique.
func scanImports(dir string) ([]string, error) {
	seen := map[string]bool{}
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) && p != dir {
				return nil // removed while the walk ran
			}
			return err
		}
		if d.IsDir() {
			if p != dir && (d.Name() == ".build" || strings.HasPrefix(d.Name(), ".")) {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(p) != ".swift" {
			return nil
		}
		src, err := os.ReadFile(p)
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		if err != nil {
			return err
		}
		for _, m := range importLine.FindAllStringSubmatch(string(src), -1) {
			seen[m[1]] = true
		}
		return nil
	})
	out := make([]string, 0, len(seen))
	for name := range seen {
		out = append(out, name)
	}
	sort.Strings(out)
	return out, err
}
