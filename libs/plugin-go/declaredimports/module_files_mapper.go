package declaredimports

import (
	"strings"
)

// modulePaths are the module paths a go.mod or go.work file names.
type modulePaths struct {
	// module is the go.mod's own module path.
	module string
	// required are the modules of its require directives.
	required []string
	// replaced are the modules a replace directive redirects.
	replaced []string
	// used are the directories of a go.work's use directives.
	used []string
}

// parseModuleFile reads the directives of a go.mod or go.work: single-line
// and parenthesised blocks, with // comments removed.
func parseModuleFile(src string) modulePaths {
	var out modulePaths
	block := ""
	for _, line := range strings.Split(src, "\n") {
		line, _, _ = strings.Cut(line, "//")
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		directive := block
		if block == "" {
			directive = fields[0]
			fields = fields[1:]
			if len(fields) == 1 && fields[0] == "(" {
				block = directive
				continue
			}
		} else if fields[0] == ")" {
			block = ""
			continue
		}
		if len(fields) == 0 {
			continue
		}
		name := unquote(fields[0])
		switch directive {
		case "module":
			out.module = name
		case "require":
			out.required = append(out.required, name)
		case "replace":
			out.replaced = append(out.replaced, name)
		case "use":
			out.used = append(out.used, name)
		}
	}
	return out
}

func unquote(s string) string { return strings.Trim(s, "\"`") }
