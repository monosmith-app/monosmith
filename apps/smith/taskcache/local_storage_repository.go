package taskcache

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LocalStorage keeps the cache in a directory:
//
//	<dir>/v1/ac/<aa>/<key>     action-cache entries (JSON)
//	<dir>/v1/cas/<aa>/<key>    blobs, named by the SHA-256 of their bytes
//	<dir>/v1/tmp/              same volume, so writes land by rename
type LocalStorage struct {
	root string
}

// NewLocalStorage opens (and creates) a cache directory.
func NewLocalStorage(dir string) (*LocalStorage, error) {
	root := filepath.Join(dir, "v1")
	for _, sub := range []string{"ac", "cas", "tmp"} {
		if err := os.MkdirAll(filepath.Join(root, sub), 0o755); err != nil {
			return nil, err
		}
	}
	return &LocalStorage{root: root}, nil
}

// Root is the versioned directory that holds ac/, cas/, last/ and tmp/.
func (s *LocalStorage) Root() string { return s.root }

func (s *LocalStorage) path(kind Kind, key string) (string, error) {
	if len(key) < 3 || strings.ContainsAny(key, `/\.`) {
		return "", fmt.Errorf("invalid cache key %q", key)
	}
	return filepath.Join(s.root, string(kind), key[:2], key), nil
}

// Has implements Storage.
func (s *LocalStorage) Has(kind Kind, key string) (bool, error) {
	p, err := s.path(kind, key)
	if err != nil {
		return false, err
	}
	_, err = os.Stat(p)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return err == nil, err
}

// Get implements Storage.
func (s *LocalStorage) Get(kind Kind, key string) (io.ReadCloser, error) {
	p, err := s.path(kind, key)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(p)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, ErrMiss
	}
	return f, err
}

// Put implements Storage: it writes to tmp/, verifies a CAS digest, then
// renames into place, retrying when another process (an antivirus scanner, an
// indexer) briefly holds the file on Windows.
func (s *LocalStorage) Put(kind Kind, key string, content io.Reader) error {
	p, err := s.path(kind, key)
	if err != nil {
		return err
	}
	if kind == CAS {
		if ok, _ := s.Has(kind, key); ok {
			// A use: a prune running beside this process must not sweep a blob
			// that an entry written next is about to reference.
			_ = s.Touch(kind, key)
			return nil
		}
	}
	tmp, err := os.CreateTemp(filepath.Join(s.root, "tmp"), "put-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	hash := sha256.New()
	if _, err := io.Copy(io.MultiWriter(tmp, hash), content); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if kind == CAS && hex.EncodeToString(hash.Sum(nil)) != key {
		return fmt.Errorf("cache blob does not match its digest %s", key)
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return retry(func() error { return os.Rename(tmp.Name(), p) })
}

// Delete implements Storage.
func (s *LocalStorage) Delete(kind Kind, key string) error {
	p, err := s.path(kind, key)
	if err != nil {
		return err
	}
	err = retry(func() error { return os.Remove(p) })
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	return err
}

// Walk implements Storage. Last use is the file's modification time, which
// Touch refreshes (access times are unreliable on Windows and noatime mounts).
func (s *LocalStorage) Walk(kind Kind, visit func(key string, size int64, lastUsed time.Time) error) error {
	return filepath.WalkDir(filepath.Join(s.root, string(kind)), func(path string, d fs.DirEntry, err error) error {
		if errors.Is(err, fs.ErrNotExist) {
			return nil // removed concurrently (a prune, a clear)
		}
		if err != nil || d.IsDir() {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return nil // removed concurrently
		}
		return visit(d.Name(), info.Size(), info.ModTime())
	})
}

// Touch implements Storage.
func (s *LocalStorage) Touch(kind Kind, key string) error {
	p, err := s.path(kind, key)
	if err != nil {
		return err
	}
	now := time.Now()
	return os.Chtimes(p, now, now)
}

func retry(op func() error) error {
	var err error
	for attempt := 0; attempt < 6; attempt++ {
		if err = op(); err == nil || errors.Is(err, fs.ErrNotExist) {
			return err
		}
		time.Sleep(time.Duration(10<<attempt) * time.Millisecond)
	}
	return err
}
