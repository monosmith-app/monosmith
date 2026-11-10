package doctorchecks

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/recipedrift"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// RecipeChecks reports each project made from a recipe: whether it still
// follows the recipe as it is now. A workspace with no such project has none.
func RecipeChecks(s *session.Session) []protocol.Finding {
	var out []protocol.Finding
	for _, r := range recipedrift.Check(s) {
		out = append(out, RecipeFinding(r))
	}
	return out
}

// RecipeFinding is a drift report as a finding: a project that has drifted,
// or that cannot be checked, is a warning that names its remedy.
func RecipeFinding(r recipedrift.Report) protocol.Finding {
	f := protocol.Finding{Check: r.Project + " follows its recipe " + r.Recipe, Project: r.Project}
	problem := strings.TrimPrefix(r.Summary(), r.Project+" ")
	switch {
	case r.Problem != "":
		f.Warning = true
		f.Problem = problem
		f.Detail = r.Summary()
		f.Remedy = "fix what the detail names; `smith recipe check <recipe>` shows what the recipe resolves to, and a recipe at a URL is read from its pinned copy when the network is down"
	case r.Drifted():
		f.Warning = true
		f.Problem = problem
		f.Detail = r.Summary()
		f.Remedy = "run `smith upgrade` (`smith upgrade --dry-run` shows what it changes first): it rewrites the files smith owns and leaves the ones you edit"
	default:
		f.OK = true
	}
	return f
}
