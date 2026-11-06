package doctor

import (
	"regexp"

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
	return []protocol.Finding{{
		Check: "the remote cache has its key", OK: ok, Warning: true, Area: AreaConfig, Problem: problemWhen(!ok, problem),
		Detail: "[cache] remote is " + remote + ", but SMITH_CACHE_KEY is not set: the remote cache is off on this machine, and tasks use the local cache only",
		Remedy: "make a key with `smith cache key` (once, for the whole team; see docs/reference/tasks-and-cache.md#set-up-the-remote-cache), export it as SMITH_CACHE_KEY (the pipeline keeps it as a secret; `smith ci check` lists what to create), plus the backend's credentials, or leave it unset to work with the local cache",
	}}
}
