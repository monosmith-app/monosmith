package doctor

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
)

// TrackedGenerated lists tracked files a task is likely to rewrite, so the
// working tree is dirty after a run and the release refuses it:
//   - files .gitignore ignores but git still tracks (committed before the
//     ignore rule: a .tsbuildinfo, an old cache), from
//     `git ls-files --cached --ignored --exclude-standard`
//   - tracked files inside a task's declared outputs
//
// It reads no run: doctor cannot tell a task's change from the developer's
// own edit, but these are the causes either way.
func TrackedGenerated(ignored, tracked []string, projects []*projectgraph.Project) []string {
	seen := map[string]bool{}
	var out []string
	for _, f := range ignored {
		if f != "" && !seen[f] {
			seen[f] = true
			out = append(out, f+" (tracked, but .gitignore ignores it)")
		}
	}
	for _, p := range projects {
		for _, name := range sortedTaskNames(p) {
			for _, o := range p.Tasks[name].Outputs {
				dir := path.Clean(path.Join(p.Root, o))
				if dir == "." || strings.HasPrefix(dir, "..") {
					continue
				}
				for _, f := range tracked {
					if !seen[f] && (f == dir || strings.HasPrefix(f, dir+"/")) {
						seen[f] = true
						out = append(out, fmt.Sprintf("%s (inside %s:%s's outputs)", f, p.Name, name))
					}
				}
			}
		}
	}
	sort.Strings(out)
	return out
}

func sortedTaskNames(p *projectgraph.Project) []string {
	names := make([]string, 0, len(p.Tasks))
	for n := range p.Tasks {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}
