package cachemaintenance

import (
	"io"
	"sync/atomic"

	"github.com/monosmith-app/monosmith/apps/smith/taskcache"
)

// WriteMeter wraps a Storage and counts what a run stores in it: the bytes of
// the entries and blobs written, and how many were. A blob that was already
// there is not read, so it counts for nothing. Pointers to a task's last run
// are not results and are not counted.
type WriteMeter struct {
	taskcache.Storage
	bytes, puts atomic.Int64
}

// NewWriteMeter wraps storage.
func NewWriteMeter(storage taskcache.Storage) *WriteMeter {
	return &WriteMeter{Storage: storage}
}

// Put implements taskcache.Storage.
func (m *WriteMeter) Put(kind taskcache.Kind, key string, content io.Reader) error {
	if kind == taskcache.Last {
		return m.Storage.Put(kind, key, content)
	}
	err := m.Storage.Put(kind, key, &countingReader{r: content, n: &m.bytes})
	if err == nil {
		m.puts.Add(1)
	}
	return err
}

// Bytes is what the run has stored so far.
func (m *WriteMeter) Bytes() int64 { return m.bytes.Load() }

// Stored is how many entries and blobs the run has stored so far.
func (m *WriteMeter) Stored() int { return int(m.puts.Load()) }

type countingReader struct {
	r io.Reader
	n *atomic.Int64
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.n.Add(int64(n))
	return n, err
}
