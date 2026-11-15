package wordsguard

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/words/words"
)

// A plural is two keys the code picks between (key.one and key.other), so a
// words.Count call anywhere in the app, not only in the TUI, needs both forms:
// a missing one would show the key name for the count that lacks it.
func TestEveryCountCallInTheAppHasBothFormsInTheCatalog(t *testing.T) {
	en := words.Reference()
	err := filepath.WalkDir("../..", func(p string, d fs.DirEntry, err error) error {
		if err == nil && d.IsDir() && (d.Name() == "tui-research" || d.Name() == "testdata") {
			return fs.SkipDir
		}
		if err != nil || d.IsDir() || !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "_test.go") {
			return err
		}
		f, err := parser.ParseFile(token.NewFileSet(), p, nil, 0)
		if err != nil {
			return err
		}
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || len(call.Args) == 0 {
				return true
			}
			if fn, ok := asksTheCatalog(call); !ok || fn != "Count" {
				return true
			}
			key, ok := literal(call.Args[0])
			if !ok {
				return true
			}
			for _, form := range []string{".one", ".other"} {
				if en[key+form] == "" {
					t.Errorf("%s: words.Count(%q) lacks %q in the catalog", filepath.ToSlash(p), key, key+form)
				}
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
