package taskcache

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"time"
)

// Save stores a successful run: every file under the task's output
// directories, its log and the inputs of its key, and remembers it as the
// task's last cached run. Outputs must be directories (a bare file path is
// rejected) so restoring can replace them wholesale.
func Save(s Storage, root, key, task, projectRoot string, outputs []string, log []byte, took time.Duration, inputs *KeyInputs) error {
	entry := Entry{Schema: EntrySchema, Task: task, Created: time.Now().UTC(), Duration: took, Inputs: inputs}
	for _, o := range outputs {
		dir := filepath.Join(root, filepath.FromSlash(path.Join(projectRoot, o)))
		info, err := os.Stat(dir)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return errors.New(task + ": output " + o + " is a file; declare its directory instead")
		}
		err = filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			info, err := d.Info()
			if err != nil {
				return err
			}
			digest, err := putFile(s, p)
			if err != nil {
				return err
			}
			rel, _ := filepath.Rel(root, p)
			entry.Outputs = append(entry.Outputs, OutputFile{Path: filepath.ToSlash(rel), Digest: digest, Mode: uint32(info.Mode().Perm())})
			return nil
		})
		if err != nil {
			return err
		}
	}
	logDigest := sha256.Sum256(log)
	entry.Log = hex.EncodeToString(logDigest[:])
	if err := s.Put(CAS, entry.Log, bytes.NewReader(log)); err != nil {
		return err
	}
	raw, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	if err := s.Put(AC, key, bytes.NewReader(raw)); err != nil {
		return err
	}
	return Remember(s, task, key)
}

// Lookup returns the entry for a key, or ErrMiss. A corrupt entry, or one
// whose blobs are gone, is a miss.
func Lookup(s Storage, key string) (Entry, error) {
	r, err := s.Get(AC, key)
	if err != nil {
		return Entry{}, err
	}
	defer func() { _ = r.Close() }()
	var e Entry
	if err := json.NewDecoder(r).Decode(&e); err != nil || e.Schema != EntrySchema {
		return Entry{}, ErrMiss
	}
	for _, o := range append([]OutputFile{{Digest: e.Log}}, e.Outputs...) {
		if ok, _ := s.Has(CAS, o.Digest); !ok {
			return Entry{}, ErrMiss
		}
	}
	_ = s.Touch(AC, key)
	return e, nil
}

// Restore replaces the task's output directories with the cached files and
// returns the cached log.
func Restore(s Storage, root, projectRoot string, outputs []string, e Entry) ([]byte, error) {
	for _, o := range outputs {
		dir := filepath.Join(root, filepath.FromSlash(path.Join(projectRoot, o)))
		if err := retry(func() error { return os.RemoveAll(dir) }); err != nil {
			return nil, err
		}
	}
	for _, f := range e.Outputs {
		dest := filepath.Join(root, filepath.FromSlash(f.Path))
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return nil, err
		}
		if err := copyBlob(s, f.Digest, dest, fs.FileMode(f.Mode)); err != nil {
			return nil, err
		}
	}
	r, err := s.Get(CAS, e.Log)
	if err != nil {
		return nil, err
	}
	defer func() { _ = r.Close() }()
	return io.ReadAll(r)
}

func putFile(s Storage, p string) (string, error) {
	f, err := os.Open(p)
	if err != nil {
		return "", err
	}
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		_ = f.Close()
		return "", err
	}
	digest := hex.EncodeToString(h.Sum(nil))
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		_ = f.Close()
		return "", err
	}
	err = s.Put(CAS, digest, f)
	_ = f.Close()
	return digest, err
}

func copyBlob(s Storage, digest, dest string, mode fs.FileMode) error {
	r, err := s.Get(CAS, digest)
	if err != nil {
		return err
	}
	defer func() { _ = r.Close() }()
	if mode == 0 {
		mode = 0o644
	}
	f, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, r); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}
