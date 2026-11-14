package directlink

import (
	"fmt"
	"os"
)

// Undo gives every entry that Link changed back its original, and forgets the
// link. An entry the package manager has written again since is left as it is,
// as the original is no longer what is there to restore.
//
// Parameters: o says where smith is.
// Returns: the entries that were restored and the ones that were skipped with
// why; an error when the state cannot be removed.
func Undo(o Options) (restored []Entry, skipped []string, err error) {
	state, ok := LoadState(o.Home)
	if !ok {
		return nil, nil, nil
	}
	for _, e := range state.Entries {
		if e.Backup == "" {
			continue
		}
		if err := restore(e); err != nil {
			skipped = append(skipped, fmt.Sprintf("%s: %v", e.Path, err))
			continue
		}
		restored = append(restored, e)
	}
	// The links of the run folder go too, those that no program holds.
	sweep(RunDir(o.Home), "")
	return restored, skipped, forgetState(o.Home)
}

// restore gives one entry back its original.
func restore(e Entry) error {
	switch e.Kind {
	case CmdShim, Ps1Shim, ShShim:
		text, err := os.ReadFile(e.Backup)
		if err != nil {
			return err
		}
		return writeBeside(e.Path, text)
	case Symlink:
		if err := os.Remove(e.Path); err != nil {
			return err
		}
		return os.Symlink(e.Backup, e.Path)
	case Launcher:
		// The link is removed; a program that runs through it keeps running.
		if err := replaceFile(e.Path, func() error { return os.Rename(e.Backup, e.Path) }); err != nil {
			return err
		}
	}
	return nil
}
