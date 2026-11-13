package functionhost

import (
	"archive/zip"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// zipFolder writes the content of a folder into a zip file, with the paths
// relative to the folder (the zip has no top folder). A link to a file is
// stored as the file; a link to a folder is skipped.
func zipFolder(folder, zipPath string) error {
	out, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	zw := zip.NewWriter(out)
	walkErr := filepath.WalkDir(folder, func(p string, d fs.DirEntry, err error) error {
		if err != nil || p == folder {
			return err
		}
		info, err := os.Stat(p)
		if err != nil || info.IsDir() {
			// A broken link has nothing to store, a folder is made by its files.
			return nil
		}
		rel, err := filepath.Rel(folder, p)
		if err != nil {
			return err
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(rel)
		header.Method = zip.Deflate
		w, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}
		in, err := os.Open(p)
		if err != nil {
			return err
		}
		defer func() { _ = in.Close() }()
		_, err = io.Copy(w, in)
		return err
	})
	if err := zw.Close(); err != nil && walkErr == nil {
		walkErr = err
	}
	if err := out.Close(); err != nil && walkErr == nil {
		walkErr = err
	}
	return walkErr
}
