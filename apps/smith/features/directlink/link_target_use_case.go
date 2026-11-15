package directlink

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// expectedTarget is the program the entries of this version call: the link of the
// run folder on Windows, the installed program elsewhere. It makes nothing.
func expectedTarget(o Options) string {
	if o.GOOS != "windows" {
		return o.installed()
	}
	return filepath.Join(RunDir(o.Home), "smith-"+safeName(o.Version)+".exe")
}

// makeTarget returns the program the entries call. On Windows it is a hard
// link of the installed program in the run folder, named for the version, so
// every smith of that version shares one file and the installed program is
// never the one that runs; a copy replaces the link where the system cannot
// link (another drive). Elsewhere it is the installed program itself.
//
// Parameters: o says where smith is.
// Returns: the program to call; an error when the run folder cannot be written.
func makeTarget(o Options) (string, error) {
	if o.GOOS != "windows" {
		return o.installed(), nil
	}
	dir := RunDir(o.Home)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	target := expectedTarget(o)
	if installed := o.installed(); !sameFile(target, installed) {
		if err := replaceFile(target, func() error { return linkOrCopy(installed, target) }); err != nil {
			return "", err
		}
	}
	sweep(dir, target)
	return target, nil
}

// safeName keeps the characters a file name can hold.
func safeName(version string) string {
	return strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r == '.' || r == '-' {
			return r
		}
		return '_'
	}, version)
}

// replaceFile makes the file at path anew with make. A file that runs cannot be
// deleted on Windows, but it can be renamed: it is moved aside first, and the
// sweep removes it once its process has ended.
func replaceFile(path string, make func() error) error {
	if _, err := os.Lstat(path); err == nil {
		if os.Remove(path) != nil {
			aside := fmt.Sprintf("%s.old-%d", path, os.Getpid())
			if err := os.Rename(path, aside); err != nil {
				return err
			}
		}
	}
	return make()
}

// linkOrCopy makes path a hard link of from, or a copy when the system cannot link.
func linkOrCopy(from, path string) error {
	if os.Link(from, path) == nil {
		return nil
	}
	return copyFile(from, path)
}

// copyFile copies a file, keeping its mode.
func copyFile(from, to string) error {
	in, err := os.Open(from)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	info, err := in.Stat()
	if err != nil {
		return err
	}
	out, err := os.OpenFile(to, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode().Perm()|0o600)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}

// sweep removes the programs of other versions and the files moved aside, in
// the run folder. A program that still runs cannot be deleted: it stays until a
// later sweep, once its process has ended.
func sweep(dir, keep string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		path := filepath.Join(dir, e.Name())
		if e.IsDir() || path == keep || !strings.HasPrefix(e.Name(), "smith-") {
			continue
		}
		_ = os.Remove(path)
	}
}
