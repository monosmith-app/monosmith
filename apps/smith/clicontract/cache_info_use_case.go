package clicontract

import (
	"fmt"
	"os"

	"github.com/monosmith-app/monosmith/apps/smith/cachemaintenance"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// cacheInfo measures the local task cache of the workspace at root: a stat of
// each file, nothing hashed. It is nil when the workspace or the cache cannot
// be read, so smith info still answers.
func cacheInfo(root string) *CacheInfo {
	cfg, err := workspace.Load(root)
	if err != nil {
		return nil
	}
	dir := session.LocalCacheDir(root, cfg, os.Getenv)
	limit := session.CacheLimitBytes(cfg)
	status, err := cachemaintenance.Inspect(dir, limit)
	if err != nil {
		return nil
	}
	return &CacheInfo{Dir: dir, Bytes: status.Bytes, LimitBytes: limit}
}

// Line is the one line `smith info` shows: the size against the limit.
func (c CacheInfo) Line() string {
	if c.LimitBytes <= 0 {
		return fmt.Sprintf("%s, no limit", megabytes(c.Bytes))
	}
	return fmt.Sprintf("%s of %s (%.0f%%)", megabytes(c.Bytes), megabytes(c.LimitBytes), float64(c.Bytes)*100/float64(c.LimitBytes))
}

func megabytes(n int64) string {
	if n < 1<<20 {
		return fmt.Sprintf("%.1f MB", float64(n)/(1<<20))
	}
	return fmt.Sprintf("%.0f MB", float64(n)/(1<<20))
}
