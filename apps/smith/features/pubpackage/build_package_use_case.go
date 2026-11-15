package pubpackage

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Archive is one platform's release archive, holding the binary at its root.
type Archive struct {
	// Path is workspace-relative.
	Path     string
	OS, Arch string
}

// Release is an app's release as a pub.dev package.
type Release struct {
	// Name is the pub package.
	Name    string
	Binary  string
	Version string
	// Description (60 to 180 characters on pub.dev), Homepage and
	// Repository describe the package.
	Description, Homepage, Repository string
	// License is the workspace-relative license file; pub.dev requires one.
	License  string
	Archives []Archive
	// OutDir is the workspace-relative directory the package is written to.
	OutDir string
}

// Build writes the package: every archive's binary under
// lib/src/bin/<os>-<arch>/, the launcher, and the files pub.dev expects. It
// returns the package directory and the archives Dart does not run on.
func Build(root string, r Release) (dir string, skipped []string, err error) {
	dir = filepath.Join(root, filepath.FromSlash(r.OutDir))
	if err := os.RemoveAll(dir); err != nil {
		return "", nil, err
	}
	abis := map[string]string{}
	for _, a := range r.Archives {
		platform, abi, ok := Platform(a.OS, a.Arch)
		if !ok {
			skipped = append(skipped, a.OS+"/"+a.Arch)
			continue
		}
		name := r.Binary
		if a.OS == "windows" {
			name += ".exe"
		}
		to := filepath.Join(dir, "lib", "src", "bin", platform, name)
		if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
			return "", skipped, err
		}
		if err := extract(filepath.Join(root, filepath.FromSlash(a.Path)), name, to); err != nil {
			return "", skipped, err
		}
		abis[abi] = platform
	}
	if len(abis) == 0 {
		return "", skipped, fmt.Errorf("no archive for a platform Dart runs on")
	}
	license, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(r.License)))
	if err != nil {
		return "", skipped, fmt.Errorf("pub.dev needs a license: %w", err)
	}
	files := map[string]string{
		"pubspec.yaml":              Pubspec(r),
		"README.md":                 Readme(r),
		"CHANGELOG.md":              Changelog(r),
		"LICENSE":                   string(license),
		"bin/" + r.Binary + ".dart": Launcher(r, abis),
		"lib/" + r.Name + ".dart":   Library(r),
	}
	for rel, content := range files {
		path := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return "", skipped, err
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return "", skipped, err
		}
	}
	return dir, skipped, nil
}

// extract copies one file out of a zip archive, executable.
func extract(archive, name, to string) error {
	zr, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}
	defer func() { _ = zr.Close() }()
	for _, f := range zr.File {
		if f.Name != name {
			continue
		}
		in, err := f.Open()
		if err != nil {
			return err
		}
		defer func() { _ = in.Close() }()
		out, err := os.OpenFile(to, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, in); err != nil {
			_ = out.Close()
			return err
		}
		return out.Close()
	}
	return fmt.Errorf("%s has no %s", filepath.Base(archive), name)
}
