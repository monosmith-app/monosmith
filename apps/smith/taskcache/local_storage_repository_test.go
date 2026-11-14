package taskcache

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"io/fs"
	"testing"
	"time"
)

func digest(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// StorageConformance is the behaviour every Storage backend must have.
func storageConformance(t *testing.T, s Storage) {
	t.Helper()
	blob := []byte("hello")
	key := digest(blob)
	if ok, _ := s.Has(CAS, key); ok {
		t.Fatal("empty storage has the key")
	}
	if _, err := s.Get(CAS, key); !errors.Is(err, ErrMiss) {
		t.Fatalf("want ErrMiss, got %v", err)
	}
	if err := s.Put(CAS, key, bytes.NewReader(blob)); err != nil {
		t.Fatal(err)
	}
	if err := s.Put(CAS, key, bytes.NewReader(blob)); err != nil {
		t.Fatal("putting an existing blob must succeed: ", err)
	}
	if err := s.Put(CAS, digest([]byte("other")), bytes.NewReader(blob)); err == nil {
		t.Fatal("a blob that does not match its digest must be rejected")
	}
	r, err := s.Get(CAS, key)
	if err != nil {
		t.Fatal(err)
	}
	got, _ := io.ReadAll(r)
	_ = r.Close()
	if !bytes.Equal(got, blob) {
		t.Fatalf("got %q", got)
	}
	if err := s.Put(AC, key, bytes.NewReader([]byte("{}"))); err != nil {
		t.Fatal(err)
	}
	if err := s.Touch(AC, key); err != nil {
		t.Fatal(err)
	}
	count := 0
	err = s.Walk(CAS, func(k string, size int64, last time.Time) error {
		count++
		if k != key || size != int64(len(blob)) || last.IsZero() {
			t.Errorf("walk saw %s %d %v", k, size, last)
		}
		return nil
	})
	if err != nil && !errors.Is(err, ErrNotSupported) {
		t.Fatal(err)
	}
	if err := s.Delete(CAS, key); err != nil {
		t.Fatal(err)
	}
	if ok, _ := s.Has(CAS, key); ok {
		t.Fatal("deleted key still present")
	}
	if err := s.Delete(CAS, key); err != nil {
		t.Fatal("deleting a missing key must succeed")
	}
	if _, err := s.Get(CAS, "../escape"); err == nil {
		t.Fatal("path traversal must be rejected")
	}
}

func TestLocalStorageConformance(t *testing.T) {
	s, err := NewLocalStorage(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	storageConformance(t, s)
}

func TestWindowsRefusalsOfAFileInUseAreRecognised(t *testing.T) {
	denied := &fs.PathError{Op: "open", Path: "x", Err: fs.ErrPermission}
	sharing := &fs.PathError{Op: "open", Path: "x", Err: errSharingViolation}
	missing := &fs.PathError{Op: "open", Path: "x", Err: fs.ErrNotExist}
	if !refusedWhileInUse(denied) || !refusedWhileInUse(sharing) {
		t.Error("a file another process is deleting or holds was not recognised")
	}
	if refusedWhileInUse(missing) || refusedWhileInUse(nil) {
		t.Error("an absent file or no error was taken for a refusal")
	}
}
