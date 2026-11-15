package releasepack

import (
	"os"
	"path/filepath"
	"testing"
)

func TestChecksumWritesASha256sumFileBesideEachPackage(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "out"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "out", "core-0.5.2.tgz"), []byte("abc"), 0o644); err != nil {
		t.Fatal(err)
	}
	files, sidecars, err := Checksum(root, []string{"out/core-0.5.2.tgz"})
	if err != nil {
		t.Fatal(err)
	}
	const want = "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
	if files[0].SHA256 != want || files[0].Name != "core-0.5.2.tgz" || files[0].Size != 3 || sidecars[0] != "out/core-0.5.2.tgz.sha256" {
		t.Fatalf("%+v %v", files, sidecars)
	}
	if raw, _ := os.ReadFile(filepath.Join(root, "out", "core-0.5.2.tgz.sha256")); string(raw) != want+"  core-0.5.2.tgz\n" {
		t.Fatalf("a checksum file sha256sum -c accepts: %q", raw)
	}
}
