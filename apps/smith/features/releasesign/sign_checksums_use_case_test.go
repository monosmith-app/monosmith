package releasesign

import (
	"crypto/rand"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"aead.dev/minisign"
)

func TestSignWritesAVerifiableSignatureBesideTheChecksums(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "dist", "release", "smith")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	sums := []byte("abc  monosmith_1.0.0_linux_amd64.zip\n")
	if err := os.WriteFile(filepath.Join(dir, "checksums.txt"), sums, 0o644); err != nil {
		t.Fatal(err)
	}
	public, private, err := minisign.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	key, _ := private.MarshalText()
	assets := []string{"dist/release/smith/monosmith_1.0.0_linux_amd64.zip", "dist/release/smith/checksums.txt"}
	got, err := Sign(root, assets, string(key), "smith@1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(got, "dist/release/smith/checksums.txt.minisig") || len(got) != 3 {
		t.Fatalf("the signature is attached too: %v", got)
	}
	signature, err := os.ReadFile(filepath.Join(dir, "checksums.txt.minisig"))
	if err != nil {
		t.Fatal(err)
	}
	if !minisign.Verify(public, sums, signature) {
		t.Fatal("the signature does not verify with the public key")
	}
	if minisign.Verify(public, []byte("tampered"), signature) {
		t.Fatal("a changed list must not verify")
	}
}

func TestSignWithoutAKeyChangesNothing(t *testing.T) {
	assets := []string{"dist/release/smith/checksums.txt"}
	got, err := Sign(t.TempDir(), assets, "", "smith@1.0.0")
	if err != nil || !slices.Equal(got, assets) {
		t.Fatalf("got %v, %v", got, err)
	}
	if _, err := Sign(t.TempDir(), assets, "not a key", "smith@1.0.0"); err == nil {
		t.Fatal("a key that cannot be read must say so")
	}
}
