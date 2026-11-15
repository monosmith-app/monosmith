package taskcache

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"strings"
)

// ErrNoLastRun means a task has no cached run to compare with.
var ErrNoLastRun = errors.New("no cached run of this task")

// Remember records key as the task's most recent cached run.
func Remember(s Storage, task, key string) error {
	return s.Put(Last, lastKey(task), strings.NewReader(key))
}

// LastRun returns the task's most recent cached run and its key, or
// ErrNoLastRun when it never had one or the entry was pruned since.
func LastRun(s Storage, task string) (Entry, string, error) {
	r, err := s.Get(Last, lastKey(task))
	if errors.Is(err, ErrMiss) {
		return Entry{}, "", ErrNoLastRun
	}
	if err != nil {
		return Entry{}, "", err
	}
	raw, err := io.ReadAll(r)
	_ = r.Close()
	if err != nil {
		return Entry{}, "", err
	}
	key := strings.TrimSpace(string(raw))
	e, err := Lookup(s, key)
	if errors.Is(err, ErrMiss) {
		return Entry{}, "", ErrNoLastRun
	}
	return e, key, err
}

func lastKey(task string) string {
	sum := sha256.Sum256([]byte(task))
	return hex.EncodeToString(sum[:])
}
