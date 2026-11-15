package gitclient

import (
	"sort"
	"strings"
	"sync"
)

// refCache holds the tags and the branch read once, so a command that asks for
// them for every project starts one git process, not one per project.
type refCache struct {
	mu        sync.Mutex
	tagsReady bool
	tags      []string
	commits   map[string]string
	branchSet bool
	branch    string
}

// WithRefCache returns a client for the same repository that reads the tags
// and the current branch once and remembers them. It is for a command that
// only reads (a release plan): creating a tag through the client, or running
// git tag, checkout, switch or branch through Run, forgets what was read.
func (g *Git) WithRefCache() *Git { return &Git{Dir: g.Dir, refs: &refCache{}} }

// forget drops what the cache holds after a command that can change it.
func (g *Git) forget(args []string) {
	if g.refs == nil || len(args) == 0 {
		return
	}
	switch args[0] {
	case "tag", "checkout", "switch", "branch", "fetch", "pull", "reset":
		g.refs.mu.Lock()
		g.refs.tagsReady, g.refs.branchSet = false, false
		g.refs.mu.Unlock()
	}
}

// loadTags reads every tag and the commit it points to (an annotated tag is
// followed to its commit). It runs once per cache.
func (g *Git) loadTags() error {
	c := g.refs
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.tagsReady {
		return nil
	}
	out, err := g.raw("for-each-ref", "--format=%(refname:strip=2)%00%(objectname)%00%(*objectname)", "refs/tags")
	if err != nil {
		return err
	}
	c.tags, c.commits = nil, map[string]string{}
	for _, line := range strings.Split(strings.ReplaceAll(out, "\r\n", "\n"), "\n") {
		parts := strings.Split(line, "\x00")
		if len(parts) < 3 || parts[0] == "" {
			continue
		}
		commit := parts[2]
		if commit == "" {
			commit = parts[1]
		}
		c.tags = append(c.tags, parts[0])
		c.commits[parts[0]] = commit
	}
	sort.Strings(c.tags)
	c.tagsReady = true
	return nil
}

// cachedTags lists the tags that start with prefix, from the cache.
func (g *Git) cachedTags(prefix string) ([]string, error) {
	if err := g.loadTags(); err != nil {
		return nil, err
	}
	g.refs.mu.Lock()
	defer g.refs.mu.Unlock()
	var out []string
	for _, t := range g.refs.tags {
		if strings.HasPrefix(t, prefix) {
			out = append(out, t)
		}
	}
	return out, nil
}

// cachedTagCommit is the commit a tag points to, "" when there is no such tag.
func (g *Git) cachedTagCommit(tag string) (string, bool) {
	if err := g.loadTags(); err != nil {
		return "", false
	}
	g.refs.mu.Lock()
	defer g.refs.mu.Unlock()
	return g.refs.commits[tag], true
}
