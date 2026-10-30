package declaredimports

import (
	"go/ast"
	"strings"
)

// ignored reports whether a file's comments hold the `//go:build ignore`
// constraint: a generator or example run with go run, never built with the
// module, whose imports the module need not require.
func ignored(comments []*ast.CommentGroup) bool {
	for _, group := range comments {
		for _, c := range group.List {
			if strings.TrimSpace(c.Text) == "//go:build ignore" {
				return true
			}
		}
	}
	return false
}
