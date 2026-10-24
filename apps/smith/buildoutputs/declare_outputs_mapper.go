package buildoutputs

import (
	"regexp"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/configedit"
)

var bareKey = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// tableKey quotes a project name that is not a bare TOML key (ms.sharepoint).
func tableKey(name string) string {
	if bareKey.MatchString(name) {
		return name
	}
	return configedit.Quote(name)
}

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
		src = configedit.SetKey(src, "projects."+tableKey(g.Project)+".tasks."+Task, "outputs", "["+quoteAll(g.Outputs)+"]")
	}
	return src
}
