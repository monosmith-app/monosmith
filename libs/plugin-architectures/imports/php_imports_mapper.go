package imports

import (
	"regexp"
	"strings"
)

var (
	phpNamespace = regexp.MustCompile(`(?m)^[ \t]*namespace[ \t]+([\w\\]+)`)
	phpUse       = regexp.MustCompile(`(?m)^[ \t]*use[ \t]+(?:function[ \t]+|const[ \t]+)?\\?([\w\\]+)(?:[ \t]+as[ \t]+\w+)?[ \t]*;`)
)

var phpLanguage = Language{
	Name:       "php",
	Extensions: []string{".php"},
	Inputs:     []string{"**/*.php"},
	IsTest:     func(name string) bool { return strings.HasSuffix(name, "Test.php") },
	Extract: func(_ Project, _, text string) []Ref {
		return namespacedImports(text, phpNamespace, phpUse, `\`)
	},
}
