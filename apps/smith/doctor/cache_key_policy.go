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
	return []protocol.Finding{{
		Check: "the remote cache has its key", OK: ok, Warning: true,
		Detail: "[cache] remote is " + remote + ", but SMITH_CACHE_KEY is not set: the remote cache is off on this machine, and tasks use the local cache only",
		Remedy: "export SMITH_CACHE_KEY with the value the team shares (the pipeline keeps it as a secret; `smith ci check` lists what to create), plus the backend's credentials, or leave it unset to work with the local cache",
	}}
}
