package workspace

import (
	"fmt"
	"path"
	"sort"
	"strings"
)

// validateDeclaredProjects refuses a [projects.<name>] table that declares a
// project badly: a language or kind with no root, a root with no language, a
// root that is not a clean workspace-relative folder, an unknown kind, or two
// projects at one root. A project with a manifest of its own is declared by no
// table: the plugin finds it.
func validateDeclaredProjects(projects map[string]ProjectOverride) error {
	names := make([]string, 0, len(projects))
	for name := range projects {
		names = append(names, name)
	}
	sort.Strings(names)
	roots := map[string]string{}
	for _, name := range names {
		po := projects[name]
		at := fmt.Sprintf("%s: [projects.%s]", FileName, name)
		switch {
		case po.Root == "" && (po.Language != "" || po.Kind != ""):
			return fmt.Errorf("%s language and kind need root: the folder of the project the table declares", at)
		case po.Root == "":
			continue
		case po.Language == "":
			return fmt.Errorf("%s root needs language: the plugin that describes the project (\"go\", say)", at)
		case po.Root == "." || path.Clean(po.Root) != po.Root || strings.HasPrefix(po.Root, "/") || strings.HasPrefix(po.Root, "../") || strings.Contains(po.Root, `\`):
			return fmt.Errorf("%s root must be a workspace-relative folder with forward slashes, such as apps/%s, not %q", at, name, po.Root)
		case po.Kind != "" && po.Kind != "app" && po.Kind != "lib" && po.Kind != "package":
			return fmt.Errorf("%s kind must be \"app\", \"lib\" or \"package\", not %q", at, po.Kind)
		}
		if other, ok := roots[po.Root]; ok {
			return fmt.Errorf("%s root %s is also the root of [projects.%s]", at, po.Root, other)
		}
		roots[po.Root] = name
	}
	return nil
}
