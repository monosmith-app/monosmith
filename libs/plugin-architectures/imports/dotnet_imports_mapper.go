package imports

import (
	"regexp"
	"strings"
)

var (
	dotnetNamespace = regexp.MustCompile(`(?m)^[ \t]*namespace[ \t]+([\w.]+)`)
	dotnetUsing     = regexp.MustCompile(`(?m)^[ \t]*(?:global[ \t]+)?using[ \t]+(?:static[ \t]+)?(?:\w+[ \t]*=[ \t]*)?([\w.]+)[ \t]*;`)
)

var dotnetLanguage = Language{
	Name:       "dotnet",
	Extensions: []string{".cs"},
	Inputs:     []string{"**/*.cs"},
	IsTest: func(name string) bool {
		return strings.HasSuffix(name, "Tests.cs") || strings.HasSuffix(name, "Test.cs")
	},
	Extract: func(_ Project, _, text string) []Ref {
		return namespacedImports(text, dotnetNamespace, dotnetUsing, ".")
	},
}
