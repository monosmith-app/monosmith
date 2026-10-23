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

// GoRunner runs the go command; tests replace it.
var GoRunner = func(dir string, argv []string) error { return toolrun.Run(dir, argv) }

type templateData struct {
	Dir, Name, Module, Package string
}

// Generate scaffolds one Go module from the greeting starter and adds it to
// go.work (creating go.work the first time).
func Generate(workspaceRoot string, params protocol.GenerateParams) (protocol.GenerateResult, error) {
	templates := maps.Clone(sliceTemplates)
	if params.Kind == protocol.ProjectApp {
		maps.Copy(templates, appTemplates)
	} else {
		maps.Copy(templates, libTemplates)
	}
	res, err := GenerateFrom(workspaceRoot, params, templates, nil)
	if err == nil {
		res.Notes = append(res.Notes, "Import it from another Go project as "+modulePath(params)+"/<slice>.")
	}
	return res, err
}

func modulePath(params protocol.GenerateParams) string {
	prefix := params.Workspace.ModulePrefix
	if prefix == "" {
		prefix = "example.com/" + strings.ToLower(params.Workspace.Name)
	}
	return prefix + "/" + params.Directory
}

// GenerateFrom scaffolds one Go module from templates (paths and contents use
// [[.Dir]], [[.Name]], [[.Module]] and [[.Package]]), requires the modules
// listed ("path@version"), and adds it to go.work.
func GenerateFrom(workspaceRoot string, params protocol.GenerateParams, templates map[string]string, require []string) (protocol.GenerateResult, error) {
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(params.Directory))
	if _, err := os.Stat(dir); err == nil {
		return protocol.GenerateResult{}, fmt.Errorf("%s already exists", params.Directory)
	}
	data := templateData{Dir: params.Directory, Name: params.Name, Module: modulePath(params), Package: PackageName(params.Name)}

	var files []string
	// A module added to an existing workspace declares the Go version the
	// workspace builds with, not the version of the toolchain that runs this:
	// go mod init writes the latter, and go work use then raises go.work to it.
	workspaceGo := WorkspaceGoVersion(workspaceRoot)
	if _, err := os.Stat(filepath.Join(workspaceRoot, "go.work")); errors.Is(err, fs.ErrNotExist) {
		if err := GoRunner(workspaceRoot, []string{"go", "work", "init"}); err != nil {
			return protocol.GenerateResult{}, err
		}
		files = append(files, "go.work")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return protocol.GenerateResult{}, err
	}
	if err := GoRunner(dir, []string{"go", "mod", "init", data.Module}); err != nil {
		return protocol.GenerateResult{}, err
	}
	if workspaceGo != "" {
		if err := GoRunner(dir, []string{"go", "mod", "edit", "-go=" + workspaceGo}); err != nil {
			return protocol.GenerateResult{}, err
		}
	}
	if err := declareKind(filepath.Join(dir, "go.mod"), params.Kind, params.Stack); err != nil {
		return protocol.GenerateResult{}, err
	}
	files = append(files, params.Directory+"/go.mod")

	written, err := templating.Write(workspaceRoot, templates, data)
	if err != nil {
		return protocol.GenerateResult{}, err
	}
	files = append(files, written...)
	for _, module := range require {
		if err := GoRunner(dir, []string{"go", "get", module}); err != nil {
			return protocol.GenerateResult{Files: files}, err
		}
	}
	if len(require) > 0 {
		if err := GoRunner(dir, []string{"go", "mod", "tidy"}); err != nil {
			return protocol.GenerateResult{Files: files}, err
		}
		files = append(files, params.Directory+"/go.sum")
	}
	if err := GoRunner(workspaceRoot, []string{"go", "work", "use", "./" + params.Directory}); err != nil {
		return protocol.GenerateResult{}, err
	}
	return protocol.GenerateResult{Files: files}, nil
}

// WorkspaceGoVersion is the version of the workspace's go.work `go` line
// ("1.26.0"), or "" when there is no go.work or it names none.
func WorkspaceGoVersion(workspaceRoot string) string {
	src, err := os.ReadFile(filepath.Join(workspaceRoot, "go.work"))
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(src), "\n") {
		if fields := strings.Fields(line); len(fields) == 2 && fields[0] == "go" {
			return fields[1]
		}
	}
	return ""
}

// declareKind records the kind, and the stack when there is one, as the
// first lines of go.mod: comments the Go toolchain keeps.
func declareKind(goMod string, kind protocol.ProjectKind, stack string) error {
	src, err := os.ReadFile(goMod)
	if err != nil {
		return err
	}
	head := "// smith:kind " + string(kind) + "\n"
	if stack != "" {
		head += "// smith:stack " + stack + "\n"
	}
	return os.WriteFile(goMod, append([]byte(head), src...), 0o644)
}

// PackageName turns a project name into a valid Go package name.
func PackageName(name string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(name) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	out := b.String()
	if out == "" || unicode.IsDigit(rune(out[0])) {
		out = "pkg" + out
	}
	return out
}
