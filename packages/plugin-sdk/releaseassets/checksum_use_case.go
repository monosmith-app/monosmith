package releaseassets

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ChecksumsFile is attached beside the archives; the installers verify
// downloads against it.
const ChecksumsFile = "checksums.txt"

// Digest is the hex SHA-256 of a file.
func Digest(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// WriteChecksums writes the SHA-256 of each archive (a file path) to the file
// checksums, one line each in the format `sha256sum -c` reads, named by the
// archive's file name.
func WriteChecksums(checksums string, archives ...string) error {
	var b strings.Builder
	for _, a := range archives {
		sum, err := Digest(a)
		if err != nil {
			return err
		}
		fmt.Fprintf(&b, "%s  %s\n", sum, filepath.Base(a))
	}
	return os.WriteFile(checksums, []byte(b.String()), 0o644)
}
