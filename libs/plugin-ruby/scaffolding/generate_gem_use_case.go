package scaffolding

import (
	"cmp"
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

// Runner runs bundle; tests replace it.
var Runner = func(dir string, argv []string) error { return toolrun.Run(dir, argv) }

type templateData struct {
	Dir, Name, Workspace, Gem, Module, Kind, Stack string
}

// snake is a gem name part: lowercase letters, digits and underscores,
// starting with a letter (web-api → web_api).
func snake(name string) string {
	var b strings.Builder
	for _, c := range strings.ToLower(name) {
		switch {
		case unicode.IsLetter(c) || unicode.IsDigit(c):
			if b.Len() == 0 && unicode.IsDigit(c) {
				b.WriteByte('x')
			}
			b.WriteRune(c)
		case b.Len() > 0 && !strings.HasSuffix(b.String(), "_"):
			b.WriteByte('_')
		}
	}
	return strings.TrimSuffix(b.String(), "_")
}

// GemName is a project's gem: the workspace's name first, so a published
// gem does not clash on RubyGems (acme_core).
// workspace is the package prefix (protocol.WorkspaceInfo.PackagePrefix: the scope, else the name).
func GemName(workspace, name string) string {
	if workspace == "" {
		return snake(name)
	}
	return snake(workspace) + "_" + snake(name)
}

// moduleName is a gem's Ruby module: acme_core → AcmeCore.
func moduleName(gem string) string {
	var b strings.Builder
	for _, part := range strings.Split(gem, "_") {
		if part != "" {
			b.WriteString(strings.ToUpper(part[:1]) + part[1:])
		}
	}
	return b.String()
}

// Generate writes a gem, lists it in the workspace's Gemfile (writing the
// Gemfile and .rubocop.yml the first time), and installs the bundle.
func Generate(workspaceRoot string, params protocol.GenerateParams) (protocol.GenerateResult, error) {
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(params.Directory))
	if _, err := os.Stat(dir); err == nil {
		return protocol.GenerateResult{}, fmt.Errorf("%s already exists", params.Directory)
	}
	gem := cmp.Or(params.PackageName, GemName(params.Workspace.PackagePrefix(), params.Name))
	data := templateData{Dir: params.Directory, Name: params.Name, Workspace: params.Workspace.Name, Gem: gem, Module: moduleName(gem), Kind: string(params.Kind), Stack: params.Stack}
	templates := maps.Clone(sliceTemplates)
	maps.Copy(templates, gemTemplate)
	if params.Kind == protocol.ProjectApp {
		maps.Copy(templates, appTemplates)
	}
	for name, body := range rootTemplates {
		if _, err := os.Stat(filepath.Join(workspaceRoot, name)); errors.Is(err, fs.ErrNotExist) {
			templates[name] = body
		} else if err != nil {
			return protocol.GenerateResult{}, err
		}
	}
	files, err := templating.Write(workspaceRoot, templates, data)
	if err != nil {
		return protocol.GenerateResult{}, err
	}
	path := filepath.Join(workspaceRoot, gemfile)
	src, err := os.ReadFile(path)
	if err != nil {
		return protocol.GenerateResult{}, err
	}
	if out := AddGemspec(string(src), params.Directory); out != string(src) {
		if err := os.WriteFile(path, []byte(out), 0o644); err != nil {
			return protocol.GenerateResult{}, err
		}
		if _, written := templates[gemfile]; !written {
			files = append(files, gemfile)
		}
	}
	if err := Runner(workspaceRoot, []string{"bundle", "install"}); err != nil {
		return protocol.GenerateResult{Files: files}, fmt.Errorf("bundle install: %w", err)
	}
	return protocol.GenerateResult{Files: append(files, "Gemfile.lock"), Notes: []string{"Depend on it from another gem with spec.add_dependency '" + gem + "' (the workspace's Gemfile resolves it locally)."}}, nil
}

// Remove takes a removed gem out of the workspace's Gemfile.
func Remove(workspaceRoot string, params protocol.RemoveParams) (protocol.RemoveResult, error) {
	path := filepath.Join(workspaceRoot, gemfile)
	src, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return protocol.RemoveResult{}, nil
	}
	if err != nil {
		return protocol.RemoveResult{}, err
	}
	out, changed := DropGemspec(string(src), params.Project.Root)
	if !changed {
		return protocol.RemoveResult{}, nil
	}
	if !params.DryRun {
		if err := os.WriteFile(path, []byte(out), 0o644); err != nil {
			return protocol.RemoveResult{}, err
		}
	}
	return protocol.RemoveResult{Files: []string{gemfile}}, nil
}
