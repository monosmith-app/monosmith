package session

import "regexp"

// cacheEnvironment are the variables that configure the remote cache and
// its backends.
var cacheEnvironment = []string{
	"SMITH_CACHE_KEY", "SMITH_CACHE_TOKEN", "AZURE_STORAGE_SAS_TOKEN",
	"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_SESSION_TOKEN", "AWS_REGION",
}

// azureMacro is how Azure Pipelines leaves a variable it has no value for:
// the step's env: maps $(SMITH_CACHE_KEY) to the text "$(SMITH_CACHE_KEY)",
// which is what a build of a fork gets, as secrets are withheld from it.
var azureMacro = regexp.MustCompile(`^\$\([A-Za-z0-9_.]+\)$`)

// unexpandedMacros are the names among those given whose value is an Azure
// macro that was never replaced. Such a value is no value: a fork's build
// must run with the remote cache off, not sign with a key that reads
// "$(SMITH_CACHE_KEY)".
func unexpandedMacros(getenv func(string) string, names []string) []string {
	var out []string
	for _, name := range names {
		if azureMacro.MatchString(getenv(name)) {
			out = append(out, name)
		}
	}
	return out
}

// pullRequest reports whether CI is building a pull request or a merge
// request. Those runs read the shared cache and never write it: a
// contributor's build could otherwise store a result that every later run
// replays, and the signature only proves the object came from someone who
// holds the key.
func pullRequest(getenv func(string) string) bool {
	switch {
	case getenv("GITHUB_EVENT_NAME") == "pull_request", getenv("GITHUB_EVENT_NAME") == "pull_request_target":
		return true
	case getenv("BUILD_REASON") == "PullRequest":
		return true
	case getenv("CI_PIPELINE_SOURCE") == "merge_request_event":
		return true
	}
	return false
}
