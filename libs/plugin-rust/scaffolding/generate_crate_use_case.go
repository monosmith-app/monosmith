package scaffolding

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-rust/discovery"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/templating"
)

type templateData struct {
	Dir, Name, Crate, Kind, Stack string
}

// CrateName is a project's crate name: the workspace's name first, so a
// published crate does not clash on crates.io (acme-core).
func CrateName(workspace, name string) string {
	workspace = crateIdentifier(workspace)
	if workspace == "" || strings.HasPrefix(name, workspace+"-") {
		return name
	}
	return workspace + "-" + name
}

// crateIdentifier keeps what Cargo accepts in a package name (letters,
// digits, dashes, underscores): any other run of characters, such as the dots
// of a workspace named jato.auto.packages, becomes one dash.
func crateIdentifier(name string) string {
	var b strings.Builder
	for _, c := range strings.ToLower(name) {
		switch {
		case c >= 'a' && c <= 'z', c >= '0' && c <= '9', c == '_':
			b.WriteRune(c)
		case b.Len() > 0 && !strings.HasSuffix(b.String(), "-"):
			b.WriteByte('-')
		}
	}
	return strings.TrimSuffix(b.String(), "-")
}

// Generate writes a crate, lists it in the root Cargo.toml's workspace
// members (creating the workspace the first time), and points Cargo's
// target directory at dist/cargo.
func Generate(workspaceRoot string, params protocol.GenerateParams) (protocol.GenerateResult, error) {
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(params.Directory))
	if _, err := os.Stat(dir); err == nil {
		return protocol.GenerateResult{}, fmt.Errorf("%s already exists", params.Directory)
	}
	rootManifest := filepath.Join(workspaceRoot, "Cargo.toml")
	var files []string
	src, err := os.ReadFile(rootManifest)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		src = []byte(NewWorkspace(params.Directory))
	case err != nil:
		return protocol.GenerateResult{}, err
	default:
		if crate, err := discovery.ParseCargo(src); err == nil && crate.Name != "" {
			return protocol.GenerateResult{}, fmt.Errorf("the root Cargo.toml is a crate (%s), not a workspace: make it a [workspace] first", crate.Name)
		}
		updated, err := AddMember(string(src), params.Directory)
		if err != nil {
			return protocol.GenerateResult{}, err
		}
		src = []byte(updated)
	}
	templates := maps.Clone(sliceTemplates)
	maps.Copy(templates, manifestTemplate)
	if params.Kind == protocol.ProjectApp {
		maps.Copy(templates, appTemplates)
	} else {
		maps.Copy(templates, libTemplates)
	}
	data := templateData{Dir: params.Directory, Name: params.Name, Crate: CrateName(params.Workspace.Name, params.Name), Kind: string(params.Kind), Stack: params.Stack}
	written, err := templating.Write(workspaceRoot, templates, data)
	if err != nil {
		return protocol.GenerateResult{}, err
	}
	files = append(files, written...)
	if err := os.WriteFile(rootManifest, src, 0o644); err != nil {
		return protocol.GenerateResult{}, err
	}
	files = append(files, "Cargo.toml")
	if added, err := buildIntoDist(workspaceRoot); err != nil {
		return protocol.GenerateResult{}, err
	} else if added {
		files = append(files, cargoConfig)
	}
	return protocol.GenerateResult{Files: files, Notes: []string{"Import it from a crate in another top-level folder as " + data.Crate + " = { path = \"../../" + params.Directory + "\" }."}}, nil
}

// cargoConfig is the workspace's Cargo configuration.
const cargoConfig = ".cargo/config.toml"

// targetDir is where every Cargo build goes: the workspace's dist/, like
// every other build (smith), never target/ beside the sources.
var targetDir = protocol.BuildDir + "/cargo"

// buildIntoDist writes .cargo/config.toml with the target directory the
// first time; a workspace that has one keeps its own.
func buildIntoDist(workspaceRoot string) (bool, error) {
	path := filepath.Join(workspaceRoot, filepath.FromSlash(cargoConfig))
	if _, err := os.Stat(path); err == nil || !errors.Is(err, fs.ErrNotExist) {
		return false, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return false, err
	}
	body := "# Every build goes to the workspace's dist/ (smith).\n[build]\ntarget-dir = \"" + targetDir + "\"\n"
	return true, os.WriteFile(path, []byte(body), 0o644)
}

// Remove takes a removed crate out of the workspace's members.
func Remove(workspaceRoot string, params protocol.RemoveParams) (protocol.RemoveResult, error) {
	path := filepath.Join(workspaceRoot, "Cargo.toml")
	src, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return protocol.RemoveResult{}, nil
	}
	if err != nil {
		return protocol.RemoveResult{}, err
	}
	out, changed := DropMember(string(src), params.Project.Root)
	if !changed {
		return protocol.RemoveResult{}, nil
	}
	if !params.DryRun {
		if err := os.WriteFile(path, []byte(out), 0o644); err != nil {
			return protocol.RemoveResult{}, err
		}
	}
	return protocol.RemoveResult{Files: []string{"Cargo.toml"}}, nil
}
