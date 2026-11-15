package packagemanifests

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

// Built is a release's archives as the plugin built them.
type Built struct {
	// Assets are workspace-relative files; one is checksums.txt.
	Assets []string
	// Platforms maps each archive among Assets to its "os/arch".
	Platforms map[string]string
}

// Write renders the manifests from the built archives and their checksums
// into the archives' directory, and returns the files it wrote
// (workspace-relative) and the managers it skipped, with why.
func Write(root string, built Built, r Release, managers []string) (written, skipped []string, err error) {
	sums := map[string]string{}
	dir := ""
	for _, a := range built.Assets {
		if path.Base(a) != "checksums.txt" {
			continue
		}
		dir = path.Dir(a)
		f, err := os.Open(filepath.Join(root, filepath.FromSlash(a)))
		if err != nil {
			return nil, nil, err
		}
		scan := bufio.NewScanner(f)
		for scan.Scan() {
			if sum, name, ok := strings.Cut(strings.TrimSpace(scan.Text()), "  "); ok {
				sums[name] = sum
			}
		}
		_ = f.Close()
	}
	if dir == "" {
		return nil, managers, fmt.Errorf("no checksums.txt among the assets to describe them by")
	}
	for asset, platform := range built.Platforms {
		goos, goarch, _ := strings.Cut(platform, "/")
		name := path.Base(asset)
		if sum, ok := sums[name]; ok {
			r.Archives = append(r.Archives, Archive{Name: name, OS: goos, Arch: goarch, SHA256: sum})
		}
	}
	sort.Slice(r.Archives, func(i, j int) bool { return r.Archives[i].Name < r.Archives[j].Name })
	files, skipped := Render(r, managers)
	var names []string
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		rel := path.Join(dir, name)
		if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(rel)), []byte(files[name]), 0o644); err != nil {
			return written, skipped, err
		}
		written = append(written, rel)
	}
	return written, skipped, nil
}
