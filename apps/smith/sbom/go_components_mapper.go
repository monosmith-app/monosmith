package sbom

import (
	"bufio"
	"strings"
)

// GoComponents lists the modules a go.mod requires, direct and indirect.
func GoComponents(gomod string) []Component {
	var out []Component
	inBlock := false
	sc := bufio.NewScanner(strings.NewReader(gomod))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if i := strings.Index(line, "//"); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		switch {
		case line == "require (":
			inBlock = true
			continue
		case line == ")":
			inBlock = false
			continue
		case strings.HasPrefix(line, "require "):
			line = strings.TrimSpace(strings.TrimPrefix(line, "require "))
		case !inBlock:
			continue
		}
		f := strings.Fields(line)
		if len(f) != 2 {
			continue
		}
		out = append(out, Component{Name: f[0], Version: f[1], PURL: "pkg:golang/" + f[0] + "@" + f[1]})
	}
	return out
}

// GoModule is the module path a go.mod declares.
func GoModule(gomod string) string {
	for _, line := range strings.Split(gomod, "\n") {
		if rest, ok := strings.CutPrefix(strings.TrimSpace(line), "module "); ok {
			return strings.Trim(strings.TrimSpace(rest), `"`)
		}
	}
	return ""
}
