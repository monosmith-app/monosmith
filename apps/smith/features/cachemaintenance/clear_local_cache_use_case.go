package cachemaintenance

import (
	"errors"
	"io/fs"
	"os"
	"time"
)

// Clear deletes the local cache: the versioned store under dir (entries, blobs,
// last-run pointers and the prune record) and nothing else in dir, so a
// [cache] dir that points at a shared folder loses only what smith put there.
// The remote tier is not touched. It reports what the store held.
func Clear(dir string) (ClearResult, error) {
	result := ClearResult{Dir: dir}
	status, err := Inspect(dir, 0)
	if err != nil || !status.Exists {
		return result, err
	}
	store, _, err := openExisting(dir)
	if err != nil {
		return result, err
	}
	result.Entries, result.Blobs, result.Bytes = status.Entries, status.Blobs, status.Bytes
	err = retryRemove(store.Root())
	return result, err
}

// retryRemove removes a directory tree, trying again for a moment when another
// process (an antivirus scanner, a smith reading a blob) holds a file on Windows.
func retryRemove(path string) error {
	var err error
	for attempt := 0; attempt < 6; attempt++ {
		if err = os.RemoveAll(path); err == nil || errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		time.Sleep(time.Duration(10<<attempt) * time.Millisecond)
	}
	return err
}
