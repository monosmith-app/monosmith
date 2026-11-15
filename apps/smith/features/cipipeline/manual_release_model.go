package cipipeline

import (
	"cmp"
	"strings"
)

// ManualRelease is a release a person started by hand from the CI provider
// (the release_bump input of the GitHub workflow, an Azure runtime parameter, a
// GitLab or Bitbucket variable): smith ci release --bump p=level --force.
type ManualRelease struct {
	// Bumps are the --bump values, "project=level".
	Bumps []string
	// Force is --force: a package no commit releases is released anyway.
	Force bool
}

// Active says whether the run is a manual release.
func (r ManualRelease) Active() bool { return len(r.Bumps) > 0 || r.Force }

// actorEnv lists, per provider, the variable that names who started a run.
var actorEnv = []string{"GITHUB_ACTOR", "BUILD_REQUESTEDFOR", "GITLAB_USER_LOGIN", "CIRCLE_USERNAME", "BITBUCKET_STEP_TRIGGERER_UUID"}

// Actor names who started the run, from the provider's environment.
//
// Parameters: getenv reads an environment variable.
// Returns: the name, or "unknown" when the provider does not say.
func Actor(getenv func(string) string) string {
	for _, name := range actorEnv {
		if v := getenv(name); v != "" {
			return v
		}
	}
	return "unknown"
}

// Sentence says in the step summary and the log that the release was manual,
// with what was asked and who asked.
//
// Parameters: actor is who started the run.
// Returns: one sentence.
func (r ManualRelease) Sentence(actor string) string {
	return "Manual release: started by " + actor + " with --bump " + cmp.Or(strings.Join(r.Bumps, " --bump "), "(none)") + " --force. No push triggered it."
}
