package doctor

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
)

// SharedOutputs lists the pairs of tasks whose outputs overlap: one directory
// is the other, or holds it. A cache hit replaces a task's output directories
// wholesale, so restoring one task deletes what the other wrote, and two
// package tasks that both declare dist/drop fail each other mid-write. A
// directory in allowed ([workspace] shared_outputs) may be written by several
// tasks, each into its own files: overlap inside it is not reported.
func SharedOutputs(projects []*projectgraph.Project, allowed ...string) []string {
	type output struct{ task, dir string }
	var all []output
	for _, p := range projects {
		for n, t := range p.Tasks {
			for _, o := range t.Outputs {
				all = append(all, output{p.Name + ":" + n, path.Clean(path.Join(p.Root, o))})
			}
		}
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i].task != all[j].task {
			return all[i].task < all[j].task
		}
		return all[i].dir < all[j].dir
	})
	var out []string
	seen := map[string]bool{}
	for i, a := range all {
		for _, b := range all[i+1:] {
			overlap := a.dir == b.dir || strings.HasPrefix(b.dir, a.dir+"/") || strings.HasPrefix(a.dir, b.dir+"/")
			shorter := a.dir
			if len(b.dir) < len(shorter) {
				shorter = b.dir
			}
			if a.task == b.task || !overlap || InsideAny(shorter, allowed) {
				continue
			}
			where := a.dir
			if a.dir != b.dir {
				where = fmt.Sprintf("%s and %s", a.dir, b.dir)
			}
			line := fmt.Sprintf("%s and %s both write %s", a.task, b.task, where)
			if !seen[line] {
				seen[line] = true
				out = append(out, line)
			}
		}
	}
	return out
}
