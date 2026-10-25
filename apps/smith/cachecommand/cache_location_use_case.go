package cachecommand

import (
	"os"

	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// location is where the workspace keeps its local cache and what it is
// allowed to grow to.
type location struct {
	Dir        string
	LimitBytes int64
	Remote     session.RemotePlan
}

// locate finds the workspace containing the current directory (honouring
// SMITH_WORKSPACE_ROOT) and reads its cache settings. It does not open a
// session: the cache commands need no plugins and no git.
func locate() (location, error) {
	return locateFrom(".")
}

func locateFrom(dir string) (location, error) {
	root, err := workspace.FindRoot(dir)
	if err != nil {
		return location{}, err
	}
	cfg, err := workspace.Load(root)
	if err != nil {
		return location{}, err
	}
	return location{
		Dir:        session.LocalCacheDir(root, cfg, os.Getenv),
		LimitBytes: session.CacheLimitBytes(cfg),
		Remote:     session.PlanRemote(cfg, os.Getenv),
	}, nil
}
