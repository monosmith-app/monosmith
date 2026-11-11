package taskcache

import (
	"errors"
	"io"
	"time"
)

// Kind separates action-cache entries from content-addressed blobs.
type Kind string

// Storage kinds.
const (
	AC  Kind = "ac"
	CAS Kind = "cas"
	// Last holds, per task, the key of its most recent cached run (keyed by
	// the SHA-256 of the task id), so a miss can be compared with it.
	Last Kind = "last"
)

// ErrMiss is returned by Get when an entry does not exist.
var ErrMiss = errors.New("cache miss")

// Storage is a cache backend. Keys are lowercase hex SHA-256 digests. A CAS
// Put must verify the content matches its digest; putting an existing digest
// is a success.
type Storage interface {
	Has(kind Kind, key string) (bool, error)
	Get(kind Kind, key string) (io.ReadCloser, error)
	Put(kind Kind, key string, content io.Reader) error
	Delete(kind Kind, key string) error
	// Walk visits every entry of a kind with its size and last use. Backends
	// that cannot enumerate return ErrNotSupported.
	Walk(kind Kind, visit func(key string, size int64, lastUsed time.Time) error) error
	// Touch records a use, so pruning keeps recently used entries.
	Touch(kind Kind, key string) error
}

// Layered is a Storage with a local tier in front of a remote one
// (remotecache.Tiered).
type Layered interface {
	// HasLocally reports whether the local tier holds an entry, without
	// asking the remote.
	HasLocally(kind Kind, key string) (bool, error)
}

// Source says which tier served a cache hit.
type Source string

// The tiers of a hit.
const (
	// SourceLocal is a hit this machine's cache held.
	SourceLocal Source = "local"
	// SourceRemote is a hit the remote cache served.
	SourceRemote Source = "remote"
)

// SourceOf says which tier will serve the action-cache entry key, asked
// before the lookup: an entry the local tier lacks comes from the remote.
//
// Parameters: s is the run's storage; key is the task's cache key.
// Returns: SourceRemote when s is Layered and its local tier lacks the
// entry, SourceLocal otherwise.
func SourceOf(s Storage, key string) Source {
	if l, ok := s.(Layered); ok {
		if has, err := l.HasLocally(AC, key); err == nil && !has {
			return SourceRemote
		}
	}
	return SourceLocal
}

// ErrNotSupported is returned by optional Storage operations.
var ErrNotSupported = errors.New("not supported by this storage backend")
