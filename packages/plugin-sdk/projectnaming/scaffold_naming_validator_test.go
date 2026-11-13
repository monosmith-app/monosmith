package projectnaming

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// namingPackage is the import path every scaffold must use for a name.
const namingPackage = `"github.com/monosmith-app/monosmith/packages/plugin-sdk/projectnaming"`

// TestEveryScaffoldBuildsItsNamesThroughProjectnaming is the conformance rule
// of ADR 0098: a language plugin whose scaffolding builds a project name must
// do it through this package. It fails for a plugin that builds no name through it,
// or that defines a private name sanitizer.
func TestEveryScaffoldBuildsItsNamesThroughProjectnaming(t *testing.T) {
	scaffolds, err := filepath.Glob(filepath.Join("..", "..", "..", "libs", "plugin-*", "scaffolding"))
	if err != nil || len(scaffolds) == 0 {
		t.Fatalf("no plugin scaffolding found: %v", err)
	}
	// A plugin that makes no package name has nothing to build.
	exempt := map[string]string{"plugin-bundle": "writes no package", "plugin-recipe": "writes no package", "plugin-applescript": "a script has no package name", "plugin-docker": "writes a Dockerfile, no package", "plugin-bicep": "an infrastructure project has no package name"}
	private := map[string]bool{"identifier": true, "kebab": true, "pascal": true, "snake": true, "moduleName": true, "distributionIdentifier": true, "crateIdentifier": true, "packageName": true}
	for _, dir := range scaffolds {
		plugin := filepath.Base(filepath.Dir(dir))
		if _, skip := exempt[plugin]; skip {
			continue
		}
		files, _ := filepath.Glob(filepath.Join(dir, "*.go"))
		uses, generates := false, false
		for _, file := range files {
			if strings.HasSuffix(file, "_test.go") {
				continue
			}
			src, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			text := string(src)
			if strings.Contains(text, namingPackage) {
				uses = true
			}
			if strings.Contains(text, "protocol.GenerateParams") {
				generates = true
			}
			tree, err := parser.ParseFile(token.NewFileSet(), file, src, 0)
			if err != nil {
				t.Fatal(err)
			}
			for _, decl := range tree.Decls {
				if fn, ok := decl.(*ast.FuncDecl); ok && fn.Recv == nil && private[fn.Name.Name] {
					t.Errorf("%s: private sanitizer %s: use projectnaming", file, fn.Name.Name)
				}
			}
		}
		if generates && !uses {
			t.Errorf("%s: the scaffolding builds no name through projectnaming", plugin)
		}
	}
}
