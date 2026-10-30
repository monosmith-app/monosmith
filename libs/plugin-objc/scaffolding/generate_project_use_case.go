package scaffolding

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/templating"
)

type templateData struct {
	Dir, Name, Module, Kind, Stack, Workspace string
}

// Generate writes a CocoaPods library: a lib used inside the workspace or a
// package others depend on by git tag. An app is refused: a podspec describes
// a library.
func Generate(workspaceRoot string, params protocol.GenerateParams) (protocol.GenerateResult, error) {
	if params.Kind == protocol.ProjectApp {
		return protocol.GenerateResult{}, errors.New("an Objective-C project is a CocoaPods library (a lib or a package), not an app")
	}
	if _, err := os.Stat(filepath.Join(workspaceRoot, filepath.FromSlash(params.Directory))); err == nil {
		return protocol.GenerateResult{}, fmt.Errorf("%s already exists", params.Directory)
	}
	data := templateData{Dir: params.Directory, Name: params.Name, Module: moduleName(params.Name), Kind: string(params.Kind), Stack: params.Stack, Workspace: params.Workspace.Name}
	written, err := templating.Write(workspaceRoot, templates, data)
	if err != nil {
		return protocol.GenerateResult{}, err
	}
	note := "pod lib lint needs a Mac with Xcode and CocoaPods; replace the placeholder homepage, source and author in " + data.Module + ".podspec before publishing the pod."
	return protocol.GenerateResult{Files: written, Notes: []string{note}}, nil
}

// moduleName is the pod and class name of a project: its name in
// UpperCamelCase, with every character that is not a letter or digit dropped
// and a leading digit given an M.
func moduleName(name string) string {
	var b strings.Builder
	upper := true
	for _, c := range name {
		if !unicode.IsLetter(c) && !unicode.IsDigit(c) {
			upper = true
			continue
		}
		if upper {
			c = unicode.ToUpper(c)
			upper = false
		}
		b.WriteRune(c)
	}
	if b.Len() == 0 || unicode.IsDigit([]rune(b.String())[0]) {
		return "M" + b.String()
	}
	return b.String()
}
