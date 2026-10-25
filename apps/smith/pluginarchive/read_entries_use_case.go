package pluginarchive

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// ZipEntries lists the files of a packaged plugin. It refuses anything that
// is not a plain file or a directory (a symbolic link, a device), because
// nothing a plugin needs is one. Directory entries carry no content and are
// left out. It opens nothing and runs nothing.
func ZipEntries(r io.ReaderAt, size int64) ([]Entry, error) {
	zr, err := zip.NewReader(r, size)
	if err != nil {
		return nil, fmt.Errorf("not a plugin archive: %w", err)
	}
	var out []Entry
	for _, f := range zr.File {
		mode := f.Mode()
		if mode&fs.ModeSymlink != 0 || (!mode.IsRegular() && !mode.IsDir()) {
			return nil, fmt.Errorf("entry %q is a link or special file (%s); a plugin archive holds plain files only", f.Name, mode.Type())
		}
		if mode.IsDir() || strings.HasSuffix(f.Name, "/") {
			if err := ValidateName(strings.TrimSuffix(f.Name, "/")); err != nil {
				return nil, err
			}
			continue
		}
		f := f
		out = append(out, Entry{
			Name:       f.Name,
			Executable: mode&0o111 != 0,
			Size:       int64(f.UncompressedSize64),
			Open:       func() (io.ReadCloser, error) { return f.Open() },
		})
	}
	return out, nil
}

// DirEntries lists the regular files under dir, in name order. skip may
// leave a file or a whole directory out. A symbolic link or special file is
// an error: a packed plugin must not point outside itself.
func DirEntries(dir string, skip func(rel string, isDir bool) bool) ([]Entry, error) {
	var out []Entry
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == dir {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		name := filepath.ToSlash(rel)
		if skip != nil && skip(name, d.IsDir()) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("%s is a link or special file; a plugin directory must hold plain files only", name)
		}
		path = filepath.Clean(path)
		out = append(out, Entry{
			Name:       name,
			Executable: info.Mode()&0o111 != 0,
			Size:       info.Size(),
			Open:       func() (io.ReadCloser, error) { return os.Open(path) },
		})
		return nil
	})
	return out, err
}
