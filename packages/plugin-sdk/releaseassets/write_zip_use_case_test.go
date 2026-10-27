package releaseassets

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestZipHoldsProgramsExecutableAndChecksumsMatchTheArchive(t *testing.T) {
	dir := t.TempDir()
	program, readme := filepath.Join(dir, "tool"), filepath.Join(dir, "README.md")
	_ = os.WriteFile(program, []byte("binary"), 0o755)
	_ = os.WriteFile(readme, []byte("read me"), 0o644)
	archive := filepath.Join(dir, "tool.zip")
	if err := WriteZip(archive, []Entry{{Path: program, Name: "tool", Executable: true}, {Path: readme, Name: "README.md"}}); err != nil {
		t.Fatal(err)
	}
	zr, err := zip.OpenReader(archive)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = zr.Close() }()
	if len(zr.File) != 2 || zr.File[0].Mode()&0o111 == 0 || zr.File[1].Mode()&0o111 != 0 {
		t.Errorf("entries: %+v", zr.File)
	}
	sums := filepath.Join(dir, "checksums.txt")
	if err := WriteChecksums(sums, archive); err != nil {
		t.Fatal(err)
	}
	digest, _ := Digest(archive)
	if got, _ := os.ReadFile(sums); string(got) != digest+"  tool.zip\n" {
		t.Errorf("%q", got)
	}
}
