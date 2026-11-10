package adoptsplit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
)

// projectOnlyOptions are the compilerOptions that name folders or files of
// one project, so they stay in the project's tsconfig. The others (target,
// module, strict, ...) are the same for every project and go to the base.
var projectOnlyOptions = map[string]bool{
	"outDir": true, "rootDir": true, "rootDirs": true, "baseUrl": true, "paths": true,
	"tsBuildInfoFile": true, "declarationDir": true, "outFile": true,
	"typeRoots": true, "types": true, "composite": true, "plugins": true,
}

// shareTSConfig moves the compilerOptions that every project shares from the
// project's tsconfig.json into a tsconfig.base.json at the root, and makes the
// project's file extend it. It leaves the project alone, and says why, when
// there is no tsconfig.json, it already extends another, it has comments or
// commas that cannot be read, or the root already has a tsconfig.base.json.
// It returns the note for the person, or "".
func shareTSConfig(req Request, dir string) (string, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "tsconfig.json"))
	if err != nil {
		return "", nil
	}
	basePath := filepath.Join(req.Root, "tsconfig.base.json")
	if _, err := os.Stat(basePath); err == nil {
		return "tsconfig.json stays whole: the root already has a tsconfig.base.json.", nil
	}
	clean := jsondoc.StripComments(raw)
	doc, err := jsondoc.ParseDocument(clean)
	if err != nil || !json.Valid(clean) {
		return "tsconfig.json stays whole: smith cannot read it as JSON (a trailing comma?).", nil
	}
	if doc.Has("extends") {
		return "tsconfig.json stays whole: it already extends another file.", nil
	}
	options := doc.Object("compilerOptions", false)
	if options == nil {
		return "", nil
	}
	base, _ := jsondoc.ParseDocument([]byte("{}"))
	shared := base.Object("compilerOptions", true)
	var values map[string]json.RawMessage
	if err := json.Unmarshal(options.Bytes(), &values); err != nil {
		return "", nil
	}
	for _, key := range options.Keys() {
		if projectOnlyOptions[key] {
			continue
		}
		if err := shared.SetJSON(key, values[key]); err != nil {
			return "", nil
		}
		options.Delete(key)
	}
	if shared.Len() == 0 {
		return "", nil
	}

	var all map[string]json.RawMessage
	if err := json.Unmarshal(doc.Bytes(), &all); err != nil {
		return "", nil
	}
	project, _ := jsondoc.ParseDocument([]byte("{}"))
	project.SetString("extends", strings.Repeat("../", strings.Count(req.Dir, "/")+1)+"tsconfig.base.json")
	for _, key := range doc.Keys() {
		if key == "compilerOptions" && options.Len() == 0 {
			continue
		}
		if err := project.SetJSON(key, all[key]); err != nil {
			return "", nil
		}
	}
	if err := os.WriteFile(basePath, base.Bytes(), 0o644); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(dir, "tsconfig.json"), project.Bytes(), 0o644); err != nil {
		return "", err
	}
	return "the compiler options every project shares are in the root tsconfig.base.json; " + req.Dir + "/tsconfig.json extends it and keeps outDir, rootDir, include and the other paths.", nil
}
