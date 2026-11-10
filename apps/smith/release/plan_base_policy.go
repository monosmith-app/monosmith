package release

import (
	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// PreviewBase is the ref of the release branch when a plan made on the branch
// checked out is a preview of merging it there: the branch is not the release
// branch, not a channel branch and not detached (a detached pull request build
// is planned for its target branch, see PlanBranch). It tries the local branch,
// then the branch on the release remote (origin by default).
//
// Parameters: git is the repository; cfg is the [release] section.
// Returns: the ref and the release branch name, or "" and "" when the plan is
// the plain plan of the branch checked out (also when neither ref exists).
func PreviewBase(git *gitclient.Git, cfg workspace.ReleaseSection) (ref, branch string) {
	current := git.CurrentBranch()
	if current == "" || current == cfg.Branch || cfg.Branch == "" || ChannelOf(cfg, current) != nil {
		return "", ""
	}
	remote := cfg.Remote
	if remote == "" {
		remote = "origin"
	}
	for _, candidate := range []string{"refs/heads/" + cfg.Branch, "refs/remotes/" + remote + "/" + cfg.Branch} {
		if _, err := git.RevParse(candidate); err == nil {
			return candidate, cfg.Branch
		}
	}
	return "", ""
}
