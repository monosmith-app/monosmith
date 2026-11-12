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

	"github.com/monosmith-app/monosmith/libs/plugin-dart/discovery"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/templating"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// Runner runs flutter create and pub get; tests replace it.
var Runner = func(dir string, argv []string) error { return toolrun.Run(dir, argv) }

// FlutterStack is the stack whose apps are Flutter apps (for the web).
const FlutterStack = "flutter"

type templateData struct {
	Dir, Name, Package, Kind, Stack string
	Flutter                         bool
}

// identifier is a Dart package name part: lowercase letters, digits and
// underscores, starting with a letter (web-api → web_api).
func identifier(name string) string {
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

// PackageName is a project's pub package name: the workspace's name first,
// so a published package does not clash on pub.dev (acme_core).
func PackageName(workspace, name string) string {
	if workspace == "" {
		return identifier(name)
	}
	return identifier(workspace) + "_" + identifier(name)
}

// Generate writes a package (a Flutter app on the flutter stack), lists it
// in the root pub workspace (creating the workspace the first time), and
// resolves the workspace.
func Generate(workspaceRoot string, params protocol.GenerateParams) (protocol.GenerateResult, error) {
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(params.Directory))
	if _, err := os.Stat(dir); err == nil {
		return protocol.GenerateResult{}, fmt.Errorf("%s already exists", params.Directory)
	}
	flutter := params.Stack == FlutterStack
	if flutter && params.Kind != protocol.ProjectApp {
		return protocol.GenerateResult{}, errors.New("the flutter stack makes apps; a Dart package or library is --lang dart")
	}
	rootManifest := filepath.Join(workspaceRoot, "pubspec.yaml")
	src, err := os.ReadFile(rootManifest)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		src = []byte(NewWorkspace(identifier(params.Workspace.Name), params.Directory))
	case err != nil:
		return protocol.GenerateResult{}, err
	default:
		updated, err := AddMember(string(src), params.Directory)
		if err != nil {
			return protocol.GenerateResult{}, err
		}
		src = []byte(updated)
	}
	data := templateData{Dir: params.Directory, Name: params.Name, Package: cmp.Or(params.PackageName, PackageName(params.Workspace.Name, params.Name)), Kind: string(params.Kind), Stack: params.Stack, Flutter: flutter}
	templates := maps.Clone(sliceTemplates)
	maps.Copy(templates, pubspecTemplate)
	switch {
	case flutter:
		if err := createFlutterApp(workspaceRoot, dir, data.Package); err != nil {
			return protocol.GenerateResult{}, err
		}
		maps.Copy(templates, flutterAppTemplates)
	case params.Kind == protocol.ProjectApp:
		maps.Copy(templates, libraryTemplates)
		maps.Copy(templates, dartAppTemplates)
	default:
		maps.Copy(templates, libraryTemplates)
	}
	files, err := templating.Write(workspaceRoot, templates, data)
	if err != nil {
		return protocol.GenerateResult{}, err
	}
	if flutter {
		if files, err = listTree(workspaceRoot, dir); err != nil {
			return protocol.GenerateResult{}, err
		}
	}
	if err := os.WriteFile(rootManifest, src, 0o644); err != nil {
		return protocol.GenerateResult{}, err
	}
	files = append(files, "pubspec.yaml")
	if added, err := ignoreDartTool(workspaceRoot); err != nil {
		return protocol.GenerateResult{}, err
	} else if added {
		files = append(files, ".gitignore")
	}
	if err := Resolve(workspaceRoot); err != nil {
		return protocol.GenerateResult{Files: files}, fmt.Errorf("resolving the workspace: %w", err)
	}
	return protocol.GenerateResult{Files: files, Notes: []string{"Depend on it from another package with " + data.Package + ": ^0.0.0 under dependencies (the workspace resolves it locally)."}}, nil
}

// createFlutterApp runs flutter create for the web platform files, then
// removes what smith writes its own way.
func createFlutterApp(workspaceRoot, dir, pkg string) error {
	rel, _ := filepath.Rel(workspaceRoot, dir)
	if err := Runner(workspaceRoot, []string{"flutter", "create", "--no-pub", "--empty", "--platforms", "web", "--project-name", pkg, "-t", "app", filepath.ToSlash(rel)}); err != nil {
		return err
	}
	for _, f := range append(flutterCreated, pkg+".iml") {
		if err := os.RemoveAll(filepath.Join(dir, filepath.FromSlash(f))); err != nil {
			return err
		}
	}
	return nil
}

// listTree lists the files under dir, workspace-relative.
func listTree(workspaceRoot, dir string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(workspaceRoot, p)
		files = append(files, filepath.ToSlash(rel))
		return nil
	})
	return files, err
}

// ignoreDartTool adds .dart_tool/ to the workspace's .gitignore: pub and the
// test runner write there, in every package, and it is never a source.
func ignoreDartTool(workspaceRoot string) (bool, error) {
	path := filepath.Join(workspaceRoot, ".gitignore")
	raw, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return false, err
	}
	for _, line := range strings.Split(string(raw), "\n") {
		if l := strings.Trim(strings.TrimSpace(line), "/"); l == ".dart_tool" {
			return false, nil
		}
	}
	text := string(raw)
	if text != "" && !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	return true, os.WriteFile(path, []byte(text+"\n# Dart\n.dart_tool/\n"), 0o644)
}

// Resolve runs pub get for the whole workspace: flutter's when a member is
// a Flutter package, else dart's.
func Resolve(workspaceRoot string) error {
	tool := "dart"
	if usesFlutter(workspaceRoot) {
		tool = "flutter"
	}
	return Runner(workspaceRoot, []string{tool, "pub", "get"})
}

func usesFlutter(workspaceRoot string) bool {
	src, err := os.ReadFile(filepath.Join(workspaceRoot, "pubspec.yaml"))
	if err != nil {
		return false
	}
	root, err := discovery.ParsePubspec(src)
	if err != nil {
		return false
	}
	for _, member := range root.Workspace {
		if raw, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(member), "pubspec.yaml")); err == nil {
			if p, err := discovery.ParsePubspec(raw); err == nil && p.Flutter {
				return true
			}
		}
	}
	return false
}

// Remove takes a removed package out of the workspace's members.
func Remove(workspaceRoot string, params protocol.RemoveParams) (protocol.RemoveResult, error) {
	path := filepath.Join(workspaceRoot, "pubspec.yaml")
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
	return protocol.RemoveResult{Files: []string{"pubspec.yaml"}}, nil
}
