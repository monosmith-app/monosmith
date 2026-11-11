package sources

import (
	"context"
	"os"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/remotesync"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/taskcache"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// The bounds of the background questions to the remote cache: the shell
// shows "checking" until they end, so they end soon.
const (
	probeTimeout   = 5 * time.Second
	countTimeout   = 10 * time.Second
	countMaxChecks = 500
	countParallel  = 8
)

// remoteSettings reads the workspace's cache settings without opening a
// session (no plugin starts).
func remoteSettings(dir string) (root string, cfg workspace.Config, plan session.RemotePlan, err error) {
	root, err = workspace.FindRoot(dir)
	if err != nil {
		return "", cfg, plan, err
	}
	cfg, err = workspace.Load(root)
	if err != nil {
		return "", cfg, plan, err
	}
	return root, cfg, session.PlanRemote(cfg, os.Getenv), nil
}

// remoteSet says whether the workspace in dir has a remote cache.
func remoteSet(dir string) bool {
	_, _, plan, err := remoteSettings(dir)
	return err == nil && plan.URL != ""
}

// probeRemote asks the remote cache of the workspace in dir how it answers.
func probeRemote(dir string) remotesync.Status {
	_, _, plan, err := remoteSettings(dir)
	if err != nil {
		return remotesync.Status{State: remotesync.StateNone}
	}
	ctx, cancel := context.WithTimeout(context.Background(), probeTimeout)
	defer cancel()
	return remotesync.Probe(ctx, plan)
}

// countPending counts the local entries the remote cache of the workspace in
// dir lacks, capped so a large cache cannot keep Home waiting.
func countPending(dir string) RemotePending {
	root, cfg, plan, err := remoteSettings(dir)
	if err != nil {
		return RemotePending{Err: err}
	}
	remote, err := remotesync.Connect(plan)
	if err != nil {
		return RemotePending{Err: err}
	}
	cacheDir := session.LocalCacheDir(root, cfg, os.Getenv)
	if _, err := os.Stat(cacheDir); err != nil {
		return RemotePending{}
	}
	local, err := taskcache.NewLocalStorage(cacheDir)
	if err != nil {
		return RemotePending{Err: err}
	}
	ctx, cancel := context.WithTimeout(context.Background(), countTimeout)
	defer cancel()
	p, err := remotesync.CountUnpushed(ctx, local, remote.Store, remotesync.CountLimits{MaxEntries: countMaxChecks, Parallel: countParallel})
	return RemotePending{Pending: p, Err: err}
}
