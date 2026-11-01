package sandbox

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

// statusEntry is one line of `git status --porcelain=v1 -z`.
type statusEntry struct {
	XY      string
	Path    string
	OldPath string
}

// parseStatus reads the output of `git status --porcelain=v1 -z
// --untracked-files=all`: entries end with NUL; a rename or copy is followed
// by the old path.
func parseStatus(raw string) []statusEntry {
	parts := strings.Split(raw, "\x00")
	var out []statusEntry
	for i := 0; i < len(parts); i++ {
		p := parts[i]
		if len(p) < 4 {
			continue
		}
		e := statusEntry{XY: p[:2], Path: p[3:]}
		if (e.XY[0] == 'R' || e.XY[0] == 'C') && i+1 < len(parts) {
			i++
			e.OldPath = parts[i]
		}
		out = append(out, e)
	}
	return out
}

// overlay brings the uncommitted state of the real tree into the copy: a file
// that exists in the real tree is copied over, one that does not is removed.
//
// Parameters: real and sandboxDir are the two roots; entries come from parseStatus.
// Returns an error when a file cannot be copied or removed.
func overlay(real, sandboxDir string, entries []statusEntry) error {
	for _, e := range entries {
		if e.OldPath != "" {
			if err := os.Remove(filepath.Join(sandboxDir, filepath.FromSlash(e.OldPath))); err != nil && !os.IsNotExist(err) {
				return err
			}
		}
		src := filepath.Join(real, filepath.FromSlash(e.Path))
		dst := filepath.Join(sandboxDir, filepath.FromSlash(e.Path))
		info, err := os.Stat(src)
		if err != nil {
			if os.IsNotExist(err) {
				if err := os.Remove(dst); err != nil && !os.IsNotExist(err) {
					return err
				}
				continue
			}
			return err
		}
		if info.IsDir() {
			continue
		}
		if err := copyFile(src, dst, info.Mode()); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(src, dst string, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode.Perm())
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}
