package remotecache

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/taskcache"
)

// Tiered is the local cache in front of a remote store. Reads try local
// first and keep what the remote had, once verified; writes go to both
// (unless ReadOnly). The remote never fails a run: a remote error is a
// warning, and an object that fails verification is a miss.
type Tiered struct {
	Local    taskcache.Storage
	Remote   ObjectStore
	Secret   []byte
	ReadOnly bool
	// Warn reports remote problems; each kind of problem once.
	Warn   func(string)
	warned map[string]bool
}

var _ taskcache.Storage = (*Tiered)(nil)

func name(kind taskcache.Kind, key string) string { return string(kind) + "/" + key }

// ObjectName is the name of a cache entry's object in a remote store, under
// the store's prefix: "ac/<key>" or "cas/<key>".
//
// Parameters: kind is the entry's kind; key is its hex digest.
// Returns: the object name.
func ObjectName(kind taskcache.Kind, key string) string { return name(kind, key) }

// Upload signs content with secret and writes it to store under the entry's
// object name. Every remote write of an entry goes through it, so an entry
// that `smith cache push` sends is signed exactly as a run signs it.
//
// Parameters: store is the remote; secret is SMITH_CACHE_KEY; kind and key
// name the entry; content is its plain bytes.
// Returns: the store's error, nil when the object was written.
func Upload(store ObjectStore, secret []byte, kind taskcache.Kind, key string, content []byte) error {
	return store.Put(name(kind, key), Sign(secret, string(kind), key, content))
}

func (t *Tiered) warn(what, detail string) {
	if t.warned == nil {
		t.warned = map[string]bool{}
	}
	if t.Warn != nil && !t.warned[what] {
		t.warned[what] = true
		t.Warn(what + ": " + detail)
	}
}

// pull fetches an entry from the remote into the local cache, verified;
// false when the remote does not have it, could not be asked, or sent
// something that fails verification.
func (t *Tiered) pull(kind taskcache.Kind, key string) bool {
	if kind == taskcache.Last {
		return false
	}
	object, found, err := t.Remote.Get(name(kind, key))
	if err != nil {
		t.warn("the remote cache could not be read", err.Error())
		return false
	}
	if !found {
		return false
	}
	content, err := Verify(t.Secret, string(kind), key, object)
	if err == nil && kind == taskcache.CAS {
		if sum := sha256.Sum256(content); hex.EncodeToString(sum[:]) != key {
			err = ErrTampered
		}
	}
	if err != nil {
		t.warn("a remote cache object was refused", name(kind, key)+": "+err.Error())
		return false
	}
	return t.Local.Put(kind, key, bytes.NewReader(content)) == nil
}

// Has implements taskcache.Storage; a blob only the remote has is fetched,
// so what Has promises, Get delivers.
func (t *Tiered) Has(kind taskcache.Kind, key string) (bool, error) {
	if ok, err := t.Local.Has(kind, key); ok || err != nil {
		return ok, err
	}
	return t.pull(kind, key), nil
}

// Get implements taskcache.Storage.
func (t *Tiered) Get(kind taskcache.Kind, key string) (io.ReadCloser, error) {
	r, err := t.Local.Get(kind, key)
	if err != taskcache.ErrMiss {
		return r, err
	}
	if !t.pull(kind, key) {
		return nil, taskcache.ErrMiss
	}
	return t.Local.Get(kind, key)
}

// Put implements taskcache.Storage.
func (t *Tiered) Put(kind taskcache.Kind, key string, content io.Reader) error {
	data, err := io.ReadAll(content)
	if err != nil {
		return err
	}
	if err := t.Local.Put(kind, key, bytes.NewReader(data)); err != nil {
		return err
	}
	if t.ReadOnly || kind == taskcache.Last {
		return nil
	}
	if kind == taskcache.CAS {
		if ok, err := t.Remote.Has(name(kind, key)); err == nil && ok {
			return nil
		}
	}
	if err := Upload(t.Remote, t.Secret, kind, key, data); err != nil {
		t.warn("the remote cache could not be written", err.Error())
	}
	return nil
}

// HasLocally implements taskcache.Layered: it asks the local cache only, so a
// run can tell a hit the remote served from one this machine already had.
func (t *Tiered) HasLocally(kind taskcache.Kind, key string) (bool, error) {
	return t.Local.Has(kind, key)
}

// Delete implements taskcache.Storage (the local copy only).
func (t *Tiered) Delete(kind taskcache.Kind, key string) error { return t.Local.Delete(kind, key) }

// Walk implements taskcache.Storage (the local copies only).
func (t *Tiered) Walk(kind taskcache.Kind, visit func(string, int64, time.Time) error) error {
	return t.Local.Walk(kind, visit)
}

// Touch implements taskcache.Storage (the local copy only).
func (t *Tiered) Touch(kind taskcache.Kind, key string) error { return t.Local.Touch(kind, key) }
