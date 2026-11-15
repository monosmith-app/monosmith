package runhistory

import (
	"strings"

	"github.com/monosmith-app/monosmith/libs/actions/flagvocabulary"
)

// RerunFailed is the command line that runs again the tasks whose ids
// ("project:task") are given: the ones that failed, were interrupted or never
// started. It is nil when there are none. Tasks of different names on different
// projects run as a grid of both, the ones that already passed coming back from
// the cache.
func RerunFailed(ids []string) []string {
	var tasks, projects []string
	add := func(list []string, v string) []string {
		for _, have := range list {
			if have == v {
				return list
			}
		}
		return append(list, v)
	}
	for _, id := range ids {
		project, task, ok := strings.Cut(id, ":")
		if !ok {
			continue
		}
		projects, tasks = add(projects, project), add(tasks, task)
	}
	if len(tasks) == 0 {
		return nil
	}
	return append(append([]string{"run"}, tasks...), flagvocabulary.Arg("project"), strings.Join(projects, ","))
}
