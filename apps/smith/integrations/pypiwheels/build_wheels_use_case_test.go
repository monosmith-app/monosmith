package pypiwheels

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
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
		Name: "acme", Binary: "acme", Version: "5.0.0", Description: "Does things", License: "MIT", OutDir: "wheels",
		Archives: []Archive{
			{Path: archive(t, root, "acme_windows_amd64.zip", "acme.exe"), OS: "windows", Arch: "amd64"},
			{Path: archive(t, root, "acme_freebsd_amd64.zip", "acme"), OS: "freebsd", Arch: "amd64"},
		},
	}
}

func TestBuildPutsTheBinaryInTheWheelScripts(t *testing.T) {
	root := t.TempDir()
	wheels, skipped, err := Build(root, release(t, root))
	if err != nil {
		t.Fatal(err)
	}
	if len(wheels) != 1 || len(skipped) != 1 || skipped[0] != "freebsd/amd64" {
		t.Fatalf("wheels %v skipped %v", wheels, skipped)
	}
	if got := filepath.Base(wheels[0].File); got != "acme-5.0.0-py3-none-win_amd64.whl" {
		t.Errorf("file = %s", got)
	}
	zr, err := zip.OpenReader(wheels[0].File)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = zr.Close() }()
	files := map[string]string{}
	for _, f := range zr.File {
		// PyPI rejects an entry followed by a data descriptor.
		if f.Flags&0x8 != 0 {
			t.Errorf("%s has a data descriptor", f.Name)
		}
		in, _ := f.Open()
		raw, _ := io.ReadAll(in)
		_ = in.Close()
		files[f.Name] = string(raw)
	}
	script := files["acme-5.0.0.data/scripts/acme.exe"]
	if script != "binary acme.exe" {
		t.Errorf("script = %q", script)
	}
	if !strings.Contains(files["acme-5.0.0.dist-info/WHEEL"], "Root-Is-Purelib: false\nTag: py3-none-win_amd64\n") {
		t.Errorf("WHEEL = %q", files["acme-5.0.0.dist-info/WHEEL"])
	}
	if meta := files["acme-5.0.0.dist-info/METADATA"]; !strings.HasPrefix(meta, "Metadata-Version: 2.1\nName: acme\nVersion: 5.0.0\nSummary: Does things\nLicense: MIT\n") {
		t.Errorf("METADATA = %q", meta)
	}
	sum := sha256.Sum256([]byte(script))
	line := fmt.Sprintf("acme-5.0.0.data/scripts/acme.exe,sha256=%s,%d\n", base64.RawURLEncoding.EncodeToString(sum[:]), len(script))
	record := files["acme-5.0.0.dist-info/RECORD"]
	if !strings.Contains(record, line) || !strings.HasSuffix(record, "acme-5.0.0.dist-info/RECORD,,\n") {
		t.Errorf("RECORD = %q", record)
	}
}
