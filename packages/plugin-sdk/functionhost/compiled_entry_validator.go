package functionhost

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CheckCompiledEntry makes sure a staged function app holds compiled code, so
// a build whose output went to another folder fails here and not at deploy.
// With a "main" in the manifest, the file (or glob, as the Node v4 model
// uses) must match a file in the stage. Without one, the stage must hold a
// dist folder with at least one .js file.
//
// Parameters: stage is the folder that becomes the zip; manifest is its
// package.json; buildDir is the build folder, named in the message.
// Returns an error that names the missing entry and the likely cause, or nil.
func CheckCompiledEntry(stage string, manifest map[string]any, buildDir string) error {
	main, _ := manifest["main"].(string)
	main = strings.TrimPrefix(strings.TrimSpace(main), "./")
	if main != "" {
		matches, err := filepath.Glob(filepath.Join(stage, filepath.FromSlash(main)))
		if err != nil {
			// A pattern Go cannot read (such as braces) is left to the host.
			return nil
		}
		if len(matches) > 0 {
			return nil
		}
		return fmt.Errorf("the function app has no compiled code: \"main\" in package.json is %s, but %s has no such file. Check that the build (the tsconfig outDir) writes into %s, then package again", main, buildDir, filepath.Join(buildDir, "dist"))
	}
	found := false
	_ = filepath.WalkDir(filepath.Join(stage, "dist"), func(p string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(p, ".js") {
			found = true
			return filepath.SkipAll
		}
		return nil
	})
	if !found {
		return fmt.Errorf("the function app has no compiled code: %s has no .js file. Check that the build (the tsconfig outDir) writes into it, then package again", filepath.Join(buildDir, "dist"))
	}
	return nil
}
