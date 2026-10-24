package buildoutputs

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
)

// FindGaps compares, for each Node project whose build task runs its
// package.json build script, what that script writes with the outputs the
// task declares.
func FindGaps(workspaceRoot string, projects []*projectgraph.Project) []Gap {
	var gaps []Gap
	for _, p := range projects {
		task, ok := p.Tasks[Task]
		if !ok || p.Plugin != "node" || !runsScript(task.Command) {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(p.Root), "package.json"))
		if err != nil {
			continue
		}
		var pkg struct {
			Scripts map[string]string `json:"scripts"`
		}
		if json.Unmarshal(raw, &pkg) != nil || pkg.Scripts[Task] == "" {
			continue
		}
		fromConfig := strings.HasPrefix(p.Sources[Task+".outputs"], "smith.toml")
		if gap, found := GapOf(p.Name, p.Root, pkg.Scripts[Task], task.Outputs, fromConfig); found {
			gaps = append(gaps, gap)
		}
	}
	return gaps
}

// runsScript reports whether a command is `npm run build`, as the Node
// plugin's task is, and not a command smith.toml replaced it with.
func runsScript(command []string) bool {
	return len(command) >= 3 && command[0] == "npm" && command[1] == "run" && command[2] == Task
}
