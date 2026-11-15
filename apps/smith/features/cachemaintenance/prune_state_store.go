package cachemaintenance

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

const (
	stateFile = "prune-state.json"
	lockFile  = "prune.lock"
	// staleLock is how old a lock must be before it is taken over: a prune of
	// even a very large store ends well within it.
	staleLock = 10 * time.Minute
)

// pruneState is the small record beside the store that lets a run decide
// without walking it: when the store was last measured, how big it was then,
// and how much runs have added since.
type pruneState struct {
	// LastCheck is when the store was last walked (measured or pruned).
	LastCheck time.Time `json:"lastCheck"`
	// LastPrune is when a prune last ran, even if it removed nothing.
	LastPrune time.Time `json:"lastPrune"`
	// SizeBytes is the size at LastCheck, after pruning.
	SizeBytes int64 `json:"sizeBytes"`
	// AddedBytes is what runs have stored since LastCheck.
	AddedBytes int64 `json:"addedBytes"`
}

// loadState reads the record; a missing or damaged one reads as empty, which
// makes the next check measure the store.
func loadState(root string) pruneState {
	var s pruneState
	raw, err := os.ReadFile(filepath.Join(root, stateFile))
	if err != nil || json.Unmarshal(raw, &s) != nil {
		return pruneState{}
	}
	return s
}

// saveState replaces the record atomically, so a reader never sees half of it.
func saveState(root string, s pruneState) error {
	raw, err := json.Marshal(s)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Join(root, "tmp"), "state-*")
	if err != nil {
		return err
	}
	_, werr := tmp.Write(raw)
	cerr := tmp.Close()
	if err := errors.Join(werr, cerr); err != nil {
		_ = os.Remove(tmp.Name())
		return err
	}
	if err := renameWithRetry(tmp.Name(), filepath.Join(root, stateFile)); err != nil {
		_ = os.Remove(tmp.Name())
		return err
	}
	return nil
}

// renameWithRetry renames over a file another smith process may be reading: on
// Windows that is refused for a moment ("access is denied").
func renameWithRetry(from, to string) error {
	var err error
	for attempt := 0; attempt < 8; attempt++ {
		if err = os.Rename(from, to); err == nil {
			return nil
		}
		time.Sleep(time.Duration(5<<attempt) * time.Millisecond)
	}
	return err
}

// lockPrune takes the right to prune automatically: one smith process prunes
// at a time, and the others skip. It returns false when another holds it.
func lockPrune(root string, now time.Time) (release func(), ok bool) {
	path := filepath.Join(root, lockFile)
	for attempt := 0; attempt < 2; attempt++ {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err == nil {
			_ = f.Close()
			return func() { _ = os.Remove(path) }, true
		}
		info, statErr := os.Stat(path)
		if errors.Is(statErr, fs.ErrNotExist) {
			continue
		}
		if statErr != nil || now.Sub(info.ModTime()) < staleLock {
			return nil, false
		}
		_ = os.Remove(path) // left by a process that died
	}
	return nil, false
}
