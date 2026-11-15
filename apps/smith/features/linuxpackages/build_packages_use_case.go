package linuxpackages

import (
	"archive/zip"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

// Release is an app's release as Linux packages.
type Release struct {
	// Kinds are the formats asked for: deb, rpm.
	Kinds []string
	// Product names the packages; Binary is the executable in each archive.
	Product, Binary string
	Version         string
	Summary         string
	Homepage        string
	License         string
	Maintainer      string
}

// Built is a release's archives as the plugin built them.
type Built struct {
	// Assets are workspace-relative files; checksums.txt among them gets a
	// line for each package.
	Assets []string
	// Platforms maps each archive among Assets to its "os/arch".
	Platforms map[string]string
}

// Expected are the package file names a release's platforms give, so a
// resumed release can tell whether its GitHub Release has them all.
func Expected(r Release, platforms []string) []string {
	var names []string
	for _, platform := range platforms {
		goos, goarch, _ := strings.Cut(platform, "/")
		if goos != "linux" || !Supported(goarch) {
			continue
		}
		for _, kind := range r.Kinds {
			if name, err := FileName(kind, r.pkg(goarch)); err == nil {
				names = append(names, name)
			}
		}
	}
	sort.Strings(names)
	return names
}

// Build writes a package of each kind for every linux archive beside the
// archives, adds them to checksums.txt, and returns them (workspace-relative).
func Build(root string, built Built, r Release) ([]string, error) {
	var archives []string
	for asset := range built.Platforms {
		archives = append(archives, asset)
	}
	sort.Strings(archives)
	var written []string
	for _, asset := range archives {
		goos, goarch, _ := strings.Cut(built.Platforms[asset], "/")
		if goos != "linux" || !Supported(goarch) {
			continue
		}
		binary, err := read(filepath.Join(root, filepath.FromSlash(asset)), r.Binary)
		if err != nil {
			return written, err
		}
		p := r.pkg(goarch)
		for _, kind := range r.Kinds {
			name, err := FileName(kind, p)
			if err != nil {
				return written, err
			}
			var content []byte
			if kind == "deb" {
				content, err = Deb(p, binary)
			} else {
				content, err = RPM(p, binary)
			}
			if err != nil {
				return written, fmt.Errorf("%s: %w", name, err)
			}
			rel := path.Join(path.Dir(asset), name)
			if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(rel)), content, 0o644); err != nil {
				return written, err
			}
			written = append(written, rel)
		}
	}
	if len(written) == 0 {
		return nil, fmt.Errorf("no linux/amd64 or linux/arm64 archive to package")
	}
	return written, addChecksums(root, built.Assets, written)
}

func (r Release) pkg(goarch string) Package {
	return Package{Name: r.Product, Version: r.Version, Arch: goarch, Summary: r.Summary, Homepage: r.Homepage, License: r.License, Maintainer: r.Maintainer, Binary: r.Binary}
}

// addChecksums appends each package's SHA-256 to the release's checksums.txt,
// in its "<sum>  <name>" format, unless it is listed already.
func addChecksums(root string, assets, packages []string) error {
	var sums string
	for _, a := range assets {
		if path.Base(a) == "checksums.txt" {
			sums = filepath.Join(root, filepath.FromSlash(a))
		}
	}
	if sums == "" {
		return nil
	}
	existing, err := os.ReadFile(sums)
	if err != nil {
		return err
	}
	text := string(existing)
	for _, p := range packages {
		name := path.Base(p)
		if strings.Contains(text, "  "+name+"\n") {
			continue
		}
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(p)))
		if err != nil {
			return err
		}
		if text != "" && !strings.HasSuffix(text, "\n") {
			text += "\n"
		}
		text += fmt.Sprintf("%x  %s\n", sha256.Sum256(content), name)
	}
	return os.WriteFile(sums, []byte(text), 0o644)
}

// read returns the binary at the root of a zip archive.
func read(archive, name string) ([]byte, error) {
	zr, err := zip.OpenReader(archive)
	if err != nil {
		return nil, err
	}
	defer func() { _ = zr.Close() }()
	for _, f := range zr.File {
		if f.Name != name {
			continue
		}
		in, err := f.Open()
		if err != nil {
			return nil, err
		}
		defer func() { _ = in.Close() }()
		return io.ReadAll(in)
	}
	return nil, fmt.Errorf("%s has no %s", filepath.Base(archive), name)
}
