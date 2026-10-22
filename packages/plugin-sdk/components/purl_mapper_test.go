package components

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestPURLEncodesTheSegments(t *testing.T) {
	if got := PURL("npm", "", "@scope/x", "1.0.0+b1"); got != "pkg:npm/%40scope%2Fx@1.0.0%2Bb1" {
		t.Errorf("%s", got)
	}
	if got := PURL("composer", "acme", "lib", ""); got != "pkg:composer/acme/lib" {
		t.Errorf("%s", got)
	}
}

func TestDedupeSorts(t *testing.T) {
	out := Dedupe([]protocol.Component{New("gem", "", "b", "1"), New("gem", "", "a", "1"), New("gem", "", "b", "1")})
	if len(out) != 2 || out[0].Name != "a" {
		t.Errorf("%v", out)
	}
}

func TestFindUpWalksToTheRoot(t *testing.T) {
	root := t.TempDir()
	_ = os.MkdirAll(filepath.Join(root, "apps", "web"), 0o755)
	_ = os.WriteFile(filepath.Join(root, "lock.json"), []byte("x"), 0o644)
	f, ok, err := FindUp(root, "apps/web", "lock.json")
	if err != nil || !ok || f.Rel != "lock.json" {
		t.Fatalf("%v %v %v", f, ok, err)
	}
	if _, ok, _ := FindUp(root, "apps/web", "none.json"); ok {
		t.Error("found a missing file")
	}
	_ = os.WriteFile(filepath.Join(root, "apps", "web", "lock.json"), []byte("y"), 0o644)
	if f, _, _ := FindUp(root, "apps/web", "lock.json"); f.Rel != "apps/web/lock.json" {
		t.Errorf("the project's own file must win: %s", f.Rel)
	}
}
