package releasepack

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/githubrelease"
)

// Complete reports whether a release holds the package files of an ecosystem
// and the checksum file of every one: the state of a package whose attach
// finished. The checksum files are uploaded after all the package files, so a
// package file without its checksum file is an attach that stopped half way.
func Complete(r githubrelease.Release, e Ecosystem) bool {
	have := map[string]bool{}
	for _, a := range r.Assets {
		have[a.Name] = true
	}
	files := 0
	for _, a := range r.Assets {
		if e.IsPackageFile(a.Name) {
			files++
			if !have[a.Name+ChecksumSuffix] {
				return false
			}
		}
	}
	return files > 0
}

// Recover reads what a finished release holds: the package files with the
// digests their checksum files state. It lets a re-run list a package in the
// bundle without packing it again.
//
// Parameters: c and repo reach GitHub; r is the release; e the package's ecosystem.
// Returns: the files, in asset order.
// Errors: a checksum file cannot be downloaded or is not "<hex>  <name>".
func Recover(c *githubrelease.Client, repo githubrelease.Repository, r githubrelease.Release, e Ecosystem) ([]File, error) {
	dir, err := os.MkdirTemp("", "smith-checksums-")
	if err != nil {
		return nil, err
	}
	defer func() { _ = os.RemoveAll(dir) }()
	sizes := map[string]int64{}
	for _, a := range r.Assets {
		sizes[a.Name] = a.Size
	}
	var out []File
	for _, a := range r.Assets {
		if !strings.HasSuffix(a.Name, ChecksumSuffix) || !e.IsPackageFile(strings.TrimSuffix(a.Name, ChecksumSuffix)) {
			continue
		}
		to := filepath.Join(dir, "sum")
		if err := c.DownloadAsset(repo, a, to); err != nil {
			return nil, err
		}
		raw, err := os.ReadFile(to)
		if err != nil {
			return nil, err
		}
		digest, name, ok := strings.Cut(strings.TrimSpace(string(raw)), "  ")
		if !ok || len(digest) != 64 || name != strings.TrimSuffix(a.Name, ChecksumSuffix) {
			return nil, fmt.Errorf("%s is not a checksum file", a.Name)
		}
		out = append(out, File{Name: name, SHA256: digest, Size: sizes[name]})
	}
	return out, nil
}
