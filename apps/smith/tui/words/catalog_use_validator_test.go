package words

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"
	"unicode"
)

// migrated are the packages under tui that hold no wording of their own: each
// asks the catalog. A package joins this list when it moves, and the list ends
// as every package. A literal that is not wording (a key name, a protocol word)
// carries the comment // words:literal on its line.
var migrated = slices.Concat([]string{"doctorview"}, migratedByA, migratedByB)

var placeholder = regexp.MustCompile(`\{([a-z][A-Za-z0-9]*)\}`)

// parsed is one non-test Go file under tui, with the set it was parsed in.
type parsed struct {
	path string
	fset *token.FileSet
	file *ast.File
}

// sources parses the non-test Go files of every package under tui.
func sources(t *testing.T) []parsed {
	t.Helper()
	var out []parsed
	err := filepath.WalkDir("..", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "_test.go") {
			return err
		}
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, p, nil, parser.ParseComments)
		if err != nil {
			return err
		}
		out = append(out, parsed{filepath.ToSlash(p), fset, f})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// literal is the value of a string literal expression.
func literal(e ast.Expr) (string, bool) {
	lit, ok := e.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", false
	}
	s, err := strconv.Unquote(lit.Value)
	return s, err == nil
}

// asksTheCatalog is "Say" or "Count" when call is words.Say or words.Count.
func asksTheCatalog(call *ast.CallExpr) (string, bool) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return "", false
	}
	pkg, ok := sel.X.(*ast.Ident)
	if !ok || pkg.Name != "words" || (sel.Sel.Name != "Say" && sel.Sel.Name != "Count") {
		return "", false
	}
	return sel.Sel.Name, true
}

func TestEveryKeyTheCodeAsksForIsInTheCatalogAndItsPlaceholdersAreFilled(t *testing.T) {
	en := catalogs[referenceLocale]
	for _, src := range sources(t) {
		ast.Inspect(src.file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			fn, ok := asksTheCatalog(call)
			if !ok || len(call.Args) == 0 {
				return true
			}
			key, ok := literal(call.Args[0])
			if !ok {
				return true // a key held in a variable is covered by TestNoKeyIsUnused
			}
			texts, given, rest := []string{en[key]}, map[string]bool{}, call.Args[1:]
			if fn == "Count" {
				texts, rest = []string{en[key+".one"], en[key+".other"]}, rest[min(1, len(rest)):]
				given["count"] = true
			}
			for i := 0; i < len(rest); i += 2 {
				if name, ok := literal(rest[i]); ok {
					given[name] = true
				}
			}
			for _, text := range texts {
				if text == "" {
					t.Errorf("%s: %s asks for %q, which the catalog lacks", src.path, fn, key)
					continue
				}
				for _, m := range placeholder.FindAllStringSubmatch(text, -1) {
					if !given[m[1]] {
						t.Errorf("%s: %q needs {%s}, which the call does not give", src.path, key, m[1])
					}
				}
			}
			return true
		})
	}
}

func TestNoKeyIsUnused(t *testing.T) {
	seen := map[string]bool{}
	for _, src := range sources(t) {
		if strings.Contains(src.path, "/words/") {
			continue
		}
		ast.Inspect(src.file, func(n ast.Node) bool {
			if e, ok := n.(ast.Expr); ok {
				if s, ok := literal(e); ok {
					seen[s] = true
				}
			}
			return true
		})
	}
	var unused []string
	for key := range catalogs[referenceLocale] {
		base := strings.TrimSuffix(strings.TrimSuffix(key, ".one"), ".other")
		if !seen[key] && !seen[base] {
			unused = append(unused, key)
		}
	}
	sort.Strings(unused)
	if len(unused) > 0 {
		t.Errorf("catalog keys no code asks for:\n  %s", strings.Join(unused, "\n  "))
	}
}

func TestMigratedPackagesHoldNoWordingOfTheirOwn(t *testing.T) {
	for _, src := range sources(t) {
		if !slices.Contains(migrated, strings.Split(strings.TrimPrefix(src.path, "../"), "/")[0]) {
			continue
		}
		exempt := map[int]bool{}
		for _, group := range src.file.Comments {
			for _, c := range group.List {
				if strings.Contains(c.Text, "words:literal") {
					exempt[src.fset.Position(c.Pos()).Line] = true
				}
			}
		}
		skip := map[*ast.BasicLit]bool{}
		ast.Inspect(src.file, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.ImportSpec:
				skip[x.Path] = true
			case *ast.Field:
				if x.Tag != nil {
					skip[x.Tag] = true
				}
			}
			return true
		})
		ast.Inspect(src.file, func(n ast.Node) bool {
			lit, ok := n.(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING || skip[lit] {
				return true
			}
			line := src.fset.Position(lit.Pos()).Line
			if s, err := strconv.Unquote(lit.Value); err == nil && isWording(s) && !exempt[line] {
				t.Errorf("%s:%d: wording in code, move it to the catalog: %q", src.path, line, s)
			}
			return true
		})
	}
}

// isWording reports whether a literal reads as something a person is shown: it
// has a letter and either a space or a capital first letter.
func isWording(s string) bool {
	hasLetter := strings.IndexFunc(s, unicode.IsLetter) >= 0
	first := []rune(s + " ")[0]
	return hasLetter && (strings.ContainsRune(s, ' ') || unicode.IsUpper(first))
}
