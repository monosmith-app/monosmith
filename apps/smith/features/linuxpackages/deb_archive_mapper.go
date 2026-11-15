package linuxpackages

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"fmt"
	"strings"
	"time"
)

// stamp is every entry's time: a package's bytes depend on its content only.
var stamp = time.Unix(0, 0).UTC()

// Deb is the .deb holding binary at /usr/bin/<p.Binary>: an ar archive of
// debian-binary, control.tar.gz and data.tar.gz.
func Deb(p Package, binary []byte) ([]byte, error) {
	target := "usr/bin/" + p.Binary
	data, err := tarGz([]tarEntry{
		{name: "./usr/", dir: true},
		{name: "./usr/bin/", dir: true},
		{name: "./" + target, body: binary, mode: 0o755},
	})
	if err != nil {
		return nil, err
	}
	sum := md5.Sum(binary)
	control, err := tarGz([]tarEntry{
		{name: "./control", body: []byte(debControl(p, len(binary))), mode: 0o644},
		{name: "./md5sums", body: fmt.Appendf(nil, "%x  %s\n", sum, target), mode: 0o644},
	})
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	out.WriteString("!<arch>\n")
	for _, m := range []struct {
		name string
		body []byte
	}{{"debian-binary", []byte("2.0\n")}, {"control.tar.gz", control}, {"data.tar.gz", data}} {
		fmt.Fprintf(&out, "%-16s%-12d%-6d%-6d%-8s%-10d`\n", m.name, stamp.Unix(), 0, 0, "100644", len(m.body))
		out.Write(m.body)
		if len(m.body)%2 == 1 {
			out.WriteByte('\n')
		}
	}
	return out.Bytes(), nil
}

func debControl(p Package, size int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Package: %s\nVersion: %s\nArchitecture: %s\nMaintainer: %s\n", p.Name, packageVersion(p.Version), debArches[p.Arch], p.Maintainer)
	fmt.Fprintf(&b, "Installed-Size: %d\nSection: devel\nPriority: optional\n", (size+1023)/1024)
	if p.Homepage != "" {
		fmt.Fprintf(&b, "Homepage: %s\n", p.Homepage)
	}
	summary := p.Summary
	if summary == "" {
		summary = p.Name
	}
	fmt.Fprintf(&b, "Description: %s\n", summary)
	return b.String()
}

type tarEntry struct {
	name string
	body []byte
	mode int64
	dir  bool
}

func tarGz(entries []tarEntry) ([]byte, error) {
	var buf bytes.Buffer
	gz, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return nil, err
	}
	tw := tar.NewWriter(gz)
	for _, e := range entries {
		h := &tar.Header{Name: e.name, Mode: e.mode, Size: int64(len(e.body)), ModTime: stamp, Uname: "root", Gname: "root", Typeflag: tar.TypeReg, Format: tar.FormatGNU}
		if e.dir {
			h.Typeflag, h.Mode, h.Size = tar.TypeDir, 0o755, 0
		}
		if err := tw.WriteHeader(h); err != nil {
			return nil, err
		}
		if _, err := tw.Write(e.body); err != nil {
			return nil, err
		}
	}
	if err := tw.Close(); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
