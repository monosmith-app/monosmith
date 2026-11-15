package rubygems

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Archive is one platform's release archive, holding the binary at its root.
type Archive struct {
	// Path is workspace-relative.
	Path     string
	OS, Arch string
}

// Release is an app's release as platform gems.
type Release struct {
	// Name is the gem.
	Name    string
	Binary  string
	Version string
	// Description, License, Homepage and Repository describe the gem on
	// rubygems.org.
	Description, License, Homepage, Repository string
	Archives                                   []Archive
	// OutDir is the workspace-relative directory the gems are written to.
	OutDir string
}

// Gem is one built platform gem.
type Gem struct {
	// File is the gem's absolute path.
	File string
}

// Build writes one gem per platform each archive runs on into OutDir. A gem
// holds the binary in libexec/ and a Ruby launcher in exe/ that runs it:
// RubyGems puts the launcher on PATH, and nothing is downloaded after
// install. Archives with no gem platform are returned as skipped.
func Build(root string, r Release) (gems []Gem, skipped []string, err error) {
	version, err := Version(r.Version)
	if err != nil {
		return nil, nil, err
	}
	out := filepath.Join(root, filepath.FromSlash(r.OutDir))
	if err := os.RemoveAll(out); err != nil {
		return nil, nil, err
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		return nil, nil, err
	}
	for _, a := range r.Archives {
		names, ok := Platforms(a.OS, a.Arch)
		if !ok {
			skipped = append(skipped, a.OS+"/"+a.Arch)
			continue
		}
		exe := r.Binary
		if a.OS == "windows" {
			exe += ".exe"
		}
		binary, err := read(filepath.Join(root, filepath.FromSlash(a.Path)), exe)
		if err != nil {
			return gems, skipped, err
		}
		for _, platform := range names {
			file := filepath.Join(out, FileName(r.Name, version, platform))
			if err := os.WriteFile(file, pack(r, version, platform, exe, binary), 0o644); err != nil {
				return gems, skipped, err
			}
			gems = append(gems, Gem{File: file})
		}
	}
	return gems, skipped, nil
}

// launcher is exe/<binary>: RubyGems' binstub loads it with Ruby, so it
// runs the binary in place of itself, without a shell.
func launcher(binary, exe string) string {
	return fmt.Sprintf(`#!/usr/bin/env ruby
# frozen_string_literal: true

# Runs the %[1]s binary this platform gem carries.
binary = File.expand_path("../libexec/%[2]s", __dir__)
abort "%[1]s: #{binary} is missing; reinstall the gem" unless File.file?(binary)
exec([binary, binary], *ARGV)
`, binary, exe)
}

// stamp is every entry's time: a gem's bytes depend on its content only.
var stamp = time.Date(1980, 1, 2, 0, 0, 0, 0, time.UTC)

type entry struct {
	name string
	data []byte
	mode int64
}

// pack is a .gem: a tar of metadata.gz, data.tar.gz and checksums.yaml.gz.
func pack(r Release, version, platform, exe string, binary []byte) []byte {
	data := []entry{
		{"exe/" + r.Binary, []byte(launcher(r.Binary, exe)), 0o755},
		{"libexec/" + exe, binary, 0o755},
	}
	spec := Spec{
		Name: r.Name, Version: version, Platform: platform,
		Summary: r.Description, Authors: []string{r.Name}, License: r.License,
		Homepage: r.Homepage, SourceCode: r.Repository,
		Files: []string{data[0].name, data[1].name}, Executables: []string{r.Binary},
	}
	if spec.Summary == "" {
		spec.Summary = r.Binary
	}
	if r.Description != "" {
		spec.Description = r.Description + "\n\nThis gem holds the prebuilt " + r.Binary + " binary for " + platform + " and needs no Ruby code at run time."
	}
	metadata := gzipped([]byte(spec.YAML()))
	payload := gzipped(tarball(data))
	var sums strings.Builder
	sums.WriteString("---\n")
	for _, alg := range []struct {
		name string
		sum  func([]byte) []byte
	}{
		{"SHA256", func(b []byte) []byte { s := sha256.Sum256(b); return s[:] }},
		{"SHA512", func(b []byte) []byte { s := sha512.Sum512(b); return s[:] }},
	} {
		fmt.Fprintf(&sums, "%s:\n  metadata.gz: %s\n  data.tar.gz: %s\n", alg.name, hex.EncodeToString(alg.sum(metadata)), hex.EncodeToString(alg.sum(payload)))
	}
	return tarball([]entry{
		{"metadata.gz", metadata, 0o444},
		{"data.tar.gz", payload, 0o444},
		{"checksums.yaml.gz", gzipped([]byte(sums.String())), 0o444},
	})
}

func tarball(entries []entry) []byte {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	for _, e := range entries {
		// Writing to a bytes.Buffer cannot fail.
		_ = tw.WriteHeader(&tar.Header{
			Name: e.name, Mode: e.mode, Size: int64(len(e.data)), ModTime: stamp,
			Typeflag: tar.TypeReg, Uname: "wheel", Gname: "wheel", Format: tar.FormatUSTAR,
		})
		_, _ = tw.Write(e.data)
	}
	_ = tw.Close()
	return buf.Bytes()
}

func gzipped(data []byte) []byte {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	zw.ModTime = stamp
	_, _ = zw.Write(data)
	_ = zw.Close()
	return buf.Bytes()
}

// read returns one file of a zip archive.
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
