package linuxpackages

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"strconv"
	"strings"
	"testing"
)

// arMembers reads an ar archive as member name → content.
func arMembers(t *testing.T, raw []byte) ([]string, map[string][]byte) {
	t.Helper()
	if !bytes.HasPrefix(raw, []byte("!<arch>\n")) {
		t.Fatalf("not an ar archive: %q", raw[:8])
	}
	var names []string
	members := map[string][]byte{}
	for rest := raw[8:]; len(rest) > 0; {
		header := rest[:60]
		if string(header[58:60]) != "`\n" {
			t.Fatalf("bad ar header %q", header)
		}
		name := strings.TrimSpace(string(header[:16]))
		size, err := strconv.Atoi(strings.TrimSpace(string(header[48:58])))
		if err != nil {
			t.Fatal(err)
		}
		names = append(names, name)
		members[name] = rest[60 : 60+size]
		rest = rest[60+size+size%2:]
	}
	return names, members
}

func tarFiles(t *testing.T, raw []byte) map[string]*tar.Header {
	t.Helper()
	gz, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	tr := tar.NewReader(gz)
	files := map[string]*tar.Header{}
	for {
		h, err := tr.Next()
		if err == io.EOF {
			return files
		}
		if err != nil {
			t.Fatal(err)
		}
		files[h.Name] = h
	}
}

func tarFile(t *testing.T, raw []byte, name string) string {
	t.Helper()
	gz, _ := gzip.NewReader(bytes.NewReader(raw))
	tr := tar.NewReader(gz)
	for {
		h, err := tr.Next()
		if err != nil {
			t.Fatalf("%s not found: %v", name, err)
		}
		if h.Name == name {
			b, _ := io.ReadAll(tr)
			return string(b)
		}
	}
}

func TestDebHoldsTheBinaryInUsrBin(t *testing.T) {
	p := Package{Name: "monosmith", Version: "1.2.0-rc.1", Arch: "arm64", Summary: "Runs a monorepo", Homepage: "https://example.test", License: "MIT", Maintainer: "monosmith-app <https://example.test>", Binary: "smith"}
	raw, err := Deb(p, []byte("ELF binary"))
	if err != nil {
		t.Fatal(err)
	}
	names, members := arMembers(t, raw)
	if strings.Join(names, ",") != "debian-binary,control.tar.gz,data.tar.gz" {
		t.Fatalf("members = %v", names)
	}
	if string(members["debian-binary"]) != "2.0\n" {
		t.Errorf("debian-binary = %q", members["debian-binary"])
	}
	control := tarFile(t, members["control.tar.gz"], "./control")
	for _, line := range []string{"Package: monosmith\n", "Version: 1.2.0~rc.1\n", "Architecture: arm64\n", "Maintainer: monosmith-app <https://example.test>\n", "Homepage: https://example.test\n", "Description: Runs a monorepo\n"} {
		if !strings.Contains(control, line) {
			t.Errorf("control lacks %q:\n%s", line, control)
		}
	}
	if sums := tarFile(t, members["control.tar.gz"], "./md5sums"); !strings.HasSuffix(sums, "  usr/bin/smith\n") {
		t.Errorf("md5sums = %q", sums)
	}
	data := tarFiles(t, members["data.tar.gz"])
	bin, ok := data["./usr/bin/smith"]
	if !ok || bin.Mode != 0o755 || bin.Uname != "root" {
		t.Fatalf("data = %v", data)
	}
	if got := tarFile(t, members["data.tar.gz"], "./usr/bin/smith"); got != "ELF binary" {
		t.Errorf("binary = %q", got)
	}
}

func TestDebIsReproducible(t *testing.T) {
	p := Package{Name: "x", Version: "1.0.0", Arch: "amd64", Binary: "x"}
	a, _ := Deb(p, []byte("same"))
	b, _ := Deb(p, []byte("same"))
	if !bytes.Equal(a, b) {
		t.Error("two builds of one package differ")
	}
}
