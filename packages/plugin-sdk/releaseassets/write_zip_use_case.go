package releaseassets

import (
	"archive/zip"
	"io"
	"os"
)

// Entry is one file of an archive: where it is on disk, its name inside the
// archive, and whether it is a program.
type Entry struct {
	Path       string
	Name       string
	Executable bool
}

// WriteZip writes the entries into a zip archive, programs marked executable.
func WriteZip(archive string, entries []Entry) error {
	out, err := os.Create(archive)
	if err != nil {
		return err
	}
	zw := zip.NewWriter(out)
	for _, e := range entries {
		if err = addEntry(zw, e); err != nil {
			break
		}
	}
	if cerr := zw.Close(); err == nil {
		err = cerr
	}
	if cerr := out.Close(); err == nil {
		err = cerr
	}
	return err
}

func addEntry(zw *zip.Writer, e Entry) error {
	in, err := os.Open(e.Path)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	header := &zip.FileHeader{Name: e.Name, Method: zip.Deflate}
	if e.Executable {
		header.SetMode(0o755)
	} else {
		header.SetMode(0o644)
	}
	w, err := zw.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, in)
	return err
}
