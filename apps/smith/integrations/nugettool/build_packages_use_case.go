package nugettool

import (
	"archive/zip"
	"bytes"
	"compress/flate"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Archive is one platform's release archive, holding the binary at its root.
type Archive struct {
	// Path is workspace-relative.
	Path     string
	OS, Arch string
}

// Release is an app's release as a .NET tool.
type Release struct {
	// ID is the NuGet package users install; each platform's binary is in
	// <ID>.<rid>.
	ID      string
	Binary  string
	Version string
	// Authors, Description, License, Homepage and Repository describe the
	// tool on NuGet.
	Authors, Description, License, Homepage, Repository string
	Archives                                            []Archive
	// OutDir is the workspace-relative directory the packages are written to.
	OutDir string
}

// Package is one built .nupkg.
type Package struct {
	ID string
	// File is the package's absolute path.
	File string
}

// Build writes one package per platform, then the tool package that points
// at them, into OutDir: `dotnet tool install` (.NET 10 and later) picks the
// platform's package and runs its binary directly. Archives with no runtime
// identifier are returned as skipped. The tool package comes last, so it is
// pushed only after the packages it points at.
func Build(root string, r Release) (packages []Package, skipped []string, err error) {
	out := filepath.Join(root, filepath.FromSlash(r.OutDir))
	if err := os.RemoveAll(out); err != nil {
		return nil, nil, err
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		return nil, nil, err
	}
	archives := append([]Archive(nil), r.Archives...)
	sort.Slice(archives, func(i, j int) bool { return archives[i].Path < archives[j].Path })
	var ridPackages [][2]string
	for _, a := range archives {
		rid, ok := RID(a.OS, a.Arch)
		if !ok {
			skipped = append(skipped, a.OS+"/"+a.Arch)
			continue
		}
		exe := r.Binary
		if a.OS == "windows" {
			exe += ".exe"
		}
		content, err := read(filepath.Join(root, filepath.FromSlash(a.Path)), exe)
		if err != nil {
			return packages, skipped, err
		}
		meta := r.metadata(r.ID + "." + rid)
		file := filepath.Join(out, FileName(meta.ID, r.Version))
		dir := "tools/any/" + rid + "/"
		err = write(file, meta, "DotnetToolRidPackage", []entry{
			{dir + exe, content, 0o755},
			{dir + "DotnetToolSettings.xml", []byte(ridToolSettings(r.Binary, exe)), 0o644},
		})
		if err != nil {
			return packages, skipped, err
		}
		packages = append(packages, Package{ID: meta.ID, File: file})
		ridPackages = append(ridPackages, [2]string{rid, meta.ID})
	}
	if len(ridPackages) == 0 {
		return packages, skipped, nil
	}
	meta := r.metadata(r.ID)
	meta.Readme = "README.md"
	file := filepath.Join(out, FileName(meta.ID, r.Version))
	err = write(file, meta, "DotnetTool", []entry{
		{"README.md", []byte(readme(r)), 0o644},
		{"tools/any/any/DotnetToolSettings.xml", []byte(toolSettings(r.Binary, ridPackages)), 0o644},
	})
	if err != nil {
		return packages, skipped, err
	}
	return append(packages, Package{ID: meta.ID, File: file}), skipped, nil
}

func (r Release) metadata(id string) Metadata {
	m := Metadata{ID: id, Version: r.Version, Authors: r.Authors, Description: r.Description, License: r.License, ProjectURL: r.Homepage, Repository: r.Repository}
	if m.Authors == "" {
		m.Authors = r.ID
	}
	if m.Description == "" {
		m.Description = "The " + r.Binary + " command"
	}
	return m
}

func readme(r Release) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n", r.ID)
	if r.Description != "" {
		b.WriteString(r.Description + "\n\n")
	}
	fmt.Fprintf(&b, "```sh\ndotnet tool install -g %s\n%s --version\n```\n\n", r.ID, r.Binary)
	fmt.Fprintf(&b, "Each platform's prebuilt %s binary is in its own package (%s.<rid>), which the .NET 10 SDK picks on install; the command needs no .NET runtime.\n", r.Binary, r.ID)
	if r.Homepage != "" {
		fmt.Fprintf(&b, "\nMore: %s\n", r.Homepage)
	}
	return b.String()
}

type entry struct {
	name string
	data []byte
	mode os.FileMode
}

// stamp is every entry's time: a package's bytes depend on its content only.
var stamp = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

// write builds a .nupkg: the files, the manifest, and the Open Packaging
// Conventions parts NuGet reads alongside it.
func write(file string, meta Metadata, packageType string, files []entry) error {
	nuspecFile := strings.ToLower(meta.ID) + ".nuspec"
	properties := "package/services/metadata/core-properties/" + strings.ReplaceAll(strings.ToLower(meta.ID), ".", "") + ".psmdcp"
	names := []string{nuspecFile}
	for _, f := range files {
		names = append(names, f.name)
	}
	all := append([]entry{
		{"_rels/.rels", []byte(relationships(nuspecFile, properties)), 0o644},
		{nuspecFile, []byte(nuspec(meta, packageType)), 0o644},
	}, files...)
	all = append(all,
		entry{"[Content_Types].xml", []byte(contentTypes(names)), 0o644},
		entry{properties, []byte(coreProperties(meta)), 0o644},
	)
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, e := range all {
		if err := put(zw, e.name, e.data, e.mode); err != nil {
			return err
		}
	}
	if err := zw.Close(); err != nil {
		return err
	}
	return os.WriteFile(file, buf.Bytes(), 0o644)
}

// put writes one deflated entry with its sizes and checksum in the local
// header, as NuGet's own packages have them.
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
