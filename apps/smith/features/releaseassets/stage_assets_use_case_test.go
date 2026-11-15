package releaseassets

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestStageAssetsCopiesUnderTheUploadNamesWithAChecksumsFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "VideoCraw Studio_0.1.0_x64-setup.exe")
	if err := os.WriteFile(src, []byte("installer"), 0o644); err != nil {
		t.Fatal(err)
	}
	staging := filepath.Join(dir, "staging")
	if err := os.MkdirAll(staging, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(staging, "stale.exe"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := stageAssets(staging, []string{src}, map[string]string{src: "studio-0.1.0-windows-x64.exe"}, "studio-0.1.0-windows-x64.sha256")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || filepath.Base(got[0]) != "studio-0.1.0-windows-x64.exe" || filepath.Base(got[1]) != "studio-0.1.0-windows-x64.sha256" {
		t.Fatalf("staged %v (a stale file must not stay)", got)
	}
	sum := sha256.Sum256([]byte("installer"))
	b, err := os.ReadFile(got[1])
	if want := hex.EncodeToString(sum[:]) + "  studio-0.1.0-windows-x64.exe\n"; err != nil || string(b) != want {
		t.Errorf("checksums %q, want %q (%v)", b, want, err)
	}
}
