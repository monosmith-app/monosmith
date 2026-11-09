package convert

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/convertplan"
)

// maxDeletedListed is how many deleted paths the confirmation names.
const maxDeletedListed = 12

// SwitchesToNpm reports whether applying the plan moves a pnpm or yarn
// workspace to npm: it deletes their lockfile or workspace file.
func SwitchesToNpm(plan *convertplan.Plan) bool {
	for _, f := range plan.Delete {
		switch f {
		case "pnpm-lock.yaml", "pnpm-workspace.yaml", "yarn.lock", ".yarnrc.yml", ".yarnrc":
			return true
		}
	}
	return false
}

// ConfirmationText is what the person reads before the conversion applies:
// the files it deletes (by name, the first few), that package.json files are
// rewritten, that a pnpm or yarn workspace becomes an npm workspace, and the
// ways to look first or stop. Nothing is hidden behind "the old tool's files".
//
// Parameters: plan is the conversion about to apply.
func ConfirmationText(plan *convertplan.Plan) string {
	var b strings.Builder
	b.WriteString("\nApplying this conversion will:\n")
	if n := len(plan.Delete); n > 0 {
		fmt.Fprintf(&b, "  - delete %d file(s) or folder(s) of the old tooling:\n", n)
		for i, f := range plan.Delete {
			if i == maxDeletedListed {
				fmt.Fprintf(&b, "      and %d more\n", n-i)
				break
			}
			fmt.Fprintf(&b, "      %s\n", f)
		}
	}
	if len(plan.Specifiers)+len(plan.RemoveDependencies)+len(plan.RootScripts) > 0 || plan.SetWorkspaces != nil {
		b.WriteString("  - rewrite the root package.json and the package.json of projects (scripts, dependencies, workspaces)\n")
	}
	if SwitchesToNpm(plan) {
		b.WriteString("  - SWITCH this pnpm or yarn workspace to npm: the pnpm or yarn lockfile and settings go, and workspace: and catalog: ranges are rewritten\n")
	}
	b.WriteString("  - regenerate the CI pipelines\n")
	b.WriteString("Git can undo it, since the tree is committed.\n")
	return b.String()
}
