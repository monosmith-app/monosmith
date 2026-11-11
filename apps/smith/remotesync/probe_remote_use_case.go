package remotesync

import (
	"context"
	"errors"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/remotecache"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/taskcache"
)

// State is how the remote cache answers.
type State string

// The states of the remote cache.
const (
	// StateNone: no remote cache is set.
	StateNone State = "none"
	// StateConnected: the remote answers and accepts the credentials.
	StateConnected State = "connected"
	// StateUnreachable: the remote does not answer in time.
	StateUnreachable State = "unreachable"
	// StateRefused: the remote refuses the credentials, or smith has none to
	// give it (SMITH_CACHE_KEY or the store's own credentials are not set).
	StateRefused State = "refused"
)

// Status is what a probe of the remote cache found.
type Status struct {
	State State `json:"state"`
	// Detail says why, for a state other than connected and none.
	Detail string `json:"detail,omitempty"`
}

// probeKey is the key the probe asks for: no task has it, so a store that
// accepts the credentials answers "not found".
var probeKey = strings.Repeat("0", 64)

// Probe asks the remote cache of a plan for one object that does not exist,
// and reads the answer as a State. It gives up when ctx ends.
//
// Parameters: ctx bounds the wait (the caller sets a short timeout); plan is
// the workspace's remote cache settings.
// Returns: the status; never an error, since every failure is a State.
func Probe(ctx context.Context, plan session.RemotePlan) Status {
	remote, err := Connect(plan)
	switch {
	case errors.Is(err, ErrNoRemote):
		return Status{State: StateNone}
	case err != nil:
		return Status{State: StateRefused, Detail: err.Error()}
	}
	return ProbeStore(ctx, remote.Store)
}

// ProbeStore is Probe on a store already open.
//
// Parameters: ctx bounds the wait; store is the remote.
// Returns: connected for a 2xx or a 404, refused for a 401 or a 403,
// unreachable for any other error and when ctx ends first.
func ProbeStore(ctx context.Context, store remotecache.ObjectStore) Status {
	done := make(chan error, 1)
	go func() {
		_, err := store.Has(remotecache.ObjectName(taskcache.AC, probeKey))
		done <- err
	}()
	select {
	case <-ctx.Done():
		return Status{State: StateUnreachable, Detail: "no answer from " + store.Describe() + " in time"}
	case err := <-done:
		switch {
		case err == nil:
			return Status{State: StateConnected}
		case remotecache.Refused(err):
			return Status{State: StateRefused, Detail: err.Error()}
		default:
			return Status{State: StateUnreachable, Detail: err.Error()}
		}
	}
}
