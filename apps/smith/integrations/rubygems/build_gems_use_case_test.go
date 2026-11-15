package rubygems

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func archive(t *testing.T, root, name, file string) string {
	t.Helper()
	f, err := os.Create(filepath.Join(root, name))
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	w, _ := zw.Create(file)
	_, _ = w.Write([]byte("binary " + file))
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()
	return name
}

func release(t *testing.T, root string) Release {
	return Release{
		Name: "acme", Binary: "acme", Version: "5.0.0-rc.1", Description: "Does things", License: "MIT",
		Homepage: "https://acme.example", Repository: "https://github.com/acme/acme", OutDir: "gems",
		Archives: []Archive{
			{Path: archive(t, root, "acme_windows_amd64.zip", "acme.exe"), OS: "windows", Arch: "amd64"},
			{Path: archive(t, root, "acme_linux_amd64.zip", "acme"), OS: "linux", Arch: "amd64"},
			{Path: archive(t, root, "acme_freebsd_amd64.zip", "acme"), OS: "freebsd", Arch: "amd64"},
		},
	}
}

// untar reads a tar's regular files and their modes.
func untar(t *testing.T, data []byte) (map[string][]byte, map[string]int64) {
	t.Helper()
	files, modes := map[string][]byte{}, map[string]int64{}
	tr := tar.NewReader(bytes.NewReader(data))
	for {
		h, err := tr.Next()
		if err == io.EOF {
			return files, modes
		}
		if err != nil {
			t.Fatal(err)
		}
		files[h.Name], _ = io.ReadAll(tr)
		modes[h.Name] = h.Mode
	}
}

func gunzip(t *testing.T, data []byte) []byte {
	t.Helper()
	zr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	out, _ := io.ReadAll(zr)
	return out
}

func TestBuildMakesAPlatformGemPerPlatform(t *testing.T) {
	root := t.TempDir()
	gems, skipped, err := Build(root, release(t, root))
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, g := range gems {
		names = append(names, filepath.Base(g.File))
	}
	want := "acme-5.0.0.pre.rc.1-x64-mingw-ucrt.gem acme-5.0.0.pre.rc.1-x86_64-linux.gem acme-5.0.0.pre.rc.1-x86_64-linux-musl.gem"
	if strings.Join(names, " ") != want || len(skipped) != 1 || skipped[0] != "freebsd/amd64" {
		t.Fatalf("gems %v skipped %v", names, skipped)
	}

	raw, _ := os.ReadFile(gems[0].File)
	outer, _ := untar(t, raw)
	data, modes := untar(t, gunzip(t, outer["data.tar.gz"]))
	if string(data["libexec/acme.exe"]) != "binary acme.exe" || modes["libexec/acme.exe"] != 0o755 {
		t.Errorf("binary = %q mode %o", data["libexec/acme.exe"], modes["libexec/acme.exe"])
	}
	if l := string(data["exe/acme"]); !strings.HasPrefix(l, "#!/usr/bin/env ruby") || !strings.Contains(l, `"../libexec/acme.exe"`) {
		t.Errorf("launcher = %q", l)
	}
	meta := string(gunzip(t, outer["metadata.gz"]))
	for _, s := range []string{`platform: "x64-mingw-ucrt"`, "  version: \"5.0.0.pre.rc.1\"", "executables:\n- \"acme\"", "files:\n- \"exe/acme\"\n- \"libexec/acme.exe\"", `  source_code_uri: "https://github.com/acme/acme"`, `summary: "Does things"`} {
		if !strings.Contains(meta, s) {
			t.Errorf("metadata lacks %q:\n%s", s, meta)
		}
	}
	sum := sha256.Sum256(outer["data.tar.gz"])
	if sums := string(gunzip(t, outer["checksums.yaml.gz"])); !strings.Contains(sums, "  data.tar.gz: "+hex.EncodeToString(sum[:])+"\n") {
		t.Errorf("checksums = %s", sums)
	}
}

func TestBuildIsReproducible(t *testing.T) {
	root := t.TempDir()
	r := release(t, root)
	first, _, _ := Build(root, r)
	a, _ := os.ReadFile(first[1].File)
	again, _, _ := Build(root, r)
	b, _ := os.ReadFile(again[1].File)
	if !bytes.Equal(a, b) {
		t.Error("two builds differ")
	}
}
