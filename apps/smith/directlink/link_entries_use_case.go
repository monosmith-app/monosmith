package directlink

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ErrNoHome is returned when there is no smith home to keep the link in.
var ErrNoHome = errors.New("there is no smith home to keep the link in")

// ErrReplaced is returned when smith runs from the link of an older install and
// the installed program has been replaced since: linking now would put the new
// program under the name of the old version.
var ErrReplaced = errors.New("the installed smith is not the one that runs; run smith link again in a new terminal")

// Link points every smith entry on PATH at the program, so the smith command
// starts it directly. It is safe to run again: an entry that already calls the
// program is left alone. An entry smith cannot change (a folder it may not
// write) is reported and left as it was; the others are still changed.
//
// Parameters: o says where smith is.
// Returns: what changed, what already was direct and what was skipped; an error
// when the run folder cannot be written.
func Link(o Options) (Report, error) {
	if o.Home == "" {
		return Report{}, ErrNoHome
	}
	if RunningFromLink(o.Home, o.Exe) && !sameFile(o.installed(), o.Exe) {
		return Report{}, ErrReplaced
	}
	target, err := makeTarget(o)
	if err != nil {
		return Report{}, err
	}
	report := Report{Target: target}
	var changed []Entry
	for _, e := range FindEntries(o.installed(), o.GOOS, RunDir(o.Home), target) {
		if e.Kind == Direct {
			report.Already = append(report.Already, e)
			continue
		}
		done, already, err := point(o, e, target)
		switch {
		case err != nil:
			report.Skipped = append(report.Skipped, fmt.Sprintf("%s: %v", e.Path, err))
		case already:
			report.Already = append(report.Already, e)
		default:
			changed = append(changed, done)
			report.Changed = append(report.Changed, done)
		}
	}
	now := time.Now
	if o.Now != nil {
		now = o.Now
	}
	// What an earlier link changed is kept, so undo restores it too.
	state := State{Version: o.Version, Installed: o.installed(), Target: target, Entries: changed, At: now()}
	if before, ok := LoadState(o.Home); ok {
		for _, e := range before.Entries {
			if !containsPath(state.Entries, e.Path) {
				state.Entries = append(state.Entries, e)
			}
		}
	}
	if err := saveState(o.Home, state); err != nil {
		return report, err
	}
	return report, nil
}

// containsPath is true when an entry with this path is in the list.
func containsPath(entries []Entry, path string) bool {
	for _, e := range entries {
		if e.Path == path {
			return true
		}
	}
	return false
}

// point makes one entry call target. already is true when it did.
func point(o Options, e Entry, target string) (done Entry, already bool, err error) {
	switch e.Kind {
	case CmdShim, Ps1Shim, ShShim:
		want := shimText(e.Kind, target)
		have, err := os.ReadFile(e.Path)
		if err != nil {
			return e, false, err
		}
		if string(have) == want {
			return e, true, nil
		}
		backup, err := keepOriginal(o.Home, e.Path, have)
		if err != nil {
			return e, false, err
		}
		if err := writeBeside(e.Path, []byte(want)); err != nil {
			return e, false, err
		}
		e.Backup = backup
		return e, false, nil
	case Symlink:
		old, err := os.Readlink(e.Path)
		if err != nil {
			return e, false, err
		}
		if err := os.Remove(e.Path); err != nil {
			return e, false, err
		}
		if err := os.Symlink(target, e.Path); err != nil {
			_ = os.Symlink(old, e.Path)
			return e, false, err
		}
		e.Backup = old
		return e, false, nil
	case Launcher:
		aside := e.Path + ".smith-old"
		_ = os.Remove(aside)
		if err := os.Rename(e.Path, aside); err != nil {
			return e, false, err
		}
		if err := linkOrCopy(target, e.Path); err != nil {
			_ = os.Rename(aside, e.Path)
			return e, false, err
		}
		e.Backup = aside
		return e, false, nil
	}
	return e, false, fmt.Errorf("a %s entry is not changed", e.Kind)
}

// keepOriginal stores the text of an entry before it is changed, so Undo can
// give it back. The first original is kept: a second link does not overwrite it
// with text smith wrote.
func keepOriginal(home, path string, text []byte) (string, error) {
	sum := sha1.Sum([]byte(path))
	backup := filepath.Join(BackupDir(home), hex.EncodeToString(sum[:6])+"-"+filepath.Base(path))
	if _, err := os.Stat(backup); err == nil {
		return backup, nil
	}
	if err := os.MkdirAll(filepath.Dir(backup), 0o755); err != nil {
		return "", err
	}
	return backup, os.WriteFile(backup, text, 0o644)
}

// writeBeside replaces a file with new text through a file next to it, so a
// reader never sees half of the text. The mode of the original is kept.
func writeBeside(path string, text []byte) error {
	mode := os.FileMode(0o644)
	if info, err := os.Stat(path); err == nil {
		mode = info.Mode().Perm()
	}
	tmp := path + ".smith-tmp"
	if err := os.WriteFile(tmp, text, mode); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
