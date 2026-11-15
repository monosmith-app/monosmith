package pluginarchive

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// maxChecksumsBytes bounds the checksums file itself.
const maxChecksumsBytes = 4 << 20

// Extract writes entries into dest, which must be a directory the caller just
// made for this purpose. Before anything is written it checks every name, the
// number and declared sizes of the files, and, when the entries include
// ChecksumsFile, that it lists exactly the other files. While writing it
// checks each file's real size against the limits and its SHA-256 against the
// list, so a file that lies about its size or content stops the extraction.
// With requireChecksums, an entry set without ChecksumsFile is refused.
// Files are created with mode 0644, or 0755 when executable: no setuid bit,
// no link, no device ever reaches the disk. Nothing is executed.
func Extract(entries []Entry, dest string, limits Limits, requireChecksums bool) (Extraction, error) {
	limits = limits.orDefault()
	byName, checksums, err := inspect(entries, limits)
	if err != nil {
		return Extraction{}, err
	}
	if _, ok := byName[ManifestFile]; !ok {
		return Extraction{}, fmt.Errorf("the plugin has no %s at its root", ManifestFile)
	}
	var sums map[string]string
	if checksums != nil {
		raw, err := readAll(*checksums, maxChecksumsBytes)
		if err != nil {
			return Extraction{}, err
		}
		if sums, err = ParseChecksums(raw); err != nil {
			return Extraction{}, err
		}
		if err := sameSet(byName, sums); err != nil {
			return Extraction{}, err
		}
	} else if requireChecksums {
		return Extraction{}, fmt.Errorf("the archive has no %s, so nothing can be verified", ChecksumsFile)
	}

	names := make([]string, 0, len(byName))
	for n := range byName {
		names = append(names, n)
	}
	sort.Strings(names)
	var total int64
	for _, n := range names {
		written, err := extractOne(byName[n], dest, sums, limits, limits.MaxTotalBytes-total)
		if err != nil {
			return Extraction{}, err
		}
		total += written
	}
	if checksums != nil {
		raw, err := readAll(*checksums, maxChecksumsBytes)
		if err != nil {
			return Extraction{}, err
		}
		if err := os.WriteFile(filepath.Join(dest, ChecksumsFile), raw, 0o644); err != nil {
			return Extraction{}, err
		}
	}
	return Extraction{Files: len(names), Bytes: total, Checksummed: checksums != nil}, nil
}

// inspect validates the whole set and returns the entries by name, without
// the checksums file, which is returned on its own (nil when absent).
func inspect(entries []Entry, limits Limits) (map[string]Entry, *Entry, error) {
	if len(entries) > limits.MaxFiles {
		return nil, nil, fmt.Errorf("the archive holds %d files; the limit is %d", len(entries), limits.MaxFiles)
	}
	byName := make(map[string]Entry, len(entries))
	folded := map[string]string{}
	var checksums *Entry
	var declared int64
	for i := range entries {
		e := entries[i]
		if err := ValidateName(e.Name); err != nil {
			return nil, nil, err
		}
		key := strings.ToLower(e.Name)
		if other, dup := folded[key]; dup {
			return nil, nil, fmt.Errorf("entries %q and %q differ only in case, which a case-insensitive disk cannot hold", other, e.Name)
		}
		folded[key] = e.Name
		if e.Size < 0 || e.Size > limits.MaxFileBytes {
			return nil, nil, fmt.Errorf("%s is %d bytes; the limit for one file is %d", e.Name, e.Size, limits.MaxFileBytes)
		}
		if declared += e.Size; declared > limits.MaxTotalBytes {
			return nil, nil, fmt.Errorf("the archive expands to more than %d bytes", limits.MaxTotalBytes)
		}
		if e.Name == ChecksumsFile {
			checksums = &entries[i]
			continue
		}
		byName[e.Name] = e
	}
	return byName, checksums, nil
}

func sameSet(files map[string]Entry, sums map[string]string) error {
	var problems []string
	for n := range files {
		if _, ok := sums[n]; !ok {
			problems = append(problems, n+" is in the archive but not in "+ChecksumsFile)
		}
	}
	for n := range sums {
		if _, ok := files[n]; !ok {
			problems = append(problems, n+" is in "+ChecksumsFile+" but not in the archive")
		}
	}
	if len(problems) == 0 {
		return nil
	}
	sort.Strings(problems)
	return fmt.Errorf("the archive does not match its %s: %s", ChecksumsFile, strings.Join(problems, "; "))
}

func extractOne(e Entry, dest string, sums map[string]string, limits Limits, budget int64) (int64, error) {
	target := filepath.Join(dest, filepath.FromSlash(e.Name))
	if rel, err := filepath.Rel(dest, target); err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return 0, fmt.Errorf("entry %q escapes the install directory", e.Name)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return 0, err
	}
	src, err := e.Open()
	if err != nil {
		return 0, err
	}
	defer func() { _ = src.Close() }()
	mode := os.FileMode(0o644)
	if e.Executable {
		mode = 0o755
	}
	out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
	if err != nil {
		return 0, fmt.Errorf("writing %s: %w", e.Name, err)
	}
	limit := limits.MaxFileBytes
	if budget < limit {
		limit = budget
	}
	h := sha256.New()
	n, copyErr := io.Copy(io.MultiWriter(out, h), io.LimitReader(src, limit+1))
	closeErr := out.Close()
	switch {
	case copyErr != nil:
		return 0, fmt.Errorf("reading %s: %w", e.Name, copyErr)
	case closeErr != nil:
		return 0, closeErr
	case n > limit:
		return 0, fmt.Errorf("%s is larger than the size limit", e.Name)
	}
	if sums != nil {
		if got := hex.EncodeToString(h.Sum(nil)); got != sums[e.Name] {
			return 0, fmt.Errorf("digest mismatch for %s: %s lists sha256 %s, the file is %s", e.Name, ChecksumsFile, sums[e.Name], got)
		}
	}
	// A file written with a stricter umask still gets its intended mode.
	_ = os.Chmod(target, mode)
	return n, nil
}

func readAll(e Entry, limit int64) ([]byte, error) {
	src, err := e.Open()
	if err != nil {
		return nil, err
	}
	defer func() { _ = src.Close() }()
	raw, err := io.ReadAll(io.LimitReader(src, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(raw)) > limit {
		return nil, fmt.Errorf("%s is too large", e.Name)
	}
	return raw, nil
}
