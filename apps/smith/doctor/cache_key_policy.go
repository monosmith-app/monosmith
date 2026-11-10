package doctor

import (
	"regexp"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/doctorchecks"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// unexpanded is the text Azure Pipelines leaves for a variable it has no
// value for.
var unexpanded = regexp.MustCompile(`^\$\([A-Za-z0-9_.]+\)$`)

// cacheKeyFinding checks that a configured remote cache can be used here: its
// key is set. Without it the remote is off and tasks use the local cache,
// which is a slower run and never a wrong one, so this only warns.
func cacheKeyFinding(remote string, getenv func(string) string) []protocol.Finding {
	if remote == "" {
		return nil
	}
	key := getenv("SMITH_CACHE_KEY")
	ok := key != "" && !unexpanded.MatchString(key)
	problem := "SMITH_CACHE_KEY is not set, so the remote cache " + remote + " is off"
	if key != "" {
		problem = "SMITH_CACHE_KEY is " + key + ", a variable the pipeline did not expand, so the remote cache is off"
	}
	findings := []protocol.Finding{{
		Check: "the remote cache has its key", OK: ok, Warning: true, Area: doctorchecks.AreaConfig, Problem: doctorchecks.ProblemWhen(!ok, problem),
		Detail: "[cache] remote is " + remote + ", but SMITH_CACHE_KEY is not set: the remote cache is off on this machine, and tasks use the local cache only",
		Remedy: "make a key with `smith cache key` (once, for the whole team; see docs/reference/tasks-and-cache.md#set-up-the-remote-cache), export it as SMITH_CACHE_KEY (the pipeline keeps it as a secret; `smith ci check` lists what to create), plus the backend's credentials, or leave it unset to work with the local cache",
	}}
	return append(findings, cacheBackendFinding(remote, ok, getenv)...)
}

// cacheBackendFinding checks the credential the backend cannot do without: an
// Azure Blob remote needs its SAS token, and without it smith turns the remote
// off at the first task with a line that is easy to miss in a long run. It is
// only asked once the key is there, since without the key the remote is off
// already and that is what the other finding says.
func cacheBackendFinding(remote string, keyOK bool, getenv func(string) string) []protocol.Finding {
	if !keyOK || !strings.HasPrefix(remote, "azblob://") {
		return nil
	}
	token := getenv("AZURE_STORAGE_SAS_TOKEN")
	ok := token != "" && !unexpanded.MatchString(token)
	return []protocol.Finding{{
		Check: "the remote cache has its SAS token", OK: ok, Warning: true, Area: doctorchecks.AreaConfig,
		Problem: doctorchecks.ProblemWhen(!ok, "AZURE_STORAGE_SAS_TOKEN is not set, so the remote cache "+remote+" is off"),
		Detail:  "[cache] remote is " + remote + " and SMITH_CACHE_KEY is set, but AZURE_STORAGE_SAS_TOKEN is not: smith cannot reach the container, so nothing is read from it or written to it",
		Remedy:  "export AZURE_STORAGE_SAS_TOKEN with a SAS token for the container (read, write, create and list permission): `smith cache setup` makes one and copies it, and the pipeline keeps it as a secret variable",
	}}
}
