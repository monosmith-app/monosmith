package whymiss

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/features/session"
	"github.com/monosmith-app/monosmith/apps/smith/features/taskcache"
	"github.com/monosmith-app/monosmith/apps/smith/features/taskgraph"
	"github.com/monosmith-app/monosmith/apps/smith/features/taskrun"
)

// Status is the outcome of comparing a task with its last cached run.
type Status string

// Statuses.
const (
	// Hit: the task's current key is cached; it would be restored.
	Hit Status = "hit"
	// Changed: the task would run; Differences says why.
	Changed Status = "changed"
	// NotCached: the task does not use the cache.
	NotCached Status = "not-cached"
	// NoEarlierRun: there is no cached run to compare with.
	NoEarlierRun Status = "no-earlier-run"
	// NoRecord: the last cached run predates input records (smith 0.1).
	NoRecord Status = "no-record"
)

// Report is why a task would or would not be restored from the cache.
type Report struct {
	Task        string                 `json:"task"`
	Status      Status                 `json:"status"`
	Key         string                 `json:"key,omitempty"`
	LastKey     string                 `json:"lastKey,omitempty"`
	LastRun     *time.Time             `json:"lastRun,omitempty"`
	Differences []taskcache.Difference `json:"differences,omitempty"`
}

// ExplainMiss computes the key of a task ("project:task") as a run would now
// and compares it with the task's last cached run, without running anything.
func ExplainMiss(s *session.Session, id string) (Report, error) {
	project, task, ok := strings.Cut(id, ":")
	if !ok || project == "" || task == "" {
		return Report{}, fmt.Errorf("%q is not <project>:<task>", id)
	}
	p := s.Graph.Get(project)
	if p == nil {
		return Report{}, fmt.Errorf("no project named %q", project)
	}
	def, ok := p.Tasks[task]
	if !ok {
		return Report{}, fmt.Errorf("%s has no %s task", project, task)
	}
	r := Report{Task: id}
	if !def.Cache {
		r.Status = NotCached
		return r, nil
	}
	g, err := taskgraph.Build(s.Graph, []string{project}, []string{task})
	if err != nil {
		return Report{}, err
	}
	store, err := s.Storage()
	if err != nil {
		return Report{}, err
	}
	key, now := taskrun.KeyOf(g, taskrun.Options{Root: s.Root, Files: s.Files, Projects: s.Graph, ToolVersion: s.Version}, id)
	r.Key = key
	if _, err := taskcache.Lookup(store, key); err == nil {
		r.Status = Hit
		return r, nil
	}
	last, lastKey, err := taskcache.LastRun(store, id)
	if errors.Is(err, taskcache.ErrNoLastRun) {
		r.Status = NoEarlierRun
		return r, nil
	}
	if err != nil {
		return Report{}, err
	}
	r.LastKey, r.LastRun = lastKey, &last.Created
	if last.Inputs == nil {
		r.Status = NoRecord
		return r, nil
	}
	r.Status = Changed
	r.Differences = taskcache.Diff(*last.Inputs, now)
	return r, nil
}
