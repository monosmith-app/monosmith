package releaseassets

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

// stageAssets copies files into dir under their upload names, with a checksums
// file named checksumsName beside them when it is not empty, and returns what
// to upload, sorted. dir is emptied first, so nothing of an earlier run is
// uploaded.
func stageAssets(dir string, files []string, names map[string]string, checksumsName string) ([]string, error) {
	if err := os.RemoveAll(dir); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	var staged, lines []string
	files = append([]string(nil), files...)
	sort.Slice(files, func(i, j int) bool { return names[files[i]] < names[files[j]] })
	for _, f := range files {
		to := filepath.Join(dir, names[f])
		sum, err := copyWithSum(f, to)
		if err != nil {
			return nil, err
		}
		staged = append(staged, to)
		lines = append(lines, sum+"  "+names[f])
	}
	if checksumsName != "" {
		path := filepath.Join(dir, checksumsName)
		if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
			return nil, err
		}
		staged = append(staged, path)
	}
	sort.Strings(staged)
	return staged, nil
}

func copyWithSum(from, to string) (string, error) {
	in, err := os.Open(from)
	if err != nil {
		return "", err
	}
	defer func() { _ = in.Close() }()
	out, err := os.Create(to)
	if err != nil {
		return "", err
	}
	h := sha256.New()
	if _, err := io.Copy(io.MultiWriter(out, h), in); err != nil {
		_ = out.Close()
		return "", fmt.Errorf("copying %s: %w", from, err)
	}
	return hex.EncodeToString(h.Sum(nil)), out.Close()
}
