package pipeline

import (
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

var modulePathLine = regexp.MustCompile(`(?m)^module\s+"?([^\s"]+)"?\s*$`)

// CheckUncoveredImports reports the packages of a module that a project
// declared in smith.toml imports and no declared or discovered project holds:
// a shared internal/ folder, say. Nothing rebuilds the project when such a
// folder changes, because it is part of no cache key. One warning per project,
// naming the folders to declare as libs.
func CheckUncoveredImports(workspaceRoot string, projects []protocol.Project) []protocol.Finding {
	var findings []protocol.Finding
	for _, p := range projects {
		if !declaredInModule(workspaceRoot, p.Root) {
			continue
		}
		src, err := os.ReadFile(ModuleFile(workspaceRoot, p.Root))
		if err != nil {
			continue
		}
		m := modulePathLine.FindStringSubmatch(string(src))
		if m == nil {
			continue
		}
		module := m[1]
		uncovered := map[string]bool{}
		for _, imp := range p.Imports {
			if imp != module && !strings.HasPrefix(imp, module+"/") {
				continue
			}
			if !covered(imp, projects) {
				uncovered[strings.TrimPrefix(strings.TrimPrefix(imp, module), "/")] = true
			}
		}
		if len(uncovered) == 0 {
			continue
		}
		folders := make([]string, 0, len(uncovered))
		for f := range uncovered {
			if f == "" {
				f = "(the module's root package)"
			}
			folders = append(folders, f)
		}
		sort.Strings(folders)
		findings = append(findings, protocol.Finding{
			Check:   p.Name + " imports only packages that a project holds",
			Warning: true,
			Detail:  p.Root + " imports " + strings.Join(folders, ", ") + " of its module, which no project holds: a change there does not rebuild it, because the folder is part of no cache key",
			Remedy:  "declare each folder as a lib: [projects.<name>] with root = \"<folder>\", language = \"go\", kind = \"lib\"",
		})
	}
	return findings
}

// covered says whether an import path is, or is below, a package a project exports.
func covered(imp string, projects []protocol.Project) bool {
	for _, p := range projects {
		for _, e := range p.Exports {
			if imp == e || strings.HasPrefix(imp, e+"/") {
				return true
			}
		}
	}
	return false
}
