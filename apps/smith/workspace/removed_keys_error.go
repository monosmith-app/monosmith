package workspace

import (
	"errors"
	"strings"
)

// RemovedKeysError says smith.toml still uses a table smith renamed. It is the
// same kind of message as NewerSmithError: one sentence and the one command
// that fixes it, because the old file is not wrong, only old.
type RemovedKeysError struct {
	// Keys are the renamed tables the file uses.
	Keys []RemovedKey
}

// Error returns the sentence smith prints.
func (e *RemovedKeysError) Error() string {
	parts := make([]string, len(e.Keys))
	for i, k := range e.Keys {
		parts[i] = k.Display(k.Old) + ", which is now " + k.Display(k.New)
	}
	return FileName + " uses " + strings.Join(parts, " and uses ") + ": run smith doctor --fix"
}

// AsRemovedKeys returns the RemovedKeysError inside err, if there is one.
func AsRemovedKeys(err error) (*RemovedKeysError, bool) {
	var removed *RemovedKeysError
	ok := errors.As(err, &removed)
	return removed, ok
}
