package remotesync

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/remotecache"
	"github.com/monosmith-app/monosmith/apps/smith/taskcache"
)

// RemoteObjects are the objects a remote clear deletes.
type RemoteObjects struct {
	// Names are the object names ("ac/<key>", "cas/<key>").
	Names []string
	// Listed is true when the store listed them. A plain https:// store
	// cannot list: Names are then the entries this machine's cache holds,
	// which the remote may or may not have.
	Listed bool
}

// FindObjects lists the cache objects under the remote's prefix. A store that
// cannot list gives the names of the local cache's entries and blobs instead.
//
// Parameters: remote is the remote cache (Connect); local is the local cache,
// nil when there is none.
// Returns: the objects; the store's error (remotecache.Refused for a token
// without list rights).
func FindObjects(remote Remote, local taskcache.Storage) (RemoteObjects, error) {
	var out RemoteObjects
	collect := func(name string) error { out.Names = append(out.Names, name); return nil }
	err := remote.Store.List("ac/", collect)
	if err == nil {
		err = remote.Store.List("cas/", collect)
		out.Listed = true
	}
	if errors.Is(err, remotecache.ErrCannotList) {
		out = RemoteObjects{}
		if local != nil {
			for _, kind := range []taskcache.Kind{taskcache.AC, taskcache.CAS} {
				err := local.Walk(kind, func(key string, _ int64, _ time.Time) error {
					return collect(remotecache.ObjectName(kind, key))
				})
				if err != nil {
					return out, err
				}
			}
		}
		err = nil
	}
	sort.Strings(out.Names)
	return out, err
}

// ClearResult is the outcome of a remote clear.
type ClearResult struct {
	// Store names the remote (StoreName).
	Store string `json:"store"`
	// Listed is false when the store could not list and smith deleted the
	// names the local cache knows.
	Listed bool `json:"listed"`
	// Found is how many objects were to go; Deleted and Failed split them.
	Found   int `json:"found"`
	Deleted int `json:"deleted"`
	Failed  int `json:"failed"`
	// Errors are the first reasons a delete failed.
	Errors []string `json:"errors,omitempty"`
}

// maxErrors is how many failure reasons a ClearResult keeps.
const maxErrors = 5

// ClearRemote deletes objects from the remote cache, 8 at a time. It stops at
// the first refusal (a token without delete rights) and when ctx ends.
//
// Parameters: ctx stops it between objects; remote is the remote cache;
// objects are what FindObjects found; progress, when set, is told the count
// done so far, one call at a time.
// Returns: what happened; ErrReadOnly for a read-only remote, the refusal of
// the store, or ctx.Err(). Other failures are counted and do not stop it.
func ClearRemote(ctx context.Context, remote Remote, objects RemoteObjects, progress func(done int)) (ClearResult, error) {
	result := ClearResult{Store: StoreName(remote.URL), Listed: objects.Listed, Found: len(objects.Names)}
	if remote.ReadOnly {
		return result, ErrReadOnly
	}
	var (
		mu      sync.Mutex
		refused error
		wg      sync.WaitGroup
		work    = make(chan string)
	)
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for name := range work {
				err := remote.Store.Delete(name)
				mu.Lock()
				switch {
				case remotecache.Refused(err):
					if refused == nil {
						refused = err
					}
				case err != nil:
					result.Failed++
					if len(result.Errors) < maxErrors {
						result.Errors = append(result.Errors, err.Error())
					}
				default:
					result.Deleted++
				}
				if progress != nil && err == nil {
					progress(result.Deleted)
				}
				mu.Unlock()
			}
		}()
	}
feed:
	for _, name := range objects.Names {
		mu.Lock()
		stop := refused != nil
		mu.Unlock()
		if stop {
			break
		}
		select {
		case <-ctx.Done():
			break feed
		case work <- name:
		}
	}
	close(work)
	wg.Wait()
	if refused != nil {
		return result, refused
	}
	return result, ctx.Err()
}
