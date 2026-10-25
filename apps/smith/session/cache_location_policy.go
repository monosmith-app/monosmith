package session

import (
	"path/filepath"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// LocalCacheDir is the directory of the workspace's local cache: [cache] dir,
// or SMITH_CACHE_DIR when that is set, made absolute against the workspace
// root. It reads no file, so the cache commands can use it without opening a
// session.
func LocalCacheDir(root string, cfg workspace.Config, getenv func(string) string) string {
	dir := cfg.Cache.Dir
	if dir == "" {
		dir = ".smith/cache"
	}
	if env := getenv("SMITH_CACHE_DIR"); env != "" {
		dir = env
	}
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(root, filepath.FromSlash(dir))
	}
	return dir
}

// CacheLimitBytes is [cache] max_size_mb in bytes; zero or less in smith.toml
// means no limit, which is reported as 0.
func CacheLimitBytes(cfg workspace.Config) int64 {
	if cfg.Cache.MaxSizeMB <= 0 {
		return 0
	}
	return int64(cfg.Cache.MaxSizeMB) << 20
}

// RemotePlan is what the remote cache settings amount to.
type RemotePlan struct {
	// URL is the configured remote, "" when there is none.
	URL string
	// Active is true when the remote is used: it needs SMITH_CACHE_KEY, and
	// nothing is shared unsigned.
	Active bool
	// ReadOnly is true when the remote is read and never written.
	ReadOnly bool
}

// PlanRemote reads the remote cache settings of a workspace from its
// configuration and the environment.
func PlanRemote(cfg workspace.Config, getenv func(string) string) RemotePlan {
	plan := RemotePlan{URL: cfg.Cache.Remote}
	if env := getenv("SMITH_CACHE_REMOTE"); env != "" {
		plan.URL = env
	}
	if plan.URL == "" {
		return plan
	}
	// A key that is an unexpanded Azure macro is no key.
	secret := getenv("SMITH_CACHE_KEY")
	plan.Active = secret != "" && len(unexpandedMacros(func(string) string { return secret }, []string{"SMITH_CACHE_KEY"})) == 0
	plan.ReadOnly = cfg.Cache.RemoteReadOnly || getenv("SMITH_CACHE_READ_ONLY") != "" || pullRequest(getenv)
	return plan
}
