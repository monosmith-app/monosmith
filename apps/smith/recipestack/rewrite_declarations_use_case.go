package recipestack

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// Rewrite brings the declarations a recipe wrote into smith.toml text up to
// date. It removes the recipe record and settings of each declaration's
// project, and the stacks named in replace, then writes each declaration
// again at the end, as `smith add` did: everything else in the text, the
// hand-written tables and comments included, stays exactly as it was. A
// declaration whose Reuse is false writes its stack, so the stack named in
// replace comes back with the first of them. The result must still parse.
func Rewrite(src string, decls []Declaration, replace []string) (string, error) {
	projects := make([]string, 0, len(decls))
	for _, d := range decls {
		projects = append(projects, d.Project)
	}
	text, _ := workspace.RemoveRecipeTables(src, projects, replace)
	if text != "" && !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	for _, d := range decls {
		text += "\n" + Render(d)
	}
	if _, err := workspace.Parse([]byte(text)); err != nil {
		return "", fmt.Errorf("smith.toml with the recipes brought up to date would not load: %w", err)
	}
	return text, nil
}
