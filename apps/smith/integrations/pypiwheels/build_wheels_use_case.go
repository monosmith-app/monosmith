package pypiwheels

import (
	"archive/zip"
	"bytes"
	"compress/flate"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"hash/crc32"
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

// Release is an app's release as PyPI wheels.
type Release struct {
	// Name is the PyPI project.
	Name    string
	Binary  string
	Version string
	// Description, License and Homepage describe the project on PyPI.
	Description, License, Homepage string
	Archives                       []Archive
	// OutDir is the workspace-relative directory the wheels are written to.
	OutDir string
}

// Wheel is one built wheel.
type Wheel struct {
	// File is the wheel's absolute path.
	File string
	// Metadata is its METADATA, which the upload repeats.
	Metadata Metadata
}

// Metadata is what PyPI shows for the project.
type Metadata struct {
	Name, Version, Summary, License, Homepage, Description string
}

// text is the core metadata file (METADATA, version 2.1).
func (m Metadata) text() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Metadata-Version: 2.1\nName: %s\nVersion: %s\n", m.Name, m.Version)
	if m.Summary != "" {
		fmt.Fprintf(&b, "Summary: %s\n", m.Summary)
	}
	if m.Homepage != "" {
		fmt.Fprintf(&b, "Home-page: %s\n", m.Homepage)
	}
	if m.License != "" {
		fmt.Fprintf(&b, "License: %s\n", m.License)
	}
	b.WriteString("Description-Content-Type: text/markdown\n\n" + m.Description)
	return b.String()
}

// Build writes one wheel per archive into OutDir. The binary goes in the
// wheel's scripts, so pip, pipx and uv put it on PATH as a command with no
// Python in between. Archives with no wheel platform are returned as skipped.
func Build(root string, r Release) (wheels []Wheel, skipped []string, err error) {
	version, err := Version(r.Version)
	if err != nil {
		return nil, nil, err
	}
	meta := Metadata{Name: r.Name, Version: version, Summary: r.Description, License: r.License, Homepage: r.Homepage, Description: readme(r)}
	out := filepath.Join(root, filepath.FromSlash(r.OutDir))
	if err := os.RemoveAll(out); err != nil {
		return nil, nil, err
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		return nil, nil, err
	}
	for _, a := range r.Archives {
		tags, ok := Tags(a.OS, a.Arch)
		if !ok {
			skipped = append(skipped, a.OS+"/"+a.Arch)
			continue
		}
		file := filepath.Join(out, FileName(r.Name, version, tags))
		if err := buildOne(filepath.Join(root, filepath.FromSlash(a.Path)), file, r.Binary, a.OS, tags, meta); err != nil {
			return wheels, skipped, err
		}
		wheels = append(wheels, Wheel{File: file, Metadata: meta})
	}
	return wheels, skipped, nil
}

// FileName is a wheel's file name, its tags compressed into one
// (py3-none-a.b.c).
func FileName(name, version string, tags []string) string {
	platforms := make([]string, len(tags))
	for i, t := range tags {
		platforms[i] = strings.TrimPrefix(t, "py3-none-")
	}
	return fmt.Sprintf("%s-%s-py3-none-%s.whl", Distribution(name), version, strings.Join(platforms, "."))
}

func readme(r Release) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n", r.Name)
	if r.Description != "" {
		b.WriteString(r.Description + "\n\n")
	}
	run := r.Binary
	if r.Name != r.Binary {
		// uvx runs the command named like the package unless told otherwise.
		run = "--from " + r.Name + " " + r.Binary
	}
	fmt.Fprintf(&b, "```sh\npipx install %s     # or: uv tool install %s, pip install %s\nuvx %s --version   # run it without installing\n```\n\n", r.Name, r.Name, r.Name, run)
	fmt.Fprintf(&b, "Each wheel holds the prebuilt %s binary for one platform and installs it as a command; it needs no Python at run time.\n", r.Binary)
	if r.Homepage != "" {
		fmt.Fprintf(&b, "\nMore: %s\n", r.Homepage)
	}
	return b.String()
}

// stamp is every entry's time: a wheel's bytes depend on its content only.
var stamp = time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC)

func buildOne(archive, file, binary, goos string, tags []string, meta Metadata) error {
	exe := binary
	if goos == "windows" {
		exe += ".exe"
	}
	content, err := read(archive, exe)
	if err != nil {
		return err
	}
	prefix := Distribution(meta.Name) + "-" + meta.Version
	info := prefix + ".dist-info/"
	var wheel strings.Builder
	wheel.WriteString("Wheel-Version: 1.0\nGenerator: smith\nRoot-Is-Purelib: false\n")
	for _, t := range tags {
		fmt.Fprintf(&wheel, "Tag: %s\n", t)
	}
	entries := []struct {
		name string
		data []byte
		mode os.FileMode
	}{
		{prefix + ".data/scripts/" + exe, content, 0o755},
		{info + "METADATA", []byte(meta.text()), 0o644},
		{info + "WHEEL", []byte(wheel.String()), 0o644},
	}
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	var record strings.Builder
	for _, e := range entries {
		if err := put(zw, e.name, e.data, e.mode); err != nil {
			return err
		}
		sum := sha256.Sum256(e.data)
		fmt.Fprintf(&record, "%s,sha256=%s,%d\n", e.name, base64.RawURLEncoding.EncodeToString(sum[:]), len(e.data))
	}
	fmt.Fprintf(&record, "%sRECORD,,\n", info)
	if err := put(zw, info+"RECORD", []byte(record.String()), 0o644); err != nil {
		return err
	}
	if err := zw.Close(); err != nil {
		return err
	}
	return os.WriteFile(file, buf.Bytes(), 0o644)
}

// put writes one deflated entry with its sizes and checksum in the local
// header. PyPI rejects wheels whose entries end in a data descriptor, which
// zip.Writer.CreateHeader always writes.
func put(zw *zip.Writer, name string, data []byte, mode os.FileMode) error {
	var deflated bytes.Buffer
	fw, err := flate.NewWriter(&deflated, flate.DefaultCompression)
	if err != nil {
		return err
	}
	if _, err := fw.Write(data); err != nil {
		return err
	}
	if err := fw.Close(); err != nil {
		return err
	}
	header := &zip.FileHeader{
		Name: name, Method: zip.Deflate, Modified: stamp,
		CRC32:              crc32.ChecksumIEEE(data),
		CompressedSize64:   uint64(deflated.Len()),
		UncompressedSize64: uint64(len(data)),
	}
	header.SetMode(mode)
	w, err := zw.CreateRaw(header)
	if err != nil {
		return err
	}
	_, err = w.Write(deflated.Bytes())
	return err
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
