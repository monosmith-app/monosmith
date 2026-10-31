package scaffolding

import (
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/templating"
	"github.com/pelletier/go-toml/v2"
)

// FileRenderer derives one file of a project that a plugin scaffolded, from the
// plugin's own template (ADR 0049). The template is the specification of the
// file, so a deleted file comes back whole and as the scaffold wrote it.
type FileRenderer struct {
	// ID is the stable id of the renderer, for example "typescript.tsconfig".
	ID string
	// Version changes when the template changes its output.
	Version int
	// File is the project-relative path of the file, for example "tsconfig.json".
	File string
	// Templates returns the template set that scaffolds the project, or false
	// when the plugin did not scaffold it. params carries what a person cannot
	// change after the project exists: kind, name, directory, stack, shape and
	// facts.
	Templates func(params protocol.GenerateParams) (TemplateSet, bool)
}

// Derive answers with the file of every project the renderer serves. A project
// is served when the record owns the file, or when the file is, line for line,
// what the template renders: a file that differs and is not owned was written
// by a person, and is left alone. It writes nothing.
func (r FileRenderer) Derive(workspaceRoot string, params protocol.DeriveParams) (protocol.DeriveResult, error) {
	var files []protocol.DerivedFile
	for _, p := range params.Projects {
		rel := path.Join(p.Root, r.File)
		want, ok := r.render(workspaceRoot, p, rel)
		if !ok {
			continue
		}
		if !slices.Contains(params.Owned, rel) && !sameContent(workspaceRoot, rel, want) {
			continue
		}
		files = append(files, protocol.DerivedFile{Project: p.Name, Path: rel, Mode: protocol.DeriveFile, Renderer: r.ID, Version: r.Version, Content: want})
	}
	return protocol.DeriveResult{Files: files}, nil
}

// render is the content of the file the template writes for the project; ok is
// false when the plugin does not scaffold the project or has no such template.
func (r FileRenderer) render(workspaceRoot string, p protocol.Project, rel string) (string, bool) {
	gp := protocol.GenerateParams{Kind: p.Kind, Name: p.Name, Directory: p.Root, Stack: p.Stack, Shape: p.Shape, Facts: p.Facts}
	set, ok := r.Templates(gp)
	if !ok {
		return "", false
	}
	templates, err := kindTemplates(p.Kind, set)
	if err != nil {
		return "", false
	}
	data := newTemplateData(workspaceRoot, gp, set)
	for target, text := range templates {
		name, err := templating.Render("path", target, data)
		if err != nil || path.Clean(string(name)) != rel {
			continue
		}
		body, err := templating.Render(rel, text, data)
		if err != nil {
			return "", false
		}
		return string(body), true
	}
	return "", false
}

// sameContent reports whether the file under the workspace root holds the
// content, with line endings of either kind.
func sameContent(workspaceRoot, rel, content string) bool {
	disk, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(rel)))
	return err == nil && strings.ReplaceAll(string(disk), "\r\n", "\n") == strings.ReplaceAll(content, "\r\n", "\n")
}

// StackChoices are the choices of the stack that smith.toml declares as a saved
// selection ([stacks.<name>] choices), or nil when it declares none or the file
// cannot be read. A stack that a plugin publishes is not in smith.toml: the
// plugin knows its own.
func StackChoices(workspaceRoot, stack string) []string {
	raw, err := os.ReadFile(filepath.Join(workspaceRoot, "smith.toml"))
	if err != nil {
		return nil
	}
	var cfg struct {
		Stacks map[string]struct {
			Choices []string `toml:"choices"`
		} `toml:"stacks"`
	}
	if toml.Unmarshal(raw, &cfg) != nil {
		return nil
	}
	return cfg.Stacks[stack].Choices
}
