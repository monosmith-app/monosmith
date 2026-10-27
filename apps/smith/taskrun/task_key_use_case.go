package taskrun

import (
	"sync"

	"github.com/monosmith-app/monosmith/apps/smith/taskcache"
	"github.com/monosmith-app/monosmith/apps/smith/taskgraph"
)

// keyer computes what a task's cache key is made of: its inputs, the sources
// of every project it depends on, and the keys of the tasks it depends on.
type keyer struct {
	opts     Options
	digests  *taskcache.FileDigests
	mu       sync.Mutex
	upstream map[string]string
}

func newKeyer(opts Options) *keyer {
	return &keyer{opts: opts, digests: taskcache.NewFileDigests(opts.Root), upstream: map[string]string{}}
}

// inputs builds the key inputs of t, given the keys of its dependencies by
// task id.
func (k *keyer) inputs(t *taskgraph.Task, deps map[string]string) taskcache.KeyInputs {
	selected := taskcache.SelectInputs(k.opts.Files, t.Project.Root, t.Def.Inputs, t.Def.Outputs)
	selected = append(selected, taskcache.DefaultInputs(selected, t.Def.Inputs)...)
	files := make([]taskcache.FileDigest, len(selected))
	for i, f := range selected {
		files[i] = taskcache.FileDigest{Path: f, Digest: k.digests.Digest(f)}
	}
	upstream := map[string]string{}
	for _, dep := range k.opts.Projects.TransitiveDeps(t.Project.Name) {
		upstream[dep] = k.sources(dep)
	}
	serve := ""
	if t.Def.Serve != "" {
		serve = t.Def.Serve + " " + t.Def.ServeURL
		// The app under test is part of what the suite proves.
		if sp, _, err := taskgraph.ServeTarget(k.opts.Projects, t.Project.Name, t.Def.Serve); err == nil && sp.Name != t.Project.Name {
			upstream[sp.Name] = k.sources(sp.Name)
			for _, dep := range k.opts.Projects.TransitiveDeps(sp.Name) {
				upstream[dep] = k.sources(dep)
			}
		}
	}
	return taskcache.KeyInputs{
		Serve: serve,
		Task:  t.ID, Command: t.Def.Command, Env: t.Def.Env, Files: files, Upstream: upstream,
		Deps: deps, Platform: Platform, Tool: k.opts.ToolVersion + "/" + t.Project.Toolset(), Toolchain: t.Project.Toolchain,
	}
}

func (k *keyer) sources(project string) string {
	k.mu.Lock()
	if v, ok := k.upstream[project]; ok {
		k.mu.Unlock()
		return v
	}
	k.mu.Unlock()
	p := k.opts.Projects.Get(project)
	files := taskcache.ProjectFiles(k.opts.Files, p.Root)
	digests := make([]taskcache.FileDigest, len(files))
	for i, f := range files {
		digests[i] = taskcache.FileDigest{Path: f, Digest: k.digests.Digest(f)}
	}
	v := taskcache.Combine(digests)
	k.mu.Lock()
	k.upstream[project] = v
	k.mu.Unlock()
	return v
}

// KeyOf computes the key of the task id in g, and what it is made of, as a
// run would now, without running anything.
func KeyOf(g *taskgraph.Graph, opts Options, id string) (string, taskcache.KeyInputs) {
	k := newKeyer(opts)
	keys := map[string]string{}
	var walk func(id string) taskcache.KeyInputs
	walk = func(id string) taskcache.KeyInputs {
		t := g.Tasks[id]
		deps := make(map[string]string, len(t.Deps))
		for _, d := range t.Deps {
			if _, done := keys[d]; !done {
				keys[d] = taskcache.Key(walk(d))
			}
			deps[d] = keys[d]
		}
		return k.inputs(t, deps)
	}
	in := walk(id)
	return taskcache.Key(in), in
}
