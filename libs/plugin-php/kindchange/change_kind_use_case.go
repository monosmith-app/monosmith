package kindchange

import (
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-php/scaffolding"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/fileedits"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/pathrefs"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

const manifestName = "composer.json"

var (
	smithKind      = regexp.MustCompile(`("smith"[ \t\r\n]*:[ \t\r\n]*\{[^{}]*?"kind"[ \t]*:[ \t]*)"[^"]*"`)
	typeField      = regexp.MustCompile(`("type"[ \t]*:[ \t]*)"(?:project|library)"`)
	relativeString = regexp.MustCompile(`"(\.{1,2}/[^"]*)"`)
	autoloadPath   = regexp.MustCompile(`'/(?:\.\./)+vendor/autoload\.php'`)
)

// Change answers what making the package params.Project a params.To changes:
// the kind in its composer.json (extra.smith.kind, and the type, a project for
// an app and a library for the rest), and the root composer.json's path
// repository, which names its directory.
func Change(workspaceRoot string, params protocol.ChangeKindParams) (protocol.ChangeKindResult, error) {
	p := params.Project
	edits := fileedits.New(workspaceRoot)
	manifest := path.Join(p.Root, manifestName)
	src, ok := edits.Read(manifest)
	if !ok {
		return protocol.ChangeKindResult{}, fmt.Errorf("%s not found", manifest)
	}
	var result protocol.ChangeKindResult

	src, err := setKind(src, string(params.To))
	if err != nil {
		return result, fmt.Errorf("%s: %w", manifest, err)
	}
	if typeField.MatchString(src) {
		kind := "library"
		if params.To == protocol.ProjectApp {
			kind = "project"
		}
		src = typeField.ReplaceAllString(src, `${1}"`+kind+`"`)
	}
	doc, err := jsondoc.ParseDocument([]byte(src))
	if err != nil {
		return result, fmt.Errorf("%s: %w", manifest, err)
	}
	if params.To == protocol.ProjectPackage {
		for _, field := range []string{"description", "license"} {
			if !doc.Has(field) {
				result.Warnings = append(result.Warnings, fmt.Sprintf("%s has no %q: packagist.org asks for it before a package is listed", manifest, field))
			}
		}
	}
	moving := params.NewRoot != p.Root
	if moving {
		src = relativeString.ReplaceAllStringFunc(src, func(m string) string {
			return `"` + pathrefs.Rebase(relativeString.FindStringSubmatch(m)[1], p.Root, params.NewRoot) + `"`
		})
	}
	edits.Write(manifest, src)
	result.Warnings = append(result.Warnings, programWarnings(doc, manifest, params.To)...)

	if moving {
		fixAutoloadPaths(edits, params)
		if err := retargetWorkspace(edits, params); err != nil {
			return result, err
		}
		result.Warnings = append(result.Warnings, "run smith install afterwards: vendor/ links this package from its old directory")
	}
	result.Files = edits.Edits()
	return result, nil
}

// setKind records the kind in extra.smith, in place when the key is there.
func setKind(src, kind string) (string, error) {
	if smithKind.MatchString(src) {
		return smithKind.ReplaceAllString(src, `${1}"`+kind+`"`), nil
	}
	doc, err := jsondoc.ParseDocument([]byte(src))
	if err != nil {
		return src, err
	}
	doc.Object("extra", true).Object("smith", true).SetString("kind", kind)
	return string(doc.Bytes()), nil
}

// programWarnings say what the kind change makes of the package's executables.
func programWarnings(doc *jsondoc.Document, manifest string, to protocol.ProjectKind) []string {
	switch {
	case to == protocol.ProjectApp && !doc.Has("bin"):
		return []string{manifest + ` has no "bin": an app runs a program, add one before running it`}
	case to != protocol.ProjectApp && doc.Has("bin"):
		return []string{fmt.Sprintf("this package has executables (bin); as a %s it is required by other packages and the programs are kept but not deployed", to)}
	}
	return nil
}

// fixAutoloadPaths points the package's executables at the workspace's
// vendor/autoload.php from the new depth: the generated ones climb one folder
// per directory level of the package, plus its own bin/ and the vendor's.
func fixAutoloadPaths(edits *fileedits.Set, params protocol.ChangeKindParams) {
	ups := strings.Repeat("../", strings.Count(params.NewRoot, "/")+2)
	for _, file := range edits.List(path.Join(params.Project.Root, "bin"), func(string) bool { return true }) {
		text, _ := edits.Read(file)
		edits.Write(file, autoloadPath.ReplaceAllString(text, `'/`+ups+`vendor/autoload.php'`))
	}
}

// retargetWorkspace updates the root composer.json (the path repository, and
// its key, which is the directory dashed) and composer.lock, which records
// where each path package lives.
func retargetWorkspace(edits *fileedits.Set, params protocol.ChangeKindParams) error {
	oldKey, newKey := scaffolding.RepositoryKey(params.Project.Root), scaffolding.RepositoryKey(params.NewRoot)
	for _, file := range []string{manifestName, "composer.lock"} {
		text, ok := edits.Read(file)
		if !ok {
			continue
		}
		text, _ = pathrefs.Retarget(text, "", params.Project.Root, params.NewRoot)
		text, _ = pathrefs.ReplaceReference(text, `"`+oldKey+`"`, `"`+newKey+`"`)
		edits.Write(file, text)
	}
	return nil
}
