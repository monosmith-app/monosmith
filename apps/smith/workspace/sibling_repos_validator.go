package workspace

import (
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/problem"
)

// repoName is owner/name as GitHub spells it.
var repoName = regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`)

// validateSiblingRepos checks [[ci.sibling_repos]] and fills each dir's
// default (the repository's name). The checkout is a GitHub Actions step, so
// the entries need GitHub as the only provider.
func validateSiblingRepos(cfg *Config) error {
	repos := cfg.CI.SiblingRepos
	if len(repos) == 0 {
		return nil
	}
	if len(cfg.CI.Providers) != 1 || cfg.CI.Providers[0] != "github" {
		return problem.New(
			fmt.Sprintf("%s: [[ci.sibling_repos]] works on GitHub Actions only", FileName),
			`set [ci] providers = ["github"], or check the repositories out in your own pipeline`)
	}
	envs, dirs := map[string]bool{}, map[string]bool{}
	for i := range repos {
		r := &repos[i]
		at := fmt.Sprintf("%s: [[ci.sibling_repos]] entry %d", FileName, i+1)
		if !repoName.MatchString(r.Repo) {
			return problem.New(fmt.Sprintf("%s: repo %q is not owner/name", at, r.Repo),
				`write the GitHub repository as repo = "owner/name"`)
		}
		if r.Dir == "" {
			r.Dir = r.Repo[strings.Index(r.Repo, "/")+1:]
		}
		clean := path.Clean(r.Dir)
		if clean != r.Dir || clean == "." || clean == ".." || strings.HasPrefix(clean, "../") || path.IsAbs(clean) || strings.Contains(clean, `\`) || strings.HasPrefix(clean, ".git") {
			return problem.New(fmt.Sprintf("%s: dir %q is not a plain directory inside the workspace", at, r.Dir),
				`use a relative directory with forward slashes and no "..", such as dir = "VideoPlugins"`)
		}
		if dirs[r.Dir] {
			return problem.New(fmt.Sprintf("%s: dir %q is used by another sibling repo", at, r.Dir), "give each repository its own dir")
		}
		dirs[r.Dir] = true
		if !envName.MatchString(r.Env) {
			return problem.New(fmt.Sprintf("%s: env %q is not an environment variable name", at, r.Env),
				`name the variable that will hold the checkout's path, such as env = "VIDEOPLUGINS"`)
		}
		if envs[r.Env] {
			return problem.New(fmt.Sprintf("%s: env %q is used by another sibling repo", at, r.Env), "give each repository its own variable")
		}
		envs[r.Env] = true
		if strings.ContainsAny(r.Ref, " \t\n") {
			return problem.New(fmt.Sprintf("%s: ref %q holds whitespace", at, r.Ref), `use a branch, tag or commit, such as ref = "main"`)
		}
		if r.TokenSecret != "" && !envName.MatchString(r.TokenSecret) {
			return problem.New(fmt.Sprintf("%s: token_secret %q is not a secret name", at, r.TokenSecret),
				`name the GitHub secret, such as token_secret = "VIDEOPLUGINS_TOKEN"`)
		}
	}
	return nil
}
