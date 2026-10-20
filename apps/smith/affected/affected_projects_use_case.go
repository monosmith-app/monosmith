package affected

import (
	"os"
	"path"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
)

// Result names the affected projects and why.
type Result struct {
	Projects []string `json:"projects"`
	// All is true when every project is affected; Reason says why.
	All    bool   `json:"all"`
	Reason string `json:"reason"`
	Base   string `json:"base,omitempty"`
	// Changed lists the changed files that were considered.
	Changed []string `json:"changed,omitempty"`
}

// neutralFiles never affect a project when changed at the workspace level.
var neutralFiles = []string{"*.md", "docs/**", "LICENSE", ".github/ISSUE_TEMPLATE/**", ".vscode/**", "*.code-workspace"}

// BaseFromEnv returns the pull request's target branch from the CI provider,
// or "" when not running for a pull request.
func BaseFromEnv() string {
	for _, name := range []string{"SMITH_BASE", "GITHUB_BASE_REF", "SYSTEM_PULLREQUEST_TARGETBRANCH", "CI_MERGE_REQUEST_TARGET_BRANCH_NAME"} {
		if v := strings.TrimSpace(os.Getenv(name)); v != "" {
			return strings.TrimPrefix(v, "refs/heads/")
		}
	}
	return ""
}

// Compute returns the projects affected by changes since base. It resolves
// base against origin first (CI checkouts rarely have the local branch), and
// falls back to every project whenever it cannot be sure.
func Compute(git *gitclient.Git, graph *projectgraph.Graph, base string) Result {
	all := func(reason string) Result {
		return Result{Projects: graph.Names(), All: true, Reason: reason, Base: base}
	}
	if base == "" {
		return all("no base to compare with")
	}
	ref := ""
	for _, candidate := range []string{"origin/" + base, base} {
		if _, err := git.RevParse(candidate); err == nil {
			ref = candidate
			break
		}
	}
	if ref == "" {
		_, _ = git.Run("fetch", "--no-tags", "origin", base)
		if _, err := git.RevParse("origin/" + base); err == nil {
			ref = "origin/" + base
		}
	}
	if ref == "" {
		return all("base " + base + " was not found")
	}
	mergeBase, err := git.MergeBase(ref, "HEAD")
	if err != nil {
		return all("no common ancestor with " + ref + " (shallow clone?)")
	}
	changed, err := git.ChangedSince(mergeBase)
	if err != nil {
		return all("could not list changes: " + err.Error())
	}
	touched := map[string]bool{}
	for _, f := range changed {
		owner := graph.Owner(f)
		if owner != nil {
			touched[owner.Name] = true
		}
		// A schema file belongs to its owner, wherever it lives.
		if schema := graph.SchemaOwner(f); schema != nil {
			touched[schema.Name] = true
			continue
		}
		if owner != nil {
			continue
		}
		// A file in an excluded directory belongs to no project.
		if !neutral(f) && !graph.Excluded(f) {
			r := all("workspace file " + f + " changed")
			r.Changed = changed
			return r
		}
	}
	result := map[string]bool{}
	for name := range touched {
		result[name] = true
		for _, d := range graph.TransitiveDependents(name) {
			result[d] = true
		}
	}
	names := make([]string, 0, len(result))
	for n := range result {
		names = append(names, n)
	}
	sort.Strings(names)
	reason := "changes since " + ref
	if len(names) == 0 {
		reason = "no project changed since " + ref
	}
	return Result{Projects: names, Reason: reason, Base: mergeBase, Changed: changed}
}

func neutral(file string) bool {
	for _, pattern := range neutralFiles {
		if strings.HasSuffix(pattern, "/**") {
			if strings.HasPrefix(file, strings.TrimSuffix(pattern, "**")) {
				return true
			}
			continue
		}
		if ok, _ := path.Match(pattern, file); ok {
			return true
		}
	}
	return false
}
