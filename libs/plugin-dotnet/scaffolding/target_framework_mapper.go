package scaffolding

import (
	"regexp"
	"sort"
	"strings"
)

var (
	frameworkElement  = regexp.MustCompile(`<TargetFramework>\s*([^<\s$]+)\s*</TargetFramework>`)
	frameworkLine     = regexp.MustCompile(`(?m)^[ \t]*<TargetFramework>[^<]*</TargetFramework>[ \t]*\r?\n`)
	firstPropertyOpen = regexp.MustCompile(`<PropertyGroup>[ \t]*\r?\n`)
)

// Framework is the target framework a file sets with a TargetFramework
// element (a project, or Directory.Build.props). It is "" when the file sets
// none, sets several with TargetFrameworks, or names a property ($(Name)),
// which only its owner can move.
func Framework(src string) string {
	if m := frameworkElement.FindStringSubmatch(src); m != nil {
		return m[1]
	}
	return ""
}

// ShareFramework moves the target framework the projects repeat into
// Directory.Build.props, the way new projects inherit it. The shared framework
// is the one the props file already sets, else the one most projects set (the
// lowest on a tie). A project with that framework loses its own element; one
// with another framework keeps it. projects maps a project file path to its
// text; propsSrc is "" when there is no Directory.Build.props. It returns the
// new props text, the new text of each project it changed, and the shared
// framework; the framework is "" when nothing needs to change.
func ShareFramework(propsSrc string, projects map[string]string) (props string, changed map[string]string, framework string) {
	framework = Framework(propsSrc)
	count := map[string]int{}
	for _, src := range projects {
		if f := Framework(src); f != "" {
			count[f]++
		}
	}
	if len(count) == 0 {
		return propsSrc, nil, ""
	}
	if framework == "" {
		var seen []string
		for f := range count {
			seen = append(seen, f)
		}
		sort.Strings(seen)
		framework = seen[0]
		for _, f := range seen {
			if count[f] > count[framework] {
				framework = f
			}
		}
	}
	changed = map[string]string{}
	for path, src := range projects {
		if Framework(src) == framework {
			changed[path] = frameworkLine.ReplaceAllString(src, "")
		}
	}
	return withFramework(propsSrc, framework), changed, framework
}

// withFramework makes Directory.Build.props set the framework: it adds the
// element to the first property group, or the group to a file that has none.
func withFramework(propsSrc, framework string) string {
	if Framework(propsSrc) != "" {
		return propsSrc
	}
	element := "    <TargetFramework>" + framework + "</TargetFramework>\n"
	if strings.TrimSpace(propsSrc) == "" {
		return "<Project>\n  <PropertyGroup>\n" + element + "  </PropertyGroup>\n</Project>\n"
	}
	if loc := firstPropertyOpen.FindStringIndex(propsSrc); loc != nil {
		return propsSrc[:loc[1]] + element + propsSrc[loc[1]:]
	}
	if i := strings.LastIndex(propsSrc, "</Project>"); i >= 0 {
		return propsSrc[:i] + "  <PropertyGroup>\n" + element + "  </PropertyGroup>\n" + propsSrc[i:]
	}
	return propsSrc
}
