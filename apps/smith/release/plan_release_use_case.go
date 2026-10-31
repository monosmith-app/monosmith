package release

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
)

// Item is one package's planned release.
type Item struct {
	Project string `json:"project"`
	Plugin  string `json:"plugin"`
	// Current is the last released version ("" when never released).
	Current string `json:"current,omitempty"`
	Next    string `json:"next"`
	Level   string `json:"level"`
	Tag     string `json:"tag"`
	// Reason explains why the package releases.
	Reason  string   `json:"reason"`
	Commits []string `json:"commits,omitempty"`
	Forced  bool     `json:"forced,omitempty"`
	// Changes are the releasing commits, for the release notes.
	Changes []Change `json:"changes,omitempty"`
	// APIWarnings say that the commits ask for a smaller bump than the change
	// of the public API needs, and which bump to use (see CheckAPI).
	APIWarnings []string `json:"apiWarnings,omitempty"`
}

// Plan is every package that will release, in dependency order.
type Plan struct {
	Items []Item `json:"items"`
	// Unchanged lists packages with nothing to release.
	Unchanged []string `json:"unchanged,omitempty"`
}

// TagPrefix is the release tag prefix of a project.
func TagPrefix(p *projectgraph.Project) string {
	if p.BareTags {
		return ""
	}
	if p.TagPrefix != "" {
		return p.TagPrefix
	}
	return p.Name + "@"
}

// LatestTag returns the highest released version of a project and its tag.
func LatestTag(git *gitclient.Git, p *projectgraph.Project) (Version, string, bool) {
	prefix := TagPrefix(p)
	tags, _ := git.Tags(prefix)
	var best Version
	var bestTag string
	found := false
	for _, t := range tags {
		v, err := ParseVersion(strings.TrimPrefix(t, prefix))
		if err != nil || v.Pre != "" {
			continue
		}
		if !found || best.Less(v) {
			best, bestTag, found = v, t, true
		}
	}
	return best, bestTag, found
}

// MakePlan decides each publishable package's next version from the commits
// that touched it since its last tag. A package whose dependency releases is
// released too (at least a patch), so its published range stays correct.
func MakePlan(git *gitclient.Git, graph *projectgraph.Graph) (Plan, error) {
	order, err := graph.Order()
	if err != nil {
		return Plan{}, err
	}
	var plan Plan
	releasing := map[string]string{}
	for _, name := range order {
		p := graph.Get(name)
		if !p.Releases() {
			continue
		}
		latest, tag, released := LatestTag(git, p)
		current := latest
		if !released {
			current = Version{}
			if v, err := ParseVersion(p.Version); err == nil {
				current = v
			}
		}
		raw, err := git.Log(tag, SourceRoots(graph, p)...)
		if err != nil {
			return Plan{}, err
		}
		var commits []Commit
		var subjects []string
		var changes []Change
		for _, c := range raw {
			pc := ParseCommit(c.Hash, c.Subject, c.Body)
			commits = append(commits, pc)
			if LevelOf(pc) != None {
				subjects = append(subjects, c.Subject)
				changes = append(changes, Change{Type: pc.Type, Scopes: pc.Scopes, Description: pc.Description, Breaking: pc.Breaking, BreakingNote: pc.BreakingNote, Hash: pc.Hash, Author: c.Author})
			}
		}
		level := LevelOfAll(commits)
		reason := fmt.Sprintf("%d releasing commit(s)", len(subjects))
		if level == None {
			for _, d := range p.Deps {
				if releasing[d] != "" {
					level, reason = Patch, "dependency "+d+" "+releasing[d]
					break
				}
			}
		}
		item := Item{Project: name, Plugin: p.Plugin, Commits: subjects, Changes: changes, Reason: reason, Level: level.String()}
		if released {
			item.Current = latest.String()
		}
		forced, err := directiveFor(git, tag, name)
		if err != nil {
			return Plan{}, err
		}
		next := current.Bump(level)
		if forced != "" {
			v, err := ParseVersion(forced)
			if err != nil {
				return Plan{}, fmt.Errorf("version(%s)[%s]: %w", name, forced, err)
			}
			next, item.Forced, item.Reason, item.Level = v, true, "forced by version("+name+")["+forced+"]", "forced"
		} else if level == None {
			plan.Unchanged = append(plan.Unchanged, name)
			continue
		}
		if released && !latest.Less(next) {
			return Plan{}, fmt.Errorf("%s: next version %s is not above the released %s", name, next, latest)
		}
		item.Next = next.String()
		item.Tag = TagPrefix(p) + item.Next
		plan.Items = append(plan.Items, item)
		releasing[name] = item.Next
	}
	return plan, nil
}

// SourceRoots are the directories whose commits release a project: its own
// root and the roots of the internal projects it bundles (dependencies that
// never release on their own), so a fix in a lib releases the app using it.
func SourceRoots(graph *projectgraph.Graph, p *projectgraph.Project) []string {
	roots := []string{p.Root}
	for _, d := range graph.TransitiveDeps(p.Name) {
		if dep := graph.Get(d); dep != nil && !dep.Releases() {
			roots = append(roots, dep.Root)
		}
	}
	return roots
}

// directiveFor finds the newest version(<project>)[x] directive since a tag,
// anywhere in the repository.
func directiveFor(git *gitclient.Git, since, project string) (string, error) {
	commits, err := git.Log(since)
	if err != nil {
		return "", err
	}
	for _, c := range commits {
		if d, ok := ParseDirective(c.Subject); ok && d.Project == project {
			return d.Version, nil
		}
	}
	return "", nil
}
