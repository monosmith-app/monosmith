package splitrepo

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/integrations/gitclient"
)

// Request is one release to mirror.
type Request struct {
	// Prefix is the project's directory, workspace-relative.
	Prefix string
	// Ref is the monorepo's release tag the split is taken from.
	Ref string
	// Repo is the mirror, owner/repo on GitHub.
	Repo string
	// URL overrides the mirror's address (tests push to a local repository).
	URL     string
	Version string
	Token   string
	DryRun  bool
}

// Result says what happened.
type Result struct {
	// Mirror is owner/repo@tag.
	Mirror string
	// Skipped says why nothing was pushed (the mirror already has the tag).
	Skipped string
}

// Mirror splits the project's directory out of the history at the release
// tag and pushes it to the mirror's main branch and tag. A mirror that
// already has the tag is left alone, so a re-run release resumes. The token
// travels in git's environment-set config, never in an argument, so no error
// or log line can show it.
func Mirror(g *gitclient.Git, req Request) (Result, error) {
	tag := Tag(req.Version)
	res := Result{Mirror: req.Repo + "@" + tag}
	if req.DryRun {
		return res, nil
	}
	if req.Token == "" && req.URL == "" {
		return res, fmt.Errorf("%s is not set: %s", TokenEnv, Credential(req.Repo).Description)
	}
	url := req.URL
	if url == "" {
		url = "https://github.com/" + req.Repo + ".git"
	}
	env := authEnv(req.Token)
	remote, err := g.RunWithEnv(env, "ls-remote", "--tags", url, "refs/tags/"+tag)
	if err != nil {
		return res, fmt.Errorf("reading %s: %w", req.Repo, err)
	}
	if strings.TrimSpace(remote) != "" {
		res.Skipped = req.Repo + " already has " + tag
		return res, nil
	}
	commit, err := split(g, req.Prefix, req.Ref)
	if err != nil {
		return res, err
	}
	// --atomic: main and the tag land together or not at all, so the tag is a
	// trustworthy sign the mirror is done.
	if _, err := g.RunWithEnv(env, "push", "--atomic", "--no-verify", url, commit+":refs/heads/"+Branch, commit+":refs/tags/"+tag); err != nil {
		if strings.Contains(err.Error(), "non-fast-forward") || strings.Contains(err.Error(), "fetch first") {
			return res, fmt.Errorf("%s's %s has commits the split does not: the mirror is read-only, written only by smith release; reset its %s to the last split (or delete it) and re-run: %w", req.Repo, Branch, Branch, err)
		}
		return res, fmt.Errorf("pushing to %s: %w (the monorepo is tagged; re-run the release to push the mirror)", req.Repo, err)
	}
	return res, nil
}

// split returns the commit holding the history of prefix alone, at ref. It
// is always git subtree split: another splitter (splitsh-lite) makes other
// commits from the same history, so mixing them would leave the mirror's
// main unable to fast-forward.
func split(g *gitclient.Git, prefix, ref string) (string, error) {
	out, err := g.Run("subtree", "split", "--prefix="+prefix, ref)
	if err != nil {
		return "", fmt.Errorf("splitting %s: %w", prefix, err)
	}
	lines := strings.Fields(out)
	if len(lines) == 0 {
		return "", errors.New("git subtree split printed no commit for " + prefix)
	}
	return lines[len(lines)-1], nil
}

// authEnv authenticates to github.com with the token as actions/checkout
// does, first emptying the header list, so a checkout's own GITHUB_TOKEN
// header is not sent beside it.
func authEnv(token string) []string {
	if token == "" {
		return nil
	}
	header := "AUTHORIZATION: basic " + base64.StdEncoding.EncodeToString([]byte("x-access-token:"+token))
	key := "http.https://github.com/.extraheader"
	return []string{
		"GIT_CONFIG_COUNT=2",
		"GIT_CONFIG_KEY_0=" + key, "GIT_CONFIG_VALUE_0=",
		"GIT_CONFIG_KEY_1=" + key, "GIT_CONFIG_VALUE_1=" + header,
		"GIT_TERMINAL_PROMPT=0",
	}
}

// Token is the mirror token from the environment.
func Token() string { return os.Getenv(TokenEnv) }

// Ready says what this machine lacks to mirror a release, or nothing: git
// subtree is a contrib command some git builds leave out, and the token is
// passed through GIT_CONFIG_COUNT, which git reads from 2.31 on (an older git
// ignores it and fails to authenticate with no hint why).
func Ready(g *gitclient.Git) []string {
	var problems []string
	// git subtree -h prints its usage and exits 129: read the text, wherever it went.
	usage, err := g.Run("subtree", "-h")
	if err != nil {
		usage += err.Error()
	}
	if !strings.Contains(usage, "subtree split") {
		problems = append(problems, "git subtree is not installed (it is part of git's contrib; install a git that includes it)")
	}
	version, err := g.Run("version")
	if err != nil {
		return append(problems, "git version: "+err.Error())
	}
	if !atLeast(version, 2, 31) {
		problems = append(problems, "mirroring needs git 2.31 or newer (it passes the token through GIT_CONFIG_COUNT): "+version)
	}
	return problems
}

// atLeast reads "git version 2.41.0.windows.1" and compares major.minor.
func atLeast(version string, major, minor int) bool {
	fields := strings.Fields(version)
	if len(fields) < 3 {
		return false
	}
	var gotMajor, gotMinor int
	if _, err := fmt.Sscanf(fields[2], "%d.%d", &gotMajor, &gotMinor); err != nil {
		return false
	}
	return gotMajor > major || (gotMajor == major && gotMinor >= minor)
}
