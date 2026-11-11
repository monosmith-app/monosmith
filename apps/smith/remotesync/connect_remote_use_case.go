package remotesync

import (
	"errors"
	"os"

	"github.com/monosmith-app/monosmith/apps/smith/remotecache"
	"github.com/monosmith-app/monosmith/apps/smith/session"
)

// Remote is the remote cache as smith reaches it: the store, the key that
// signs its objects, and the settings it was opened from.
type Remote struct {
	Store  remotecache.ObjectStore
	Secret []byte
	// URL is the configured remote (smith.toml or SMITH_CACHE_REMOTE).
	URL string
	// ReadOnly is true when this machine reads the remote and never writes it.
	ReadOnly bool
}

// ErrNoRemote is returned by Connect when no remote cache is configured.
var ErrNoRemote = errors.New("no remote cache is set: [cache] remote in smith.toml and SMITH_CACHE_REMOTE are empty")

// ErrNoKey is returned by Connect when SMITH_CACHE_KEY is not set.
var ErrNoKey = errors.New("SMITH_CACHE_KEY is not set, and smith shares nothing unsigned")

// Connect opens the remote cache a plan names, with the signing key from the
// environment, as a run does (session.Session.Storage).
//
// Parameters: plan is the workspace's remote cache settings (session.PlanRemote).
// Returns: the remote; ErrNoRemote without a remote, ErrNoKey without
// SMITH_CACHE_KEY, or the error of remotecache.Open (credentials that are
// missing, a URL smith cannot use).
func Connect(plan session.RemotePlan) (Remote, error) {
	if plan.URL == "" {
		return Remote{}, ErrNoRemote
	}
	if !plan.Active {
		return Remote{}, ErrNoKey
	}
	store, err := remotecache.Open(plan.URL)
	if err != nil {
		return Remote{}, err
	}
	return Remote{Store: store, Secret: []byte(os.Getenv("SMITH_CACHE_KEY")), URL: plan.URL, ReadOnly: plan.ReadOnly}, nil
}
