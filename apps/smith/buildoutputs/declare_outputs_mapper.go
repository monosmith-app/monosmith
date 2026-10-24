package buildoutputs

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/configedit"
)

func quoteAll(items []string) string {
	quoted := make([]string, len(items))
	for i, s := range items {
		quoted[i] = configedit.Quote(s)
	}
	return strings.Join(quoted, ", ")
}

// DeclareOutputs returns the smith.toml text with each gap's outputs set
// under [projects.<name>.tasks.build].
func DeclareOutputs(src string, gaps []Gap) string {
	for _, g := range gaps {
		src = configedit.SetKey(src, "projects."+configedit.TableKey(g.Project)+".tasks."+Task, "outputs", "["+quoteAll(g.Outputs)+"]")
	}
	return src
}
