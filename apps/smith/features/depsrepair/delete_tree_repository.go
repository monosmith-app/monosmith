package depsrepair

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// DeleteTree deletes the file or the folder at path with everything in it.
//
// A link is deleted as a link. A symlink or a Windows junction, at path or
// anywhere below it, is removed itself and is never followed: the folder it
// points to keeps every file. A package manager links workspace packages into
// its install folder (node_modules/@scope/lib is a junction to libs/lib on
// Windows), so following one would delete source.
//
// Parameters: ctx stops the deletion between two entries; path is absolute.
//
// Returns: nil when path is gone, also when nothing was there.
//
// Errors: the first entry that could not be read or removed (a file another
// program holds open), or ctx's error.
func DeleteTree(ctx context.Context, path string) error {
	info, err := os.Lstat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	return deleteEntry(ctx, path, info.Mode())
}

// deleteEntry removes one entry whose mode was read without following it.
// Only a real folder is walked: everything else, a link to a folder included,
// is removed as the one entry it is.
func deleteEntry(ctx context.Context, path string, mode fs.FileMode) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if !mode.IsDir() || isLink(mode) {
		return removeOne(path)
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	for _, e := range entries {
		// The type of a directory entry is read from the folder, not from the
		// target of a link, so a link never looks like a folder here.
		if err := deleteEntry(ctx, filepath.Join(path, e.Name()), e.Type()); err != nil {
			return err
		}
	}
	return removeOne(path)
}

// removeOne removes a file, a link or an empty folder. Something that is
// already gone is not an error: another program may remove it first.
func removeOne(path string) error {
	err := os.Remove(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	return err
}
