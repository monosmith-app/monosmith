package scaffolding

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/templating"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// Runner runs composer, showing its output only when it fails or smith add
// --verbose asks; tests replace it.
var Runner = runComposer

func runComposer(dir string, argv []string) error { return toolrun.Quiet(dir, argv) }

type templateData struct {
	Dir, Up, Name, Vendor, Package, Namespace, NamespaceJSON, Kind, Stack string
}

// kebab is a Composer name part: lowercase letters, digits and dashes.
func kebab(name string) string {
	var b strings.Builder
	for _, c := range strings.ToLower(name) {
		switch {
		case unicode.IsLetter(c) || unicode.IsDigit(c):
			b.WriteRune(c)
		case b.Len() > 0 && !strings.HasSuffix(b.String(), "-"):
			b.WriteByte('-')
		}
	}
	return strings.TrimSuffix(b.String(), "-")
}

// pascal is a namespace segment: web-api → WebApi, 3d → X3d.
func pascal(name string) string {
	var b strings.Builder
	upper := true
	for _, c := range name {
		if !unicode.IsLetter(c) && !unicode.IsDigit(c) {
			upper = true
			continue
		}
		if b.Len() == 0 && unicode.IsDigit(c) {
			b.WriteByte('X')
		}
		if upper {
			c = unicode.ToUpper(c)
			upper = false
		}
		b.WriteRune(c)
	}
	return b.String()
}

// PackageName is a project's Composer name: the workspace is the vendor
// (acme/core).
func PackageName(workspace, name string) string {
	vendor := kebab(workspace)
	if vendor == "" {
		vendor = kebab(name)
	}
	return vendor + "/" + kebab(name)
}

// RepositoryKey names a package's path repository in the root
// composer.json: its directory, dashed (apps/cli → apps-cli).
func RepositoryKey(dir string) string { return strings.ReplaceAll(dir, "/", "-") }

// Generate writes a package, makes the root composer.json require it from
// its directory (writing the root the first time), and updates vendor/.
func Generate(workspaceRoot string, params protocol.GenerateParams) (protocol.GenerateResult, error) {
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(params.Directory))
	if _, err := os.Stat(dir); err == nil {
		return protocol.GenerateResult{}, fmt.Errorf("%s already exists", params.Directory)
	}
	pkg := PackageName(params.Workspace.Name, params.Name)
	namespace := pascal(params.Workspace.Name) + `\` + pascal(params.Name)
	if params.Workspace.Name == "" {
		namespace = pascal(params.Name)
	}
	data := templateData{
		Dir: params.Directory, Up: strings.Repeat("../", strings.Count(params.Directory, "/")+2), Name: params.Name,
		Vendor: strings.Split(pkg, "/")[0], Package: pkg, Namespace: namespace, NamespaceJSON: strings.ReplaceAll(namespace, `\`, `\\`),
		Kind: string(params.Kind), Stack: params.Stack,
	}
	templates := maps.Clone(sliceTemplates)
	maps.Copy(templates, manifestTemplate)
	if params.Kind == protocol.ProjectApp {
		maps.Copy(templates, appTemplates)
	}
	if _, err := os.Stat(filepath.Join(workspaceRoot, rootManifest)); errors.Is(err, fs.ErrNotExist) {
		maps.Copy(templates, rootTemplates)
	} else if err != nil {
		return protocol.GenerateResult{}, err
	}
	files, err := templating.Write(workspaceRoot, templates, data)
	if err != nil {
		return protocol.GenerateResult{}, err
	}
	if _, written := templates[rootManifest]; !written {
		files = append(files, rootManifest)
	}
	if added, err := ignoreVendor(workspaceRoot); err != nil {
		return protocol.GenerateResult{}, err
	} else if added {
		files = append(files, ".gitignore")
	}
	for _, argv := range [][]string{
		{"composer", "config", "repositories." + RepositoryKey(params.Directory), "path", params.Directory},
		{"composer", "require", pkg + ":@dev", "--no-update", "--no-interaction"},
		{"composer", "update", "--no-interaction", "--no-progress"},
	} {
		if err := Runner(workspaceRoot, argv); err != nil {
			return protocol.GenerateResult{Files: files}, fmt.Errorf("%s: %w", strings.Join(argv, " "), err)
		}
	}
	return protocol.GenerateResult{Files: append(files, "composer.lock"), Notes: []string{"Require it from another package with \"" + pkg + "\": \"*\" (the workspace resolves it from its directory)."}}, nil
}

// ignoreVendor adds Composer's vendor/ to the workspace's .gitignore.
func ignoreVendor(workspaceRoot string) (bool, error) {
	path := filepath.Join(workspaceRoot, ".gitignore")
	raw, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return false, err
	}
	for _, line := range strings.Split(string(raw), "\n") {
		if l := strings.Trim(strings.TrimSpace(line), "/"); l == "vendor" {
			return false, nil
		}
	}
	text := string(raw)
	if text != "" && !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	return true, os.WriteFile(path, []byte(text+"\n# PHP\n/vendor/\n"), 0o644)
}

// Remove takes a removed package out of the root composer.json (its
// requirement and its path repository) and out of composer.lock, which would
// otherwise still name it and make composer install fail.
func Remove(workspaceRoot string, params protocol.RemoveParams) (protocol.RemoveResult, error) {
	if _, err := os.Stat(filepath.Join(workspaceRoot, rootManifest)); errors.Is(err, fs.ErrNotExist) {
		return protocol.RemoveResult{}, nil
	}
	res := protocol.RemoveResult{Files: []string{rootManifest, "composer.lock"}}
	if params.DryRun {
		return res, nil
	}
	name := params.Project.Name
	if len(params.Project.Exports) > 0 {
		name = params.Project.Exports[0]
	}
	for _, argv := range [][]string{
		{"composer", "remove", name, "--no-update", "--no-interaction"},
		{"composer", "config", "--unset", "repositories." + RepositoryKey(params.Project.Root)},
		{"composer", "update", "--no-interaction", "--no-progress"},
	} {
		if err := Runner(workspaceRoot, argv); err != nil {
			return res, err
		}
	}
	return res, nil
}
