package pluginarchive

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"sort"
	"time"
)

// stamp is the modification time every entry gets, so the same files always
// pack to the same bytes.
var stamp = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

// Write packs entries into a zip on w: entries sorted by name with a fixed
// timestamp and mode, then ChecksumsFile. The entries must not already
// include a checksums file. It returns the digests it recorded.
func Write(w io.Writer, entries []Entry) (map[string]string, error) {
	sorted := append([]Entry(nil), entries...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })
	zw := zip.NewWriter(w)
	sums := map[string]string{}
	for _, e := range sorted {
		if err := ValidateName(e.Name); err != nil {
			return nil, err
		}
		if e.Name == ChecksumsFile {
			return nil, fmt.Errorf("%s is written by smith; remove it from the plugin's files", ChecksumsFile)
		}
		if _, dup := sums[e.Name]; dup {
			return nil, fmt.Errorf("%s appears twice", e.Name)
		}
		sum, err := writeEntry(zw, e)
		if err != nil {
			return nil, err
		}
		sums[e.Name] = sum
	}
	dst, err := zw.CreateHeader(header(ChecksumsFile, false))
	if err != nil {
		return nil, err
	}
	if _, err := dst.Write(FormatChecksums(sums)); err != nil {
		return nil, err
	}
	return sums, zw.Close()
}

func header(name string, executable bool) *zip.FileHeader {
	h := &zip.FileHeader{Name: name, Method: zip.Deflate, Modified: stamp}
	if executable {
		h.SetMode(0o755)
	} else {
		h.SetMode(0o644)
	}
	return h
}

func writeEntry(zw *zip.Writer, e Entry) (string, error) {
	src, err := e.Open()
	if err != nil {
		return "", err
	}
	defer func() { _ = src.Close() }()
	dst, err := zw.CreateHeader(header(e.Name, e.Executable))
	if err != nil {
		return "", err
	}
	h := sha256.New()
	if _, err := io.Copy(io.MultiWriter(dst, h), src); err != nil {
		return "", fmt.Errorf("packing %s: %w", e.Name, err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
