package release

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
)

// commitIndex answers "which commits touched these folders since that tag" for
// every project from one read of the history, where asking git once per
// project starts a process each time (slow on a repository of many projects).
// A question it cannot answer from what it read goes to git as before.
type commitIndex struct {
	git      *gitclient.Git
	revs     []gitclient.Revision
	pos      map[string]int
	boundary string
	// excluded caches, per tag, the commits reachable from it.
	excluded map[string]map[string]bool
}

// newCommitIndex reads the history that the given tags need: everything since
// the best common ancestor of their commits, or all of it when a project has no
// tag yet (an empty tag in tags). It returns nil when git cannot be read, and
// the callers then ask git directly.
func newCommitIndex(git *gitclient.Git, tags []string) *commitIndex {
	var commits []string
	for _, t := range tags {
		if t == "" {
			commits = nil
			break
		}
		c := git.TagCommit(t)
		if c == "" {
			commits = nil
			break
		}
		commits = append(commits, c)
	}
	boundary := git.CommonAncestor(commits)
	revs, err := git.History(boundary)
	if err != nil {
		return nil
	}
	x := &commitIndex{git: git, revs: revs, pos: make(map[string]int, len(revs)), boundary: boundary, excluded: map[string]map[string]bool{}}
	for i, r := range revs {
		x.pos[r.Hash] = i
	}
	return x
}

// reachable is the set of indexed commits reachable from the tag (nil for no
// tag). ok is false when the tag's commit lies outside what was read.
func (x *commitIndex) reachable(tag string) (set map[string]bool, ok bool) {
	if tag == "" {
		return nil, true
	}
	if set, hit := x.excluded[tag]; hit {
		return set, true
	}
	start := x.git.TagCommit(tag)
	if start == "" {
		return nil, false
	}
	set = map[string]bool{}
	if _, in := x.pos[start]; !in {
		// The tag sits on the boundary: nothing read is reachable from it.
		if start != x.boundary {
			return nil, false
		}
		x.excluded[tag] = set
		return set, true
	}
	stack := []string{start}
	for len(stack) > 0 {
		h := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if set[h] {
			continue
		}
		i, in := x.pos[h]
		if !in {
			continue
		}
		set[h] = true
		stack = append(stack, x.revs[i].Parents...)
	}
	x.excluded[tag] = set
	return set, true
}

// log is git.Log(since, roots...) answered from the index: the commits not
// reachable from since that changed a file under one of roots, newest first.
func (x *commitIndex) log(since string, roots []string) []gitclient.Commit {
	if x != nil {
		if skip, ok := x.reachable(since); ok {
			var out []gitclient.Commit
			for _, r := range x.revs {
				if !skip[r.Hash] && touches(r.Files, roots) {
					out = append(out, r.Commit)
				}
			}
			return out
		}
	}
	return nil
}

// answers reports whether the index can answer for the tag.
func (x *commitIndex) answers(tag string) bool {
	if x == nil {
		return false
	}
	_, ok := x.reachable(tag)
	return ok
}

// touches reports whether a file lies under one of the folders ("" and "."
// are the whole workspace).
func touches(files, roots []string) bool {
	for _, root := range roots {
		root = strings.TrimSuffix(strings.TrimPrefix(root, "./"), "/")
		if root == "" {
			return true
		}
		if root == "." {
			if len(files) > 0 {
				return true
			}
			continue
		}
		for _, f := range files {
			if f == root || strings.HasPrefix(f, root+"/") {
				return true
			}
		}
	}
	return false
}
